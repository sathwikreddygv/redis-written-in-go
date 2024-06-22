# redis-written-in-go
![godisDB](images/godis.jpeg)

godisDB is a Redis-like in-memory key-value Database which is multithreaded unlike Redis.

## Table of Contents
* [Commands](#commands)
* [Setup](#setup)
* [Usage](#usage)

### Commands
godisDB supports the following commands as of now:
```
PING, SET, GET, SETNX, MSET, MGET, INCR, DECR, RPUSH, LPUSH, RPOP, LPOP, LRANGE, DEL, EXPIRE, TTL
```

### Setup
To set up godisDB:
1. Clone the repository:  
```
  git clone https://github.com/sathwikreddygv/redis-written-in-go.git
```
2. Run the server
```
  go run main.go resp_parser.go resp_writer.go execute_commands.go
```

### Usage
