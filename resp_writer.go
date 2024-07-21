package main

import (
	"io"
	"strconv"
)

type Writer struct {
	w io.Writer
}

func WriteBulkString(s string, conn io.ReadWriter) error {
	val := []byte(s)
	conn.Write([]byte("$"))
	conn.Write(strconv.AppendUint(nil, uint64(len(val)), 10))
	conn.Write([]byte("\r\n"))
	conn.Write(val)
	_, err := conn.Write([]byte("\r\n"))
	return err
}

func WriteArray(result []string, conn io.ReadWriter) error {
	conn.Write([]byte("*"))
	conn.Write(strconv.AppendUint(nil, uint64(len(result)), 10))
	conn.Write([]byte("\r\n"))
	var err error
	for i := 0; i < len(result); i++ {
		err = WriteBulkString(result[i], conn)
	}
	return err
}
