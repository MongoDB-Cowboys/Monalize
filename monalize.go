package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var jsonDocuments []map[string]interface{}
var byteDocuments []byte
var bsonDocument bson.D
var jsonDocument map[string]interface{}
var temporaryBytes []byte

const ShellToUse = "bash"

func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
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

func currentlog(db_uri, logpath string) { // func make log files with slow queries and send files to ftp server. Clear history.
	fmt.Println(Info("Search slow query..."))
	err, out, errout := Shellout("mongo " + db_uri + " --eval " + `'db.currentOp({"secs_running": {$gte: 1}})'`)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	fmt.Println("--- stdout ---")
	fmt.Println(Current(out))
	fmt.Println("--- stderr ---")
	fmt.Println(errout)
	fmt.Println(Info("Monitoring logs mongodb..."))
	err, output, errout := Shellout("cat " + logpath + " | grep COLLSCAN > colout.txt")
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	fmt.Println("--- stdout ---")
	fmt.Println(COLLSCAN(output))
	fmt.Println("--- stderr ---")
	fmt.Println(errout)
	err, history, errout := Shellout("history -c")
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	fmt.Println(Index("History cleaned"))
	fmt.Println(Index("Done"))
	_ = history

}

func typeofobject(x interface{}) { //func to display type object
	fmt.Sprintf("%T", x)
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

// DeleteFiles is used to simulate a 'clean up' function to run on shutdown. Because
// it's just an example it doesn't have any error handling.
func DeleteFiles() {
	fmt.Println("- Run Clean Up - Delete Our Files")
	_ = os.Remove("result.csv")
	_ = os.Remove("colout.txt")
	fmt.Println("- Good bye!")
}

func main() {
	data := [][]string{}
	CloseHandler()
	var db_uri string
	var db_name string
	var logpath string
	flag.StringVar(&db_uri, "db_uri", "mongodb://localhost:27017", "Set custom url to connect to mongodb")
	flag.StringVar(&db_name, "db_name", "", "Set target database, if nil then choose all databases")
	flag.StringVar(&logpath, "logpath", "/var/log/mongodb/mongodb.log", "Set path to log file")
	boolPtr := flag.Bool("excel", false, "Add this flag if you want to put the results in an Excel file")
	flag.Parse()
	client, err := mongo.NewClient(options.Client().ApplyURI(db_uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Connect to MongoDB is impossible. Check if it works and the entered data.")
	}
	var currentdb string   // *Current database name* create for disable duplicate in excell
	var currentcoll string // *Current collection name*  create for disable duplicate in excell
	var currentcnt string  // *Current Count docs in collection* create for disable duplicate in excell
	if db_name != "" {
		col, err := client.Database(db_name).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(Database("- Database: ", db_name))
		var dbs string = db_name // create for disable duplicate in excell

		for x, z := range col {
			c := client.Database(db_name).Collection(z)
			duration := 10 * time.Second
			batchSize := int32(10)
			cur, err := c.Indexes().List(context.Background(), &options.ListIndexesOptions{&batchSize, &duration})
			if err != nil {
				log.Fatalf("Something went wrong listing %v", err)
			}
			count, err := client.Database(db_name).Collection(z).CountDocuments(context.Background(), bson.D{})
			cnt := int(count)

			str_cnt := strconv.Itoa(cnt)
			fmt.Println(Collections("--- Collection: ", z, " Count: ", cnt))

			for cur.Next(context.Background()) {
				inbsonD := bson.D{}
				cur.Decode(&inbsonD)
				fmt.Println(Index(inbsonD[1]))

				if z == currentcoll {
					z = " "
				}

				currentcoll = z

				index := (&bsonDocument)
				err := cur.Decode(&index)
				var jsonDocument map[string]interface{}
				temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, true, true)
				err = json.Unmarshal(temporaryBytes, &jsonDocument)
				var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})
				if *boolPtr == true {

					for k, v := range jsonKey {
						value, ok := v.(map[string]interface{})

						if str_cnt == currentcnt {
							str_cnt = " "
						}

						currentcnt = str_cnt
						if dbs == currentdb {
							dbs = " "
						}
						currentdb = dbs
						if ok {
							for _, val := range value {
								var logStr string = "" + k + ":" + val.(string)
								data = append(data, []string{dbs, currentcoll, str_cnt, logStr})

							}
						} else {
							var logStr string = "" + k + ":" + v.(string)
							data = append(data, []string{dbs, currentcoll, str_cnt, logStr})

						}

					}
				}
				_ = err

			}
			_ = x
		}
	} else {

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

				cur, err := c.Indexes().List(context.Background(), &options.ListIndexesOptions{&batchSize, &duration})
				if err != nil {
					log.Fatalf("Something went wrong listing %v", err)
				}

				count, err := client.Database(s).Collection(z).CountDocuments(context.Background(), bson.D{})
				cnt := int(count)

				str_cnt := strconv.Itoa(cnt)

				fmt.Println(Collections("--- Collection: ", z, " Count: ", cnt))

				for cur.Next(context.Background()) {
					inbsonD := bson.D{}
					cur.Decode(&inbsonD)
					fmt.Println(Index(inbsonD[1]))
					index := (&bsonDocument)
					err := cur.Decode(&index)
					var jsonDocument map[string]interface{}
					temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, true, true)
					err = json.Unmarshal(temporaryBytes, &jsonDocument)
					var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})

					if z == currentcoll {
						z = " "
					}

					currentcoll = z
					if *boolPtr == true {
						for k, v := range jsonKey {
							value, ok := v.(map[string]interface{})

							if str_cnt == currentcnt {
								str_cnt = " "
							}

							currentcnt = str_cnt
							if dbs == currentdb {
								dbs = " "
							}
							currentdb = dbs
							if ok {
								for _, val := range value {
									var logStr string = "" + k + ":" + val.(string)
									data = append(data, []string{dbs, currentcoll, str_cnt, logStr})

								}
							} else {
								var logStr string = "" + k + ":" + v.(string)
								data = append(data, []string{dbs, currentcoll, str_cnt, logStr})

							}

						}
					}
					_ = err
				}
				_ = x
			}
		}
	}
	if *boolPtr == true {
		if err := tocsvExport(data); err != nil { // this code return data to csv
			log.Fatal(err)
		}
	}
	currentlog(db_uri, logpath)
}
