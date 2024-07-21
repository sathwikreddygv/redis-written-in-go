package main

import (
	// "bytes"
	"fmt"
	"slices"

	// "io"
	"strconv"
	"strings"
	"sync"
	"time"
)

type KeyValueStore struct {
	Strings     map[string]string
	Lists       map[string][]string
	Hashes      map[string]map[string]string
	Expirations map[string]time.Time
	mu          sync.RWMutex
	// CurrentTx   *Transaction
	// connectedClients map[string]net.Conn
}

func checkExpiredStrings(key string, kv *KeyValueStore) bool {
	if expirationTime, exists := kv.Expirations[key]; exists {
		if time.Now().After(expirationTime) {
			delete(kv.Expirations, key)
			delete(kv.Strings, key)
			return true
		}
	}
	return false
}

func checkExpiredLists(key string, kv *KeyValueStore) bool {
	if expirationTime, exists := kv.Expirations[key]; exists {
		if time.Now().After(expirationTime) {
			delete(kv.Expirations, key)
			delete(kv.Lists, key)
			return true
		}
	}
	return false
}

func executeSET(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'SET' command"
	} else {
		key := args[0]
		value := args[1]
		if !client.isTxn {
			kv.mu.Lock()
			defer kv.mu.Unlock()
		}
		kv.Strings[key] = value
	}
	return "OK"
}

func executeSETNX(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'SETNX' command"
	} else {
		key := args[0]
		value := args[1]
		if !client.isTxn {
			kv.mu.Lock()
			defer kv.mu.Unlock()
		}
		if _, exists := kv.Strings[key]; !exists {
			kv.Strings[key] = value
		}
	}
	return "OK"
}

func executeMSET(args []string, kv *KeyValueStore, client *Client) string {
	if len(args)%2 != 0 {
		return "(error) ERR wrong number of arguments for 'MSET' command"
	} else {
		if !client.isTxn {
			kv.mu.Lock()
			defer kv.mu.Unlock()
		}
		for i := 0; i < len(args); i += 2 {
			key := args[i]
			value := args[i+1]
			kv.Strings[key] = value
		}
	}
	return "OK"
}

func executeMGET(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'MGET' command"
	}

	var final []string
	if !client.isTxn {
		kv.mu.RLock()
		defer kv.mu.RUnlock()
	}
	for i := 0; i < len(args); i++ {
		key := args[i]
		if checkExpiredStrings(key, kv) {
			return "(nil)"
		}
		if value, exists := kv.Strings[key]; exists {
			final = append(final, value)
		} else {
			final = append(final, "(nil)")
		}
	}

	return strings.Join(final, " ")
}

func executeGET(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'GET' command"
	}
	key := args[0]
	if !client.isTxn {
		kv.mu.RLock()
		defer kv.mu.RUnlock()
	}
	if checkExpiredStrings(key, kv) {
		return "(nil)"
	}
	if value, exists := kv.Strings[key]; exists {
		return value
	}
	return "Error, key doesn't exist"
}

func executeRPUSH(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'RPUSH' command"
	} else {
		if !client.isTxn {
			kv.mu.Lock()
			defer kv.mu.Unlock()
		}
		key := args[0]
		if _, exists := kv.Lists[key]; !exists || checkExpiredLists(key, kv) {
			kv.Lists[key] = make([]string, 0)
		}

		for i := 1; i < len(args); i++ {
			kv.Lists[key] = append(kv.Lists[key], args[i])
		}
	}
	return "OK"
}

func executeLPUSH(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'LPUSH' command"
	} else {
		if !client.isTxn {
			kv.mu.Lock()
			defer kv.mu.Unlock()
		}
		key := args[0]
		if _, exists := kv.Lists[key]; !exists || checkExpiredLists(key, kv) {
			kv.Lists[key] = make([]string, 0)
		}

		for i := 1; i < len(args); i++ {
			kv.Lists[key] = append([]string{args[i]}, kv.Lists[key]...)
		}
	}
	return "OK"
}

