package main

import (
	"fmt"
	"net"
	"time"
)

func handleConnection(conn net.Conn, kv *KeyValueStore) {
	defer conn.Close()

	fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())
	// writer := bufio.NewWriter(conn)
	for {
		// commands, err := ParseRESP(conn)
		// if err != nil {
		// 	fmt.Printf("Error parsing RESP: %v\n", err)
		// 	return
		// }

		// for _, cmd := range commands {
		// 	fmt.Printf("Command: %s, Args: %v\n", cmd.Name, cmd.Args)
		// }
		// _, err = writer.WriteString("+OK")
		// if err != nil {
		// 	fmt.Println("Error writing:", err.Error())
		// 	return
		// }

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
		// if _, err := conn.Write([]byte(string(buf[:n]))); err != nil {
		// 	fmt.Print("err writing to conn:")
		// }

	}

	// Flush to ensure the data is sent to the client
	// err = writer.Flush()
	// if err != nil {
	// 	fmt.Println("Error flushing:", err.Error())
	// 	return
	// }
	// fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
}

func main() {
	listener, err := net.Listen("tcp", ":6369")
	for {
		if err != nil {
			fmt.Println("error listening")
		}
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection")
		}
		fmt.Println("client connected")
		kv := &KeyValueStore{
			Strings:                make(map[string]string),
			Lists:                  make(map[string][]string),
			Hashes:                 make(map[string]map[string]string),
			Sets:                   make(map[string]map[string]struct{}),
			SortedSets:             make(map[string][]sortedSetMember),
			Expirations:            make(map[string]time.Time),
			totalCommandsProcessed: 0,
			// connectedClients: make(map[string]net.Conn),
		}
		go handleConnection(c, kv)
	}
}
