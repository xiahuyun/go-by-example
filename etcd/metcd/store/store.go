package store

type KVStore interface {
	GetSnapshot() ([]byte, error)
	Propose(key, value string)
}

type kv struct {
	Key string
	Val string
}

func NewKVStore(snapshotterReady struct{}, proposeC chan string, commitC <-chan *string, errorC <-chan error, kvType string) KVStore {
	if kvType == "map" {
		return &mapKVStore{proposeC: proposeC, kv: make(map[string]string)}
	}

	return nil
}
