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
	if val == nil {
		// _, err := w.Write(nilBulk)
		// return err
	}
	conn.Write([]byte("$"))
	conn.Write(strconv.AppendUint(nil, uint64(len(val)), 10))
	conn.Write([]byte("\r\n"))
	conn.Write(val)
	_, err := conn.Write([]byte("\r\n"))
	return err
}
