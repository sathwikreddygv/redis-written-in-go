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
  go run main.go resp_parser.go resp_writer.go execute_commands.go persistence.go
```

### Usage
1. redis-cli can be used as a client as godisDB also uses the Redis Serialization protocol (RESP).
2. godisDB runs on port 6369 (Redis runs on 6379)
3. Use this command to connect to godisDB
```
redis-cli -p 6369
```
4. Now you are all set to talk to the godisDB server. Try "PING" command to start with and you should receive a "PONG" as response!
