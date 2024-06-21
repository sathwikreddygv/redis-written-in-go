package main

import (
	// "bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DataType int

const (
	StringType DataType = iota
	ListType
	HashType
	SetType
	SortedSetType
)

type sortedSetMember struct {
	Member string
	score  float64
}

type KeyValueStore struct {
	Strings     map[string]string
	Lists       map[string][]string
	Hashes      map[string]map[string]string
	Sets        map[string]map[string]struct{}
	SortedSets  map[string][]sortedSetMember
	Expirations map[string]time.Time
	mu          sync.RWMutex
	// CurrentTx   *Transaction
	totalCommandsProcessed int
	// connectedClients map[string]net.Conn
}

func executeSET(args []string, kv *KeyValueStore) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'SET' command"
	} else {
		key := args[0]
		value := args[1]
		kv.Strings[key] = value
	}
	return "OK"
}

func executeSETNX(args []string, kv *KeyValueStore) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'SETNX' command"
	} else {
		key := args[0]
		value := args[1]
		if _, exists := kv.Strings[key]; !exists {
			kv.Strings[key] = value
		}
	}
	return "OK"
}

func executeMSET(args []string, kv *KeyValueStore) string {
	if len(args)%2 != 0 {
		return "(error) ERR wrong number of arguments for 'MSET' command"
	} else {
		for i := 0; i < len(args); i += 2 {
			key := args[i]
			value := args[i+1]
			kv.Strings[key] = value
		}
	}
	return "OK"
}

func executeMGET(args []string, kv *KeyValueStore) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'MGET' command"
	}

	var final []string
	for i := 0; i < len(args); i++ {
		key := args[i]
		if value, exists := kv.Strings[key]; exists {
			final = append(final, value)
		} else {
			final = append(final, "(nil)")
		}
	}

	return strings.Join(final, " ")
}

func executeGET(args []string, kv *KeyValueStore) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'GET' command"
	}
	key := args[0]
	fmt.Print("GET key: ", kv.Strings)
	if value, exists := kv.Strings[key]; exists {
		fmt.Print("GET value: ", value)
		return value
	}
	return "Error, key doesn't exist"
}

func executeRPUSH(args []string, kv *KeyValueStore) string {
	if len(args) <= 1 {
		// return writeRESP([]byte{})
	} else {
		key := args[0]
		if _, exists := kv.Lists[key]; !exists {
			kv.Lists[key] = make([]string, 0)
		}

		for i := 1; i < len(args); i++ {
			kv.Lists[key] = append(kv.Lists[key], args[i])
		}
	}
	return "OK"
}

func executeLPUSH(args []string, kv *KeyValueStore) string {
	if len(args) <= 1 {
		// return writeRESP([]byte{})
	} else {
		key := args[0]
		if _, exists := kv.Lists[key]; !exists {
			kv.Lists[key] = make([]string, 0)
		}

		for i := 1; i < len(args); i++ {
			kv.Lists[key] = append([]string{args[i]}, kv.Lists[key]...)
		}
	}
	return "OK"
}

func executeRPOP(args []string, kv *KeyValueStore) string {
	key := args[0]
	if _, exists := kv.Lists[key]; !exists {
		return "Error, key does not exist"
	}

	len := len(kv.Lists[key])
	if len == 0 {
		return "Error: cannot pop from an empty slice"
	}
	element := kv.Lists[key][len-1]
	kv.Lists[key] = kv.Lists[key][:len-1]
	return element
}

func executeLPOP(args []string, kv *KeyValueStore) string {
	key := args[0]
	if _, exists := kv.Lists[key]; !exists {
		return "Error, key does not exist"
	}

	if len(kv.Lists[key]) == 0 {
		return "Error: cannot pop from an empty slice"
	}
	element := kv.Lists[key][0]
	kv.Lists[key] = kv.Lists[key][1:]
	return element
}

