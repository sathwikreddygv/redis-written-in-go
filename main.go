package main

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	conn           net.Conn
	isTxn          bool
	queuedCommands []Command
}

func handleConnection(client *Client, kv *KeyValueStore) {
	conn := client.conn
	defer conn.Close()
	fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())
	for {
		var buf []byte = make([]byte, 50*1024)
		n, err := conn.Read(buf[:])
		fmt.Print(string(buf[:n]), " ", n, len(string(buf[:n])))
		if err != nil {
			conn.Close()
			fmt.Print("client disconnected with ", err.Error())
			break
		}
		commands, err := ParseRESP(conn, buf[:n])
		fmt.Print(commands)
		executeCommands(client, commands, kv)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":6369")
	if err != nil {
		fmt.Println("error listening")
	}
	kv := &KeyValueStore{
		Strings:     make(map[string]string),
		Lists:       make(map[string][]string),
		Hashes:      make(map[string]map[string]string),
		Expirations: make(map[string]time.Time),
	}
	err = loadFromDisk("aof.gob", kv)
	if err != nil {
		fmt.Println("Error loading from file:", err)
	}

	// Start the periodic save function in a goroutine
	go periodicSave(kv, 10*time.Second)

	for {
		c, err := listener.Accept()
		client := &Client{conn: c, isTxn: false}
		if err != nil {
			fmt.Println("error accepting connection")
		}
		fmt.Println("client connected")
		go handleConnection(client, kv)
	}
}
