package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

type RESPParser struct {
	c    io.ReadWriter
	buf  *bytes.Buffer
	tbuf []byte
}

func readSimpleString(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

func readError(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

func readInt64(c io.ReadWriter, buf *bytes.Buffer) (int64, error) {
	s, err := readStringUntilSr(buf)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func readLength(buf *bytes.Buffer) (int64, error) {
	s, err := readStringUntilSr(buf)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func readBulkString(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	len, err := readLength(buf)
	if err != nil {
		return "", err
	}

	bulkStr := make([]byte, len)
	_, err = buf.Read(bulkStr)
	if err != nil {
		return "", err
	}

	// moving buffer pointer by 2 for CRLF
	buf.ReadByte()
	buf.ReadByte()

	return string(bulkStr), nil
}

func readArray(c io.ReadWriter, buf *bytes.Buffer, rp *RESPParser) (interface{}, error) {
	count, err := readLength(buf)
	if err != nil {
		return nil, err
	}
	fmt.Print("array length: ", count)
	var elems []interface{} = make([]interface{}, count)
	for i := range elems {
		elem, err := rp.parseSingle()
		if err != nil {
			return nil, err
		}
		elems[i] = elem
	}
	return elems, nil
}

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func readStringUntilSr(buf *bytes.Buffer) (string, error) {
	s, err := buf.ReadString('\r')
	if err != nil {
		return "", err
	}
	// increamenting to skip `\n`
	buf.ReadByte()
	return s[:len(s)-1], nil
}

func (rp *RESPParser) parseSingle() (interface{}, error) {
	b, err := rp.buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+':
		return readSimpleString(rp.c, rp.buf)
	case '-':
		return readError(rp.c, rp.buf)
	case ':':
		return readInt64(rp.c, rp.buf)
	case '$':
		return readBulkString(rp.c, rp.buf)
	case '*':
		fmt.Print("parsing array\n")
		return readArray(rp.c, rp.buf, rp)
	}

	return nil, errors.New("Invalid input")
}

// ParseRESP parses a RESP message and returns a list of commands and their arguments
func ParseRESP(conn io.ReadWriter, inputBuf []byte) ([]Command, error) {

	fmt.Print("inputBuf ", string(inputBuf))
	var b []byte
	var buf *bytes.Buffer = bytes.NewBuffer(b)
	buf.Write(inputBuf)
	rp := &RESPParser{
		c:    conn,
		buf:  buf,
		tbuf: make([]byte, 50*512),
	}

	value, err := rp.parseSingle()
	if err != nil {
		return nil, err
	}

	var cmds = make([]Command, 0)
	tokens, err := toArrayString(value.([]interface{}))
	if err != nil {
		return nil, err
	}
	cmd := strings.ToUpper(tokens[0])
	cmds = append(cmds, Command{
		Name: cmd,
		Args: tokens[1:],
	})
	return cmds, nil
}
