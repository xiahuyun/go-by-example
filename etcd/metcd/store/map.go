package store

import (
	"encoding/gob"
	"log"
	"strings"
)

type mapKVStore struct {
	proposeC chan<- string
	kv       map[string]string
}

func (s *mapKVStore) GetSnapshot() ([]byte, error) {
	// This is a placeholder implementation. In a real implementation, you would serialize the map to bytes.
	return []byte{}, nil
}

func (s *mapKVStore) Propose(key, value string) {
	var buf strings.Builder
	if err := gob.NewEncoder(&buf).Encode(kv{Key: key, Val: value}); err != nil {
		log.Fatalf("Failed to encode kv: %v", err)
	}
	s.proposeC <- buf.String()
}
