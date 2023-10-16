package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var jsonDocuments []map[string]interface{}
var byteDocuments []byte
var bsonDocument bson.D
var jsonDocument map[string]interface{}
var temporaryBytes []byte

func (ci *CollectionInfo) IsView() bool {
	return ci.Type == "view"
}

type CollectionInfo struct {
	Name    string `bson:"name"`
	Type    string `bson:"type"`
	Options bson.M `bson:"options"`
	Info    bson.M `bson:"info"`
}

var (
	Database    = Teal
	Collections = Yellow
	Index       = Red
	Current     = Magenta
	COLLSCAN    = Green
	Info        = Purple
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func writeToFile(filename, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func filterPrintableASCII(data []byte) string {
	var filteredData []byte
	inLine := false

	for _, b := range data {
		if b == byte('\n') {
			inLine = false
		}
		// Check if the byte is within the range of printable ASCII characters
		if b >= 32 && b < 127 {
			filteredData = append(filteredData, b)
			inLine = true
		} else if !inLine && b == byte('\n') {
			filteredData = append(filteredData, b)
		}
	}

	return string(filteredData)
}

func monitorLogs(logPath, containerName string) {
	fmt.Println(Info("Monitoring logs mongodb..."))

	var file *os.File
	var err error
	targetPath := "mongo_logs.txt"
	if containerName != "" && logPath != "" {
		fmt.Println(Info("Detected docker container usage with custom path to log file."))

		cmd := exec.Command("docker", "exec", containerName, "cat", logPath)

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}

		file, err = os.Create(targetPath) // Set the file variable
		if err != nil {
			log.Fatalf("failed to create file: %s", err)
		}
		defer file.Close()

		_, err = file.WriteString(string(output))
		if err != nil {
			log.Fatalf("error writing to file: %s", err)
		}
	} else if containerName != "" && logPath == "" {
		fmt.Println(Info("Detected docker container usage with default stream logging."))

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatalf("Failed to create Docker client: %s", err)
		}

		options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Timestamps: false}
		out, err := cli.ContainerLogs(context.Background(), containerName, options)
		if err != nil {
			log.Fatalf("Failed to get container logs: %s", err)
		}
		defer out.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, out); err != nil {
			log.Fatalf("Failed to read container logs: %s", err)
		}

		filteredData := filterPrintableASCII(buf.Bytes())
		file, err = os.Create(targetPath) // Set the file variable
		if err != nil {
			log.Fatalf("failed to create file: %s", err)
		}
		defer file.Close()

		_, err = file.WriteString(string(filteredData))
		if err != nil {
			log.Fatalf("error writing to file: %s", err)
		}

	} else {
		fmt.Println(Info("Detected default way with local mongodb log file."))
		content, err := os.ReadFile(logPath)
		if err != nil {
			log.Fatalf("failed to read file: %s", err)
		}

		err = os.WriteFile(targetPath, content, 0644)
		if err != nil {
			log.Fatalf("failed to write to file: %s", err)
		}
	}

	var collScanLines []string

	content, err := os.ReadFile(targetPath) // Use the opened file for reading
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	logs := string(content)

	// Split the logs by newlines
	logLines := strings.Split(logs, "\n")

	for _, line := range logLines {
		if strings.Contains(line, "COLLSCAN") {
			collScanLines = append(collScanLines, line)
		}
	}

	outputFilePath := "colout.txt"
	fmt.Println(Info("Output file path: ", outputFilePath))

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("failed to create output file: %s", err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range collScanLines {
		fmt.Fprintln(writer, line)
	}
	if err := writer.Flush(); err != nil {
		log.Fatalf("error while writing to file: %s", err)
	}
	err = os.Remove(targetPath)
	if err != nil {
		log.Fatalf("failed to remove file: %s", err)
	}
	fmt.Println(Info("--- Collscan Lines Written to ", outputFilePath, "---"))
}

