package main

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gogo/protobuf/proto"
	bolt "go.etcd.io/bbolt"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

const defaultPageSize = 4
const defaultBucket = "key"

type KV struct {
	Rev int64
	Key string
	Val string
}

var dumpLineRE = regexp.MustCompile(`rev=(\d+)\s+key=("(?:\\.|[^"])*"|\S+)\s+value=(.*)$`)

func decodeMainRevision(rawRev []byte) int64 {
	if len(rawRev) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(rawRev[:8]))
}

func renderRawBytes(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if utf8.Valid(raw) {
		return strings.ReplaceAll(string(raw), "\n", `\n`)
	}
	return "base64:" + base64.StdEncoding.EncodeToString(raw)
}

func dumpSnapshot(snapshotPath string, bucketName string) (string, error) {
	db, err := bolt.Open(snapshotPath, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return "", fmt.Errorf("open snapshot %q failed: %w", snapshotPath, err)
	}
	defer db.Close()

	var b strings.Builder
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found in snapshot", bucketName)
		}

		cur := bucket.Cursor()
		for rawK, rawV := cur.First(); rawK != nil; rawK, rawV = cur.Next() {
			var item mvccpb.KeyValue
			if err := proto.Unmarshal(rawV, &item); err != nil {
				continue
			}

			rev := item.ModRevision
			if rev == 0 {
				rev = item.CreateRevision
			}
			if rev == 0 {
				rev = decodeMainRevision(rawK)
			}
			if rev == 0 {
				continue
			}

			key := renderRawBytes(item.Key)
			val := renderRawBytes(item.Value)
			fmt.Fprintf(&b, "rev=%d key=%q value=%s\n", rev, key, val)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if b.Len() == 0 {
		return "", errors.New("no mvcc key-value entries extracted from snapshot")
	}
	return b.String(), nil
}

func parseDump(output string) ([]KV, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	data := make([]KV, 0, 256)

	for scanner.Scan() {
		line := scanner.Text()
		m := dumpLineRE.FindStringSubmatch(line)
		if len(m) != 4 {
			continue
		}

		rev, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			continue
		}

		rawKey := m[2]
		key := strings.Trim(rawKey, `"`)
		if unquoted, err := strconv.Unquote(rawKey); err == nil {
			key = unquoted
		}
		val := m[3]
		data = append(data, KV{Rev: rev, Key: key, Val: val})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no KV rows parsed from dump output")
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i].Rev == data[j].Rev {
			return data[i].Key < data[j].Key
		}
		return data[i].Rev < data[j].Rev
	})

	return data, nil
}

func buildLeaves(data []KV, pageSize int) [][]KV {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	leaves := make([][]KV, 0, (len(data)+pageSize-1)/pageSize)
	for i := 0; i < len(data); i += pageSize {
		end := i + pageSize
		if end > len(data) {
			end = len(data)
		}
		leaves = append(leaves, data[i:end])
	}
	return leaves
}

func escapeRecordField(s string) string {
	r := strings.NewReplacer(
		`\\`, `\\\\`,
		`{`, `\\{`,
		`}`, `\\}`,
		`|`, `\\|`,
		`<`, `\\<`,
		`>`, `\\>`,
		`"`, `\\"`,
	)
	return r.Replace(s)
}

func generateDot(leaves [][]KV, showValue bool) string {
	var b strings.Builder
	b.WriteString("digraph BPlusTree {\n")
	b.WriteString("rankdir=TB;\n")
	b.WriteString("node [shape=record, style=filled, fillcolor=\"#e9f2ff\", color=\"#4d79ff\"];\n")
	b.WriteString("edge [color=\"#5c5c5c\"];\n")

	for i, leaf := range leaves {
		fields := make([]string, 0, len(leaf))
		for _, kv := range leaf {
			entry := fmt.Sprintf("%d:%s", kv.Rev, escapeRecordField(kv.Key))
			if showValue {
				entry = fmt.Sprintf("%s=%s", entry, escapeRecordField(kv.Val))
			}
			fields = append(fields, entry)
		}
		b.WriteString(fmt.Sprintf("leaf%d [label=\"%s\", fillcolor=\"#f7fbff\"];\n", i, strings.Join(fields, "|")))
	}

	for i := 0; i < len(leaves)-1; i++ {
		b.WriteString(fmt.Sprintf("leaf%d -> leaf%d [label=\"next\"];\n", i, i+1))
	}

	rootFields := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		rootFields = append(rootFields, fmt.Sprintf("%d", leaf[0].Rev))
	}
	if len(rootFields) > 0 {
		b.WriteString(fmt.Sprintf("root [label=\"%s\", fillcolor=\"#d7e6ff\"];\n", strings.Join(rootFields, "|")))
		for i := range leaves {
			b.WriteString(fmt.Sprintf("root -> leaf%d;\n", i))
		}
	}

	b.WriteString("}\n")
	return b.String()
}

func main() {
	snapshotPath := flag.String("snapshot", "snapshot.db", "path to etcd snapshot db")
	bucketName := flag.String("bucket", defaultBucket, "etcd mvcc bucket name")
	pageSize := flag.Int("page-size", defaultPageSize, "simulated leaf page size")
	dotPath := flag.String("out", "tree.dot", "output dot file path")
	dumpOutPath := flag.String("dump-out", "", "optional path to save intermediate dump text")
	showValue := flag.Bool("show-value", false, "include value in leaf node labels")
	flag.Parse()

	fmt.Printf("Dumping snapshot: %s (bucket=%s)\n", *snapshotPath, *bucketName)
	dump, err := dumpSnapshot(*snapshotPath, *bucketName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if *dumpOutPath != "" {
		if err := os.WriteFile(*dumpOutPath, []byte(dump), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote dump text: %s\n", *dumpOutPath)
	}

	fmt.Println("Parsing dump output...")
	data, err := parseDump(dump)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d kv entries\n", len(data))
	fmt.Println("Building simulated B+ tree leaves...")
	leaves := buildLeaves(data, *pageSize)
	fmt.Printf("Built %d leaves (page-size=%d)\n", len(leaves), *pageSize)

	fmt.Println("Generating Graphviz dot...")
	dot := generateDot(leaves, *showValue)
	if err := os.WriteFile(*dotPath, []byte(dot), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Done! DOT file: %s\n", *dotPath)
	fmt.Printf("Render with: dot -Tpng %s -o tree.png\n", *dotPath)
}