func executeRPOP(args []string, kv *KeyValueStore, client *Client) string {
	key := args[0]
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
	if _, exists := kv.Lists[key]; !exists || checkExpiredLists(key, kv) {
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

func executeLPOP(args []string, kv *KeyValueStore, client *Client) string {
	key := args[0]
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
	if _, exists := kv.Lists[key]; !exists || checkExpiredLists(key, kv) {
		return "Error, key does not exist"
	}

	if len(kv.Lists[key]) == 0 {
		return "Error: cannot pop from an empty slice"
	}
	element := kv.Lists[key][0]
	kv.Lists[key] = kv.Lists[key][1:]
	return element
}

func executeLRANGE(args []string, kv *KeyValueStore, client *Client) string {
	key := args[0]
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
	if _, exists := kv.Lists[key]; !exists || checkExpiredLists(key, kv) {
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

func executeINCR(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) > 1 {
		return "(error) ERR wrong number of arguments for 'INCR' command"
	}
	key := args[0]
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}

	if _, exists := kv.Lists[key]; exists {
		return "(error) WRONGTYPE Operation against a key holding the wrong kind of value"
	}

	if _, exists := kv.Hashes[key]; exists {
		return "(error) WRONGTYPE Operation against a key holding the wrong kind of value"
	}

	if _, exists := kv.Strings[key]; !exists || checkExpiredStrings(key, kv) {
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

func executeDECR(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) > 1 {
		return "(error) ERR wrong number of arguments for 'DECR' command"
	}
	key := args[0]
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
	if _, exists := kv.Strings[key]; !exists || checkExpiredStrings(key, kv) {
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

func executeDEL(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'DEL' command"
	}
	var num int
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
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
		delete(kv.Expirations, args[i])
	}
	return fmt.Sprintf("(integer) %d", num)
}

func executeEXPIRE(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) <= 1 {
		return "(error) ERR wrong number of arguments for 'EXPIRE' command"
	}
	key := args[0]
	expiration, err := strconv.Atoi(args[1])
	if err != nil {
		return "(error) ERR value is not an integer"
	}
	var key_exists bool
	if !client.isTxn {
		kv.mu.Lock()
		defer kv.mu.Unlock()
	}
	if _, exists := kv.Strings[key]; exists {
		key_exists = true
	} else if _, exists := kv.Lists[key]; exists {
		key_exists = true
	} else if _, exists := kv.Hashes[key]; exists {
		key_exists = true
	}

	if key_exists {
		kv.Expirations[key] = time.Now().Add(time.Duration(expiration) * time.Second)
		return "(integer) 1"
	}

	return "(integer) 0"
}

func executeTTL(args []string, kv *KeyValueStore, client *Client) string {
	if len(args) < 1 {
		return "(error) ERR wrong number of arguments for 'TTL' command"
	}
	key := args[0]
	if !client.isTxn {
		kv.mu.RLock()
		defer kv.mu.RUnlock()
	}
	var key_exists bool
	if _, exists := kv.Strings[key]; exists {
		key_exists = true
	} else if _, exists := kv.Lists[key]; exists {
		key_exists = true
	} else if _, exists := kv.Hashes[key]; exists {
		key_exists = true
	}

	if key_exists {
		if expirationTime, exists := kv.Expirations[key]; exists {
			if time.Now().Before(expirationTime) {
				ttl := time.Until(expirationTime).Seconds()
				return fmt.Sprintf("(integer) %d", int(ttl))
			}
			delete(kv.Expirations, key)
			delete(kv.Strings, key)
			delete(kv.Lists, key)
			delete(kv.Hashes, key)
			return "(integer) -2"
		}
		return "(integer) -1"
	}

	return "(integer) -2"
}

func executePING(args []string, kv *KeyValueStore, client *Client) string {
	return "PONG"
}

func executeSAVE(args []string, kv *KeyValueStore, client *Client) string {
	err := saveToDisk(kv)
	if err != nil {
		return "(error) ERR Could not save to disk"
	}
	return "OK"
}

func executeMULTI(args []string, kv *KeyValueStore, client *Client) string {
	if client.isTxn {
		return "(error) ERR MULTI calls can not be nested"
	} else {
		client.isTxn = true
	}
	return "OK"
}

func executeDISCARD(args []string, kv *KeyValueStore, client *Client) string {
	if client.isTxn {
		client.queuedCommands = []Command{}
		client.isTxn = false
	} else {
		return "(error) ERR DISCARD without MULTI"
	}
	return "OK"
}

func executeEXEC(args []string, kv *KeyValueStore, client *Client) {
	if client.isTxn {
		fmt.Print("lock takeinggggggg")
		kv.mu.Lock()
		fmt.Print("lock takenn")
		defer kv.mu.Unlock()

		execResult := []string{}
		for _, cmd := range client.queuedCommands {
			response := executeCommand(cmd, client, kv)
			if response != "" {
				execResult = append(execResult, response)
			}
		}
		client.isTxn = false
		if len(execResult) != 0 {
			err := WriteArray(execResult, client.conn)
			if err != nil {
				fmt.Println("Error writing response:", err)
			}
		} else {
			err := WriteBulkString("(empty array)", client.conn)
			if err != nil {
				fmt.Println("Error writing response:", err)
			}
		}
	} else {
		WriteBulkString("(error) ERR EXEC without MULTI", client.conn)
	}
}

func executeCommand(cmd Command, client *Client, kv *KeyValueStore) string {
	switch cmd.Name {
	case "SET":
		return executeSET(cmd.Args, kv, client)
	case "GET":
		return executeGET(cmd.Args, kv, client)
	case "SETNX":
		return executeSETNX(cmd.Args, kv, client)
	case "MSET":
		return executeMSET(cmd.Args, kv, client)
	case "MGET":
		return executeMGET(cmd.Args, kv, client)
	case "INCR":
		return executeINCR(cmd.Args, kv, client)
	case "DECR":
		return executeDECR(cmd.Args, kv, client)
	case "RPUSH":
		return executeRPUSH(cmd.Args, kv, client)
	case "LPUSH":
		return executeLPUSH(cmd.Args, kv, client)
	case "RPOP":
		return executeRPOP(cmd.Args, kv, client)
	case "LPOP":
		return executeLPOP(cmd.Args, kv, client)
	case "LRANGE":
		return executeLRANGE(cmd.Args, kv, client)
	case "DEL":
		return executeDEL(cmd.Args, kv, client)
	case "EXPIRE":
		return executeEXPIRE(cmd.Args, kv, client)
	case "TTL":
		return executeTTL(cmd.Args, kv, client)
	case "PING":
		return executePING(cmd.Args, kv, client)
	case "SAVE":
		return executeSAVE(cmd.Args, kv, client)
	case "MULTI":
		return executeMULTI(cmd.Args, kv, client)
	case "EXEC":
		executeEXEC(cmd.Args, kv, client)
		return "OK"
	case "DISCARD":
		return executeDISCARD(cmd.Args, kv, client)
	default:
		return "(error) ERR unknown command"
	}
}

func executeCommands(client *Client, commands []Command, kv *KeyValueStore) {
	// var response []byte
	// buf := bytes.NewBuffer(response)
	conn := client.conn
	for _, cmd := range commands {
		// buf.Write(executeCommand(cmd, conn, kv))
		response := ""
		splCommands := []string{"MULTI", "EXEC", "DISCARD"}
		fmt.Print(splCommands, cmd.Name)
		if !slices.Contains(splCommands, cmd.Name) && client.isTxn {
			client.queuedCommands = append(client.queuedCommands, cmd)
			response = "QUEUED"
		} else {
			response = executeCommand(cmd, client, kv)
		}
		if cmd.Name != "EXEC" {
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
}
