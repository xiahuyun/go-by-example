package main

import (
	"fmt"
	"os"
)

func main() {
	str := "0000000000000078-000000004937eb8a.wal"
	var seq, index uint64
	if _, err := fmt.Sscanf(str, "%016x-%016x.wal", &seq, &index); err != nil {
		os.Exit(1)
	}

	fmt.Println(seq, index)
}
