package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
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

	var bytesRem int64 = len + 2 // 2 for \r\n
	bytesRem = bytesRem - int64(buf.Len())
	for bytesRem > 0 {
		tbuf := make([]byte, bytesRem)
		n, err := c.Read(tbuf)
		if err != nil {
			return "", nil
		}
		buf.Write(tbuf[:n])
		bytesRem = bytesRem - int64(n)
	}

	bulkStr := make([]byte, len)
	_, err = buf.Read(bulkStr)
	if err != nil {
		return "", err
	}

	// moving buffer pointer by 2 for \r and \n
	buf.ReadByte()
	buf.ReadByte()

	// reading `len` bytes as string
	return string(bulkStr), nil
}

// reads a RESP encoded array from data and returns
// the array, the delta, and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
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
	for {
		if bytes.Contains(rp.buf.Bytes(), []byte{'\r', '\n'}) {
			break
		}
		fmt.Print("parsing single\n")
		n, err := rp.c.Read(rp.tbuf)
		fmt.Print("parsed single ", n, "\n", string(rp.tbuf[:n]), "\n")
		// this condition needs to be explicitly added to ensure
		// we break the loop if we read `0` bytes from the socket
		// implying end of the input
		if n <= 0 {
			break
		}
		rp.buf.Write(rp.tbuf[:n])
		if err != nil {
			if err == io.EOF {
				fmt.Print("breaking at EOF\n")
				break
			}
			return nil, err
		}

		if bytes.Contains(rp.tbuf, []byte{'\r', '\n'}) {
			fmt.Print("breaking at CRLF\n")
			break
		}

		if rp.buf.Len() > (50 * 1024) {
			return nil, fmt.Errorf("input too long. max input can be %d bytes", 50*1024)
		}
	}
	fmt.Print("parsing single done\n")
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

	// this also captures the Cross Protocol Scripting attack.
	// Since we do not support simple strings, anything that does
	// not start with any of the above special chars will be a potential
	// attack.
	// Details: https://bou.ke/blog/hacking-developers/
	log.Println("possible cross protocol scripting attack detected. dropping the request.")
	return nil, errors.New("possible cross protocol scripting attack detected")
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
		tbuf: make([]byte, 512),
	}

	var values []interface{} = make([]interface{}, 0)
	for {
		fmt.Print("here")
		value, err := rp.parseSingle()
		fmt.Print("poarse once", value)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		if rp.buf.Len() == 0 {
			break
		}
	}
	var cmds = make([]Command, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmd := strings.ToUpper(tokens[0])
		cmds = append(cmds, Command{
			Name: cmd,
			Args: tokens[1:],
		})

		// if cmd == "ABORT" {
		// 	hasABORT = true
		// }
	}
	return cmds, nil
}