func executeLRANGE(args []string, kv *KeyValueStore) string {
	key := args[0]
	if _, exists := kv.Lists[key]; !exists {
		return "Error, key does not exist"
	}

	startIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return "Error, Start Index must be an integer"
	}
	endIndex, err := strconv.Atoi(args[2])
	if err != nil {
		return "Error, End Index must be an integer"
	}

	if startIndex < 0 {
		startIndex = len(kv.Lists[key]) + startIndex
	}
	if endIndex < 0 {
		endIndex = len(kv.Lists[key]) + endIndex
	}
	if startIndex > endIndex || startIndex >= len(kv.Lists[key]) {
		return "empty"
	}
	if endIndex >= len(kv.Lists[key]) {
		endIndex = len(kv.Lists[key]) - 1
	}

	str := strings.Join(kv.Lists[key][startIndex:endIndex+1], " ")
	return str
}

func executeINCR(args []string, kv *KeyValueStore) string {
	if len(args) > 1 {
		return "(error) ERR wrong number of arguments for 'INCR' command"
	}
	key := args[0]
	if _, exists := kv.Strings[key]; !exists {
		kv.Strings[key] = "1"
		return "(integer) 1"
	} else {
		num, err := strconv.Atoi(kv.Strings[key])
		if err != nil {
			return "(error) ERR value is not an integer"
		}
		num++
		kv.Strings[key] = strconv.Itoa(num)
		return fmt.Sprintf("(integer) %d", num)
	}
}

func executeDECR(args []string, kv *KeyValueStore) string {
	if len(args) > 1 {
		return "(error) ERR wrong number of arguments for 'DECR' command"
	}
	key := args[0]
	if _, exists := kv.Strings[key]; !exists {
		kv.Strings[key] = "-1"
		return "(integer) -1"
	} else {
		num, err := strconv.Atoi(kv.Strings[key])
		if err != nil {
			return "(error) ERR value is not an integer"
		}
		num--
		kv.Strings[key] = strconv.Itoa(num)
		return fmt.Sprintf("(integer) %d", num)
	}
}

func executeDEL(args []string, kv *KeyValueStore) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'DEL' command"
	}
	var num int
	for i := 0; i < len(args); i++ {
		if _, exists := kv.Strings[args[i]]; exists {
			delete(kv.Strings, args[i])
			num++
		}
		if _, exists := kv.Lists[args[i]]; exists {
			delete(kv.Lists, args[i])
			num++
		}
		if _, exists := kv.Hashes[args[i]]; exists {
			delete(kv.Hashes, args[i])
			num++
		}
	}
	return fmt.Sprintf("(integer) %d", num)
}

func executePING(args []string, kv *KeyValueStore) string {
	return "PONG"
}

func executeCommand(cmd Command, conn io.ReadWriter, kv *KeyValueStore) string {
	switch cmd.Name {
	case "SET":
		return executeSET(cmd.Args, kv)
	case "GET":
		return executeGET(cmd.Args, kv)
	case "SETNX":
		return executeSETNX(cmd.Args, kv)
	case "MSET":
		return executeMSET(cmd.Args, kv)
	case "MGET":
		return executeMGET(cmd.Args, kv)
	case "INCR":
		return executeINCR(cmd.Args, kv)
	case "DECR":
		return executeDECR(cmd.Args, kv)
	case "RPUSH":
		return executeRPUSH(cmd.Args, kv)
	case "LPUSH":
		return executeLPUSH(cmd.Args, kv)
	case "RPOP":
		return executeRPOP(cmd.Args, kv)
	case "LPOP":
		return executeLPOP(cmd.Args, kv)
	case "LRANGE":
		return executeLRANGE(cmd.Args, kv)
	case "DEL":
		return executeDEL(cmd.Args, kv)
	default:
		return executePING(cmd.Args, kv)
	}
}

func executeCommands(conn io.ReadWriter, commands []Command, kv *KeyValueStore) {
	// var response []byte
	// buf := bytes.NewBuffer(response)
	for _, cmd := range commands {
		// buf.Write(executeCommand(cmd, conn, kv))
		response := executeCommand(cmd, conn, kv)
		fmt.Print("print response: ", response)
		if response != "" {
			err := WriteBulkString(response, conn)
			if err != nil {
				fmt.Println("Error writing response:", err)
				break
			}
		}
	}
}
