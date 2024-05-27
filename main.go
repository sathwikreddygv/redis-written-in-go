package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	buf := make([]byte, 1024)
	rec, err := conn.Read(buf)
	fmt.Println(rec)
	if err != nil {
		fmt.Println("error reading from client: ", err.Error())
		os.Exit(1)
	}
	resp := string(buf[:rec])
	fmt.Println(resp)
	if _, err := conn.Write([]byte(resp)); err != nil {
		fmt.Println("error writing to client:")
	}
}

func main() {
	listener, err := net.Listen("tcp", ":6369")
	if err != nil {
		fmt.Println("error listening")
	}
	c, err := listener.Accept()
	if err != nil {
		fmt.Println("error accepting connection")
	}
	fmt.Println("client connected")
	handleConnection(c)
}