func printCurrentInfo(client *mongo.Client) {
	fmt.Println(Info("Search slow query..."))

	result := bson.M{}
	err := client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "currentOp", Value: 1}, {Key: "secs_running", Value: bson.D{{Key: "$gte", Value: 1}}}}).Decode(&result)
	if err != nil {
		log.Printf("error while fetching slow queries: %v\n", err)
	}
	// Convert the result to a JSON string
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("error while marshalling result to JSON: %v\n", err)
	}
	if err == nil {
		fmt.Println("--- stdout ---")
		fmt.Println(Current(string(jsonResult)))
	} else {
		fmt.Println("--- stderr ---")
		fmt.Println(err)
	}
}

func arrgsToString(strArray []string) string { // func to convert to string
	return strings.Join(strArray, " ")
}

func tocsvExport(data [][]string) error { // func for export data to csv file
	file, err := os.Create("result.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, value := range data {

		if err := writer.Write(value); err != nil {
			return err
		}
	}

	return nil
}
func CloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		DeleteFiles()
		os.Exit(0)
	}()
}

// DeleteFiles is used to simulate a 'clean up' function to run on shutdown.

func DeleteFiles() {
	fmt.Println("- Run Clean Up - Delete Our Files")
	_ = os.Remove("result.csv")
	_ = os.Remove("colout.txt")
	fmt.Println("- Good bye!")
}

// function that edits and returns readable indexes
func jsonToStr(args string) string {
	args = strings.Replace(args, `{"$numberInt":`, "", -1)
	args = strings.Replace(args, `{"$numberDouble":`, "", -1)
	re := regexp.MustCompile(`"`)
	args = re.ReplaceAllString(args, "")
	args = strings.TrimSuffix(args, "}")

	return args
}

