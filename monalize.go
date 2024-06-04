package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mongodb "github.com/MongoDB-Cowboys/Monalize/mongo"
	"github.com/MongoDB-Cowboys/Monalize/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func CloseHandler() {
	c := make(chan os.Signal, 1) // Buffered channel with capacity 1
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

var data = [][]string{}

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
	boolPodman := flag.Bool("podman", false, "Add this flag if you are using podman with custom logfile in container")
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
		data = mongodb.ProcessSingleDB(client, ctx, dbName, boolExcel)
	} else {
		data = mongodb.ProcessAllDBs(client, ctx, boolExcel)
	}
	if *boolExcel {
		if err := utils.ToCsvExport(data); err != nil { // this code return data to csv
			log.Fatal(err)
		}
	}
	mongodb.PrintCurrentInfo(client)
	utils.MonitorLogs(logPath, containerName, boolPodman)
	fmt.Println(utils.Index("Done"))
}
