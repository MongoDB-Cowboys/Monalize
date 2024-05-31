package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/MongoDB-Cowboys/Monalize/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bsonDocument bson.D
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

var currentdb string   // *Current database name* create for disable duplicate in excell
var currentcoll string // *Current collection name*  create for disable duplicate in excell
var currentcnt string  // *Current Count docs in collection* create for disable duplicate in excell
var data = [][]string{}

func ProcessSingleDB(client *mongo.Client, ctx context.Context, dbName string, boolExcel *bool) [][]string {
	col, err := client.Database(dbName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(utils.Database("- Database: ", dbName))
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
		if err != nil {
			fmt.Printf("Error documents counting: %v\n", err)
			continue
		}
		cnt := int(count)

		str_cnt := strconv.Itoa(cnt) // convert int to str
		fmt.Println(utils.Collections("--- Collection: ", z, " Count: ", cnt))

		for cur.Next(context.Background()) {
			if z == currentcoll { // create for disable duplicate in excell
				z = " "
			}

			currentcoll = z // create for disable duplicate in excell

			index := (&bsonDocument)
			err := cur.Decode(&index)
			if err != nil {
				fmt.Printf("Error decoding document: %v\n", err)
				continue
			}
			var jsonDocument map[string]interface{}
			temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, true, true)
			if err != nil {
				fmt.Printf("Error marshaling document: %v\n", err)
				continue
			}
			err = json.Unmarshal(temporaryBytes, &jsonDocument)
			var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})
			args, _ := json.Marshal(jsonKey) // marshal map[string]interface{} to str
			fmt.Println(utils.Index(utils.JsonToStr(string(args))))
			if *boolExcel {

				if str_cnt == currentcnt { // create for disable duplicate in excell
					str_cnt = " "
				}
				currentcnt = str_cnt  // create for disable duplicate in excell
				if dbs == currentdb { // create for disable duplicate in excell
					dbs = " "
				}
				currentdb = dbs // create for disable duplicate in excell

				logStr := utils.JsonToStr(string(args))
				data = append(data, []string{dbs, currentcoll, str_cnt, logStr}) // append to csv
			}
			_ = err

		}
		_ = x
	}
	return data
}

func ProcessAllDBs(client *mongo.Client, ctx context.Context, boolExcel *bool) [][]string {
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range databases {

		col, err := client.Database(s).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(utils.Database(i, "- Database: ", s))
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
			if err != nil {
				fmt.Printf("Error documents counting: %v\n", err)
				continue
			}
			cnt := int(count)

			str_cnt := strconv.Itoa(cnt) // convert int to str

			fmt.Println(utils.Collections("--- Collection: ", z, " Count: ", cnt))

			for cur.Next(context.Background()) {
				index := (&bsonDocument)
				err := cur.Decode(&index)
				if err != nil {
					fmt.Printf("Error decoding document: %v\n", err)
					continue
				}
				var jsonDocument map[string]interface{}
				temporaryBytes, err := bson.MarshalExtJSON(bsonDocument, true, true)
				if err != nil {
					fmt.Printf("Error marshaling document: %v\n", err)
					continue
				}
				err = json.Unmarshal(temporaryBytes, &jsonDocument)
				var jsonKey map[string]interface{} = jsonDocument["key"].(map[string]interface{})

				args, _ := json.Marshal(jsonKey) // marshal map[string]interface{} to str
				fmt.Println(utils.Index(utils.JsonToStr(string(args))))

				if z == currentcoll { // create for disable duplicate in excell
					z = " "
				}
				currentcoll = z // create for disable duplicate in excell
				if *boolExcel {

					if str_cnt == currentcnt { // create for disable duplicate in excell
						str_cnt = " "
					}
					currentcnt = str_cnt  // create for disable duplicate in excell
					if dbs == currentdb { // create for disable duplicate in excell
						dbs = " "
					}
					currentdb = dbs // create for disable duplicate in excell

					logStr := utils.JsonToStr(string(args))
					data = append(data, []string{dbs, currentcoll, str_cnt, logStr}) // append to csv

				}
				_ = err
			}
			_ = x
		}
	}
	return data
}

func GetCollections(database *mongo.Database, name string) (*mongo.Cursor, error) {
	filter := bson.D{}
	if len(name) > 0 {
		filter = append(filter, primitive.E{Key: "name", Value: name})
	}

	ctx := context.TODO()
	cursor, err := database.ListCollections(ctx, filter)
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
	ctx := context.TODO()

	for iter.Next(ctx) {
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

func PrintCurrentInfo(client *mongo.Client) {
	fmt.Println(utils.Info("Search slow query..."))

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
		fmt.Println(utils.Current(string(jsonResult)))
	} else {
		fmt.Println("--- stderr ---")
		fmt.Println(err)
	}
}
