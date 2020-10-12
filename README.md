# Monalize v1.1.0

Monaliza is an application for scanning and analyzing Mongodb databases. The main task is fast output Names of all databases and collections, indexes and slow queries.
The usage is very simple, compile the binary if it is not compiled.
`go build monalize.go`

And run on any Unix like system ./monalize

Flags:

* -db_name (optional) If you need scan only one database. (default: nil)
* -db_uri (optional) Uri to connect to mongodb. (default: "mongodb://localhost:27017")
* -excel (optional) Make output to excel file. (default: false)
* -logpath (optional) Set path to log file. (default "/var/log/mongodb/mongodb.log")

Get little help `./monalize -h`.

Examples uri: 

* `mongodb://User:Pwd@ip.ip.ip.ip:port`

After successfully use all data files from `monalize` saves in working directory.

This is:
* `colout.txt` COLLSCAN logs.
* `result.csv` optional excel file.
