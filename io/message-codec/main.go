package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errExceedSizeLimit = errors.New("message size exceeds limit")

func main() {
	var stream bytes.Buffer

	encodeMessage(&stream, []byte("raft MsgSnap metadata"))
	stream.WriteString("boltdb-page-1\nboltdb-page-2\n")

	msg, err := decodeLimit(&stream, 1024)
	if err != nil {
		panic(err)
	}

	fmt.Printf("decoded message: %q\n", msg)

	rest, err := io.ReadAll(&stream)
	if err != nil {
		panic(err)
	}
	fmt.Printf("remaining stream:\n%s", rest)

	fmt.Println("\ntry a too-small limit:")
	limited := newStream("this message is too large", "snapshot bytes")
	_, err = decodeLimit(limited, 8)
	fmt.Println("decode error:", err)

	fmt.Println("\ntry a short payload:")
	short := bytes.NewBuffer(nil)
	_ = binary.Write(short, binary.BigEndian, uint64(10))
	short.WriteString("abc")
	_, err = decodeLimit(short, 1024)
	fmt.Println("decode error:", err)

	fmt.Println("\ntry a normal payload:")
	normal := bytes.NewBuffer(nil)
	encodeMessage(normal, []byte("raft MsgApp msg"))
	// _ = binary.Write(normal, binary.BigEndian, uint64(10))
	normal.WriteString("abc")
	msg, err = decodeLimit(normal, 1024)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded message: %q\n", msg)

	rest, err = io.ReadAll(normal)
	if err != nil {
		panic(err)
	}
	fmt.Printf("remaining normal:\n%s", rest)
}

func encodeMessage(w io.Writer, msg []byte) {
	_ = binary.Write(w, binary.BigEndian, uint64(len(msg)))
	_, _ = w.Write(msg)
}

func decodeLimit(r io.Reader, numBytes uint64) ([]byte, error) {
	var l uint64
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return nil, err
	}
	fmt.Println("binary.Read length prefix:", l)

	if l > numBytes {
		return nil, errExceedSizeLimit
	}

	buf := make([]byte, int(l))
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func newStream(message string, rest string) *bytes.Buffer {
	var stream bytes.Buffer
	encodeMessage(&stream, []byte(message))
	stream.WriteString(strings.TrimSpace(rest))
	return &stream
}
