package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	endpoints := parseEndpoints()
	fmt.Printf("connecting etcd endpoints: %v\n\n", endpoints)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("create etcd client failed: %v", err)
	}
	defer cli.Close()

	key := "demo/mvcc/order"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cli.Delete(ctx, key)
	if err != nil {
		log.Fatalf("cleanup key failed: %v", err)
	}

	put1 := mustPut(cli, key, "v1-created")
	put2 := mustPut(cli, key, "v2-updated")
	put3 := mustPut(cli, key, "v3-updated")

	fmt.Println("== 1) latest version ==")
	latestKV := mustGetSingle(cli, key)
	printKV("latest", latestKV)
	fmt.Printf("cluster revision now: %d\n\n", put3.Header.Revision)

	fmt.Println("== 2) historical reads by revision (MVCC snapshot) ==")
	h1 := mustGetSingleAtRev(cli, key, put1.Header.Revision)
	h2 := mustGetSingleAtRev(cli, key, put2.Header.Revision)
	h3 := mustGetSingleAtRev(cli, key, put3.Header.Revision)
	printKV(fmt.Sprintf("at rev=%d", put1.Header.Revision), h1)
	printKV(fmt.Sprintf("at rev=%d", put2.Header.Revision), h2)
	printKV(fmt.Sprintf("at rev=%d", put3.Header.Revision), h3)
	fmt.Println()

	fmt.Println("== 3) compare-and-swap with ModRevision ==")
	txnResp, err := cli.Txn(context.Background()).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", latestKV.ModRevision)).
		Then(clientv3.OpPut(key, "v4-by-cas")).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		log.Fatalf("txn failed: %v", err)
	}
	fmt.Printf("cas with latest mod revision succeeded: %v\n", txnResp.Succeeded)
	if txnResp.Succeeded {
		afterCAS := mustGetSingle(cli, key)
		printKV("after cas", afterCAS)
		latestKV = afterCAS
	}
	fmt.Println()

	fmt.Println("== 4) stale CAS should fail ==")
	staleTxnResp, err := cli.Txn(context.Background()).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", put1.Header.Revision)).
		Then(clientv3.OpPut(key, "should-not-write")).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		log.Fatalf("stale txn failed: %v", err)
	}
	fmt.Printf("cas with stale revision succeeded: %v (expect false)\n\n", staleTxnResp.Succeeded)

	fmt.Println("== 5) watch from a specific revision ==")
	watchStartRev := put2.Header.Revision + 1
	wctx, wcancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer wcancel()

	wch := cli.Watch(wctx, key, clientv3.WithRev(watchStartRev))
	_ = mustPut(cli, key, "v5-watch-event")

	for wr := range wch {
		if wr.Err() != nil {
			log.Fatalf("watch error: %v", wr.Err())
		}
		for _, ev := range wr.Events {
			fmt.Printf("watch event @rev=%d type=%s value=%s\n", ev.Kv.ModRevision, ev.Type, ev.Kv.Value)
		}
		break
	}
	fmt.Println()

	fmt.Println("== 6) compact old revisions ==")
	compactRev := put2.Header.Revision
	_, err = cli.Compact(context.Background(), compactRev)
	if err != nil {
		log.Fatalf("compact failed: %v", err)
	}
	fmt.Printf("compacted to revision: %d\n", compactRev)

	_, err = cli.Get(context.Background(), key, clientv3.WithRev(put1.Header.Revision))
	if errors.Is(err, rpctypes.ErrCompacted) {
		fmt.Printf("read at rev=%d failed as expected: %v\n", put1.Header.Revision, err)
	} else if err != nil {
		fmt.Printf("read at rev=%d failed (non-compacted error): %v\n", put1.Header.Revision, err)
	} else {
		fmt.Printf("read at rev=%d unexpectedly succeeded\n", put1.Header.Revision)
	}

	fmt.Println("\nMVCC demo finished.")
}

func parseEndpoints() []string {
	raw := os.Getenv("ETCD_ENDPOINTS")
	if strings.TrimSpace(raw) == "" {
		return []string{"localhost:2379", "localhost:22379", "localhost:32379"}
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"localhost:2379", "localhost:22379", "localhost:32379"}
	}
	return out
}

func mustPut(cli *clientv3.Client, key, val string) *clientv3.PutResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Put(ctx, key, val)
	if err != nil {
		log.Fatalf("put key=%s failed: %v", key, err)
	}
	fmt.Printf("put key=%s value=%s (cluster rev=%d)\n", key, val, resp.Header.Revision)
	return resp
}

func mustGetSingle(cli *clientv3.Client, key string) *mvccpb.KeyValue {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, key)
	if err != nil {
		log.Fatalf("get key=%s failed: %v", key, err)
	}
	if len(resp.Kvs) == 0 {
		log.Fatalf("get key=%s returned empty", key)
	}
	return resp.Kvs[0]
}

func mustGetSingleAtRev(cli *clientv3.Client, key string, rev int64) *mvccpb.KeyValue {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, key, clientv3.WithRev(rev))
	if err != nil {
		log.Fatalf("get key=%s at rev=%d failed: %v", key, rev, err)
	}
	if len(resp.Kvs) == 0 {
		log.Fatalf("get key=%s at rev=%d returned empty", key, rev)
	}
	return resp.Kvs[0]
}

func printKV(label string, kv *mvccpb.KeyValue) {
	fmt.Printf(
		"%s -> key=%s value=%s createRev=%d modRev=%d version=%d\n",
		label, kv.Key, kv.Value, kv.CreateRevision, kv.ModRevision, kv.Version,
	)
}
