# Monalize tool overview 

Monaliza is a tool for scanning and analyzing MongoDB database for any performance issues, which lead to high CPU consumption. 

The main task is a fast output of names of all databases and collections, indexes and slow queries stats.

## Compilation

`go build monalize.go`

## Usage 

Compile the tool from src code or install via 
```
MONALIZE=$( curl --silent "https://api.github.com/repos/MongoDB-Cowboys/Monalize/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' ) \
;wget https://github.com/MongoDB-Cowboys/Monalize/releases/download/$MONALIZE/monalize
```
Don't forget to make it executable:
```
chmod +x monalize
```
Then you can run on any Unix like system via `./monalize`

Available flags:

* --db_name (optional) If you need to scan only specific database. (default: nil)
* --db_uri (optional) Uri to connect to mongodb service. (default: "mongodb://localhost:27017")
* --excel (optional) To save an output of the script to excel file. (default: false)
* --logpath (optional) Specify a path to MongoDB service log file. (default "/var/log/mongodb/mongodb.log")

A help is available via `./monalize --help`.

Examples uri: 

* `mongodb://User:Pwd@ip.ip.ip.ip:port`

Full request example:

* `monalize --db_uri "mongodb://user:passwd@127.0.0.1:27017/?&authSource=admin"`

After successfull run of the `monalize` tool, all the output artifacts will be saved in working directory:

* `colout.txt` COLLSCAN logs.
* `result.csv` optional excel file.

## License 

GPL-3.0 License