func GetCollections(database *mongo.Database, name string) (*mongo.Cursor, error) {
	filter := bson.D{}
	if len(name) > 0 {
		filter = append(filter, primitive.E{Key: "name", Value: name})
	}

	cursor, err := database.ListCollections(nil, filter)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func GetCollectionInfo(coll *mongo.Collection) (*CollectionInfo, error) {
	iter, err := GetCollections(coll.Database(), coll.Name())
	if err != nil {
		return nil, err
	}
	defer iter.Close(context.Background())
	comparisonName := coll.Name()

	var foundCollInfo *CollectionInfo
	for iter.Next(nil) {
		collInfo := &CollectionInfo{}
		err = iter.Decode(collInfo)
		if err != nil {
			return nil, err
		}
		if collInfo.Name == comparisonName {
			foundCollInfo = collInfo
			break
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return foundCollInfo, nil
}

var currentdb string   // *Current database name* create for disable duplicate in excell
var currentcoll string // *Current collection name*  create for disable duplicate in excell
var currentcnt string  // *Current Count docs in collection* create for disable duplicate in excell
var data = [][]string{}

func processSingleDB(client *mongo.Client, ctx context.Context, dbName string, boolExcel *bool) {
	col, err := client.Database(dbName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(Database("- Database: ", dbName))
	var dbs string = dbName // create for disable duplicate in excell

	for x, z := range col {
		c := client.Database(dbName).Collection(z)
		duration := 10 * time.Second
		batchSize := int32(10)
		cur, err := c.Indexes().List(context.Background(), &options.ListIndexesOptions{BatchSize: &batchSize, MaxTime: &duration})
		if err != nil {
			isView := true
			// failure to get CollectionInfo should not cause the function to exit. We only use this to
			// determine if a collection is a view.
			collInfo, err := GetCollectionInfo(c)
			if collInfo != nil {
				isView = collInfo.IsView()
				_ = isView
				continue
			} else {
				log.Fatalf("Something went wrong listing %v", err)
			}

		}
		count, err := client.Database(dbName).Collection(z).CountDocuments(context.Background(), bson.D{})
		cnt := int(count)

		str_cnt := strconv.Itoa(cnt) // convert int to str
		fmt.Println(Collections("--- Collection: ", z, " Count: ", cnt))

		for cur.Next(context.Background()) {
			if z == currentcoll { // create for disable duplicate in excell
				z = " "
			}

			currentcoll = z // create for disable duplicate in excell

			index := (&bsonDocument)
			err := cur.Decode(&index)
			var jsonDocument map[string]interface{}
			temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, true, true)
			err = json.Unmarshal(temporaryBytes, &jsonDocument)
			var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})
			args, _ := json.Marshal(jsonKey) // marshal map[string]interface{} to str
			fmt.Println(Index(jsonToStr(string(args))))
			if *boolExcel == true {

				if str_cnt == currentcnt { // create for disable duplicate in excell
					str_cnt = " "
				}
				currentcnt = str_cnt  // create for disable duplicate in excell
				if dbs == currentdb { // create for disable duplicate in excell
					dbs = " "
				}
				currentdb = dbs // create for disable duplicate in excell

				logStr := jsonToStr(string(args))
				data = append(data, []string{dbs, currentcoll, str_cnt, logStr}) // append to csv
			}
			_ = err

		}
		_ = x
	}
}

func processAllDBs(client *mongo.Client, ctx context.Context, boolExcel *bool) {
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range databases {

		col, err := client.Database(s).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(Database(i, "- Database: ", s))
		var dbs string = s // create for disable duplicate in excell

		for x, z := range col {

			c := client.Database(s).Collection(z)
			duration := 10 * time.Second
			batchSize := int32(10)

			cur, err := c.Indexes().List(context.Background(), &options.ListIndexesOptions{BatchSize: &batchSize, MaxTime: &duration})

			if err != nil {
				isView := true
				// failure to get CollectionInfo should not cause the function to exit. We only use this to
				// determine if a collection is a view.
				collInfo, err := GetCollectionInfo(c)
				if collInfo != nil {
					isView = collInfo.IsView()
					_ = isView
					continue
				} else {
					log.Fatalf("Something went wrong listing %v", err)
				}

			}

			count, err := client.Database(s).Collection(z).CountDocuments(context.Background(), bson.D{})
			cnt := int(count)

			str_cnt := strconv.Itoa(cnt) // convert int to str

			fmt.Println(Collections("--- Collection: ", z, " Count: ", cnt))

			for cur.Next(context.Background()) {
				index := (&bsonDocument)
				err := cur.Decode(&index)
				var jsonDocument map[string]interface{}
				temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, true, true)
				err = json.Unmarshal(temporaryBytes, &jsonDocument)
				var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})

				args, _ := json.Marshal(jsonKey) // marshal map[string]interface{} to str
				fmt.Println(Index(jsonToStr(string(args))))

				if z == currentcoll { // create for disable duplicate in excell
					z = " "
				}
				currentcoll = z // create for disable duplicate in excell
				if *boolExcel == true {

					if str_cnt == currentcnt { // create for disable duplicate in excell
						str_cnt = " "
					}
					currentcnt = str_cnt  // create for disable duplicate in excell
					if dbs == currentdb { // create for disable duplicate in excell
						dbs = " "
					}
					currentdb = dbs // create for disable duplicate in excell

					logStr := jsonToStr(string(args))
					data = append(data, []string{dbs, currentcoll, str_cnt, logStr}) // append to csv

				}
				_ = err
			}
			_ = x
		}
	}
}

func main() {

	CloseHandler()
	var dbURI, dbName, logPath, containerName string
	var contextTimeout int

	flag.StringVar(&dbURI, "db_uri", "mongodb://localhost:27017", "Set custom url to connect to mongodb")
	flag.StringVar(&dbName, "db_name", "", "Set target database, if nil then choose all databases")
	flag.StringVar(&logPath, "logpath", "", "Set path to log file")
	flag.StringVar(&containerName, "container", "", "Set name of Docker container")
	flag.IntVar(&contextTimeout, "context_timeout", 10, "Set context timeout")

	boolExcel := flag.Bool("excel", false, "Add this flag if you want to put the results in an Excel file")
	flag.Parse()
	clientOptions := options.Client().ApplyURI(dbURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Connect to MongoDB is impossible. Check if it works and the entered data.")
	}

	if dbName != "" {
		processSingleDB(client, ctx, dbName, boolExcel)
	} else {
		processAllDBs(client, ctx, boolExcel)
	}
	if *boolExcel == true {
		if err := tocsvExport(data); err != nil { // this code return data to csv
			log.Fatal(err)
		}
	}
	printCurrentInfo(client)
	monitorLogs(logPath, containerName)
	fmt.Println(Index("Done"))
}
