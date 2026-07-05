package main

import (
	"fmt"
	"go-by-example/metcd/store"
	"io"
	"log"
	"net/http"

	"go.etcd.io/etcd/raft/raftpb"
)

type httpKVAPI struct {
	store       store.KVStore
	confChangeC chan raftpb.ConfChange
}

func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	defer r.Body.Close()

	switch r.Method {
	case http.MethodPut:
		v, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on PUT (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}
		h.store.Propose(key, string(v))

		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		// read the value from the request body and propose it to raft
		// (not implemented in this snippet)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func serveHTTPKVAPI(kv store.KVStore, port int, confChangeC chan raftpb.ConfChange, errorC <-chan error) {
	// start the HTTP server
	srv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}
