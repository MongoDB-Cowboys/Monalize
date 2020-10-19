# Monalize tool overview 

Monaliza is a tool for scanning and analyzing MongoDB database for any performance issues, which lead to high CPU consumption. 

The main task is a fast output of names of all databases and collections, indexes and slow queries stats.

## Compilation

`go build monalize.go`

## Usage 

Compile the tool from src code or install via `wget https://github.com/MongoDB-Cowboys/Monalize/releases/latest`

Then you can run on any Unix like system via `./monalize`

Available flags:

* -db_name (optional) If you need to scan only specific database. (default: nil)
* -db_uri (optional) Uri to connect to mongodb service. (default: "mongodb://localhost:27017")
* -excel (optional) To save an output of the script to excel file. (default: false)
* -logpath (optional) Specify a path to MongoDB service log file. (default "/var/log/mongodb/mongodb.log")

A help is available via `./monalize -h`.

Examples uri: 

* `mongodb://User:Pwd@ip.ip.ip.ip:port`

After successfull run of the `monalize` tool, all the output artifacts will be saved in working directory:

* `colout.txt` COLLSCAN logs.
* `result.csv` optional excel file.

## License 

GPL-3.0 License
