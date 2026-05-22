package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	revBytesLen       = 17 // 8(main) + 1('_') + 8(sub)
	markedRevBytesLen = 18 // revBytesLen + 1(tombstone mark)
	markTombstone     = byte('t')
)

type config struct {
	endpoints    []string
	snapshotPath string
	prefix       string
}

type revRecord struct {
	Main      int64
	Sub       int64
	Tombstone bool
	KV        mvccpb.KeyValue
}

func main() {
	cfg := loadConfig()

	fmt.Printf("endpoints: %v\n", cfg.endpoints)
	fmt.Printf("snapshot path: %s\n", cfg.snapshotPath)
	fmt.Printf("demo key prefix: %s\n\n", cfg.prefix)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("create etcd client failed: %v", err)
	}
	defer cli.Close()

	if err := cleanupPrefix(cli, cfg.prefix); err != nil {
		log.Fatalf("cleanup prefix failed: %v", err)
	}

	// Seed one key so that delete in the same txn leaves a tombstone record.
	seedKey := cfg.prefix + "/to-delete"
	if _, err := put(cli, seedKey, "seed"); err != nil {
		log.Fatalf("seed put failed: %v", err)
	}

	txnResp, err := runSingleTxnWithMultipleRequests(cli, cfg.prefix)
	if err != nil {
		log.Fatalf("run txn failed: %v", err)
	}
	if !txnResp.Succeeded {
		log.Fatalf("txn compare failed unexpectedly")
	}

	txnMainRev := txnResp.Header.Revision
	fmt.Printf("txn committed at main revision: %d\n\n", txnMainRev)

	fmt.Println("== API view (only main revision is exposed) ==")
	if err := printAPIView(cli, cfg.prefix); err != nil {
		log.Fatalf("print API view failed: %v", err)
	}
	fmt.Println()

	fmt.Println("== Backend key bucket view (main/sub visible) ==")
	if err := saveSnapshot(cli, cfg.snapshotPath); err != nil {
		log.Fatalf("save snapshot failed: %v", err)
	}
	recs, err := readRecordsByMainRevision(cfg.snapshotPath, cfg.prefix, txnMainRev)
	if err != nil {
		log.Fatalf("read backend records failed: %v", err)
	}
	if len(recs) == 0 {
		log.Fatalf("no backend record found at main=%d (db=%s)", txnMainRev, cfg.snapshotPath)
	}
	printRecords(recs)

	fmt.Println("\nDone. You can compare API mod_revision with backend sub ordering.")
}

func loadConfig() config {
	endpoints := parseEndpoints(os.Getenv("ETCD_ENDPOINTS"))
	if len(endpoints) == 0 {
		endpoints = []string{"127.0.0.1:2379"}
	}

	snapshotPath := strings.TrimSpace(os.Getenv("ETCD_SNAPSHOT"))
	if snapshotPath == "" {
		snapshotPath = ".tmp/txn-rev-snapshot.db"
	}

	prefix := strings.TrimSpace(os.Getenv("DEMO_PREFIX"))
	if prefix == "" {
		prefix = "demo/txn-rev"
	}

	return config{
		endpoints:    endpoints,
		snapshotPath: snapshotPath,
		prefix:       prefix,
	}
}

func parseEndpoints(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func cleanupPrefix(cli *clientv3.Client, prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Delete(ctx, prefix+"/", clientv3.WithPrefix())
	return err
}

func put(cli *clientv3.Client, key, val string) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return cli.Put(ctx, key, val)
}

func runSingleTxnWithMultipleRequests(cli *clientv3.Client, prefix string) (*clientv3.TxnResponse, error) {
	guardKey := prefix + "/guard"
	keyA := prefix + "/a"
	keyB := prefix + "/b"
	deleteKey := prefix + "/to-delete"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Txn(ctx).
		If(
			clientv3.Compare(clientv3.Version(guardKey), "=", 0),
			clientv3.Compare(clientv3.Version(deleteKey), ">", 0),
		).
		Then(
			clientv3.OpPut(guardKey, "created-in-txn-v2"),
			clientv3.OpPut(keyA, "A1"),
			clientv3.OpPut(keyB, "B1"),
			clientv3.OpDelete(deleteKey),
		).
		Else(
			clientv3.OpGet(guardKey),
		).
		Commit()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func printAPIView(cli *clientv3.Client, prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix+"/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}

	fmt.Printf("response header revision(main): %d\n", resp.Header.Revision)
	for _, kv := range resp.Kvs {
		fmt.Printf(
			"key=%-22s value=%-16s createRev=%d modRev=%d version=%d\n",
			string(kv.Key), string(kv.Value), kv.CreateRevision, kv.ModRevision, kv.Version,
		)
	}
	fmt.Println("note: API only exposes main revision (no sub field).")
	return nil
}

func saveSnapshot(cli *clientv3.Client, snapshotPath string) error {
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rc, err := cli.Snapshot(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.Create(snapshotPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return err
	}
	return f.Sync()
}

func readRecordsByMainRevision(dbPath, prefix string, targetMain int64) ([]revRecord, error) {
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{ReadOnly: true, Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt db %s failed: %w", dbPath, err)
	}
	defer db.Close()

	records := make([]revRecord, 0)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("key"))
		if b == nil {
			return errors.New("bucket 'key' not found")
		}

		c := b.Cursor()
		for rk, rv := c.First(); rk != nil; rk, rv = c.Next() {
			main, sub, tombstone, ok := decodeRevisionKey(rk)
			if !ok || main != targetMain {
				continue
			}

			var kv mvccpb.KeyValue
			if err := kv.Unmarshal(rv); err != nil {
				return fmt.Errorf("unmarshal kv failed (main=%d sub=%d): %w", main, sub, err)
			}
			if !bytes.HasPrefix(kv.Key, []byte(prefix+"/")) {
				continue
			}

			records = append(records, revRecord{
				Main:      main,
				Sub:       sub,
				Tombstone: tombstone,
				KV:        kv,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Sub < records[j].Sub
	})
	return records, nil
}

func decodeRevisionKey(raw []byte) (main, sub int64, tombstone, ok bool) {
	if len(raw) != revBytesLen && len(raw) != markedRevBytesLen {
		return 0, 0, false, false
	}
	if raw[8] != '_' {
		return 0, 0, false, false
	}

	main = int64(binary.BigEndian.Uint64(raw[:8]))
	sub = int64(binary.BigEndian.Uint64(raw[9:17]))
	if len(raw) == markedRevBytesLen && raw[17] == markTombstone {
		tombstone = true
	}
	return main, sub, tombstone, true
}

func printRecords(records []revRecord) {
	for _, r := range records {
		fmt.Printf(
			"rev(main=%d, sub=%d, tombstone=%v) key=%-22s value=%-16s createRev=%d modRev=%d version=%d\n",
			r.Main,
			r.Sub,
			r.Tombstone,
			string(r.KV.Key),
			string(r.KV.Value),
			r.KV.CreateRevision,
			r.KV.ModRevision,
			r.KV.Version,
		)
	}
}
