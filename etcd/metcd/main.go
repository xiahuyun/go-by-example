package main

import (
	"flag"
	"go-by-example/metcd/store"
	"strings"

	"go.etcd.io/etcd/raft/raftpb"
)

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	var kvs store.KVStore
	getSnapshot := func() ([]byte, error) { return kvs.GetSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	kvs = store.NewKVStore(<-snapshotterReady, proposeC, commitC, errorC, "map")

	// the key-value http handler will propose updates to raft
	serveHTTPKVAPI(kvs, *kvport, confChangeC, errorC)
}
