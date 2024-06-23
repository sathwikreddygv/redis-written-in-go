package main

import (
	"fmt"
	"net"
	"time"
)

func handleConnection(conn net.Conn, kv *KeyValueStore) {
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
		executeCommands(conn, commands, kv)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":6369")
	if err != nil {
		fmt.Println("error listening")
	}
	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection")
		}
		fmt.Println("client connected")
		kv := &KeyValueStore{
			Strings:     make(map[string]string),
			Lists:       make(map[string][]string),
			Hashes:      make(map[string]map[string]string),
			Expirations: make(map[string]time.Time),
		}
		go handleConnection(c, kv)
	}
}
