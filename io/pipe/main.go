package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

func main() {
	fmt.Println("== 1. io.Pipe: writer writes, reader blocks until data arrives ==")
	manualPipe()

	fmt.Println("\n== 2. io.Copy: keep reading from src and writing to dst until EOF ==")
	copyDemo()

	fmt.Println("\n== 3. snapshot-style: goroutine streams data into a pipe ==")
	snapshotStyle()
}

func manualPipe() {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for _, chunk := range []string{"hello ", "from ", "pipe"} {
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("writer: Write(%q)\n", chunk)
			_, _ = pw.Write([]byte(chunk))
		}
	}()

	buf := make([]byte, 5)
	for {
		n, err := pr.Read(buf)
		if n > 0 {
			fmt.Printf("reader: Read -> %q\n", string(buf[:n]))
		}
		if err == io.EOF {
			fmt.Println("reader: EOF, writer closed")
			return
		}
		if err != nil {
			fmt.Println("reader error:", err)
			return
		}
	}
}

func copyDemo() {
	src := strings.NewReader("raft snapshot bytes")
	var dst bytes.Buffer

	n, err := io.Copy(&dst, src)
	if err != nil {
		fmt.Println("copy error:", err)
		return
	}

	fmt.Printf("copied %d bytes, dst=%q\n", n, dst.String())
}

func snapshotStyle() {
	reader := newSnapshotReaderCloser(strings.NewReader("db-page-1\ndb-page-2\ndb-page-3\n"))
	defer reader.Close()

	var received bytes.Buffer
	n, err := io.Copy(&received, reader)
	if err != nil {
		fmt.Println("receiver copy error:", err)
		return
	}

	fmt.Printf("receiver: copied %d bytes from pipe\n%s", n, received.String())
}

func newSnapshotReaderCloser(snapshot io.Reader) io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		n, err := io.Copy(pw, snapshot)
		if err == nil {
			fmt.Printf("snapshot writer: streamed %d bytes\n", n)
		} else {
			fmt.Printf("snapshot writer: failed after %d bytes: %v\n", n, err)
		}
		_ = pw.CloseWithError(err)
	}()

	return pr
}
