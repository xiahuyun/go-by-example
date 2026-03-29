package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	raft "go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

type command struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type stateMachineSnapshot struct {
	KV           map[string]string `json:"kv"`
	AppliedIndex uint64            `json:"applied_index"`
}

type walRecord struct {
	HardState string   `json:"hard_state,omitempty"`
	Entries   []string `json:"entries,omitempty"`
}

type bootState struct {
	hasState bool
	hard     raftpb.HardState
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
}

type diskStorage struct {
	dir      string
	walPath  string
	snapPath string
	mu       sync.Mutex
}

func newDiskStorage(baseDir string, nodeID uint64) *diskStorage {
	dir := filepath.Join(baseDir, fmt.Sprintf("node-%d", nodeID))
	return &diskStorage{
		dir:      dir,
		walPath:  filepath.Join(dir, "wal.log"),
		snapPath: filepath.Join(dir, "snapshot.bin"),
	}
}

func (d *diskStorage) ensureDir() error {
	return os.MkdirAll(d.dir, 0o755)
}

func encodeProto(m interface{ Marshal() ([]byte, error) }) (string, error) {
	data, err := m.Marshal()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func decodeHardState(encoded string) (raftpb.HardState, error) {
	var hs raftpb.HardState
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return hs, err
	}
	return hs, hs.Unmarshal(raw)
}

func decodeEntry(encoded string) (raftpb.Entry, error) {
	var ent raftpb.Entry
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ent, err
	}
	return ent, ent.Unmarshal(raw)
}

func (d *diskStorage) appendWAL(hard raftpb.HardState, entries []raftpb.Entry) error {
	if raft.IsEmptyHardState(hard) && len(entries) == 0 {
		return nil
	}
	if err := d.ensureDir(); err != nil {
		return err
	}

	rec := walRecord{}
	if !raft.IsEmptyHardState(hard) {
		encoded, err := encodeProto(&hard)
		if err != nil {
			return err
		}
		rec.HardState = encoded
	}

	if len(entries) > 0 {
		rec.Entries = make([]string, 0, len(entries))
		for _, ent := range entries {
			encoded, err := encodeProto(&ent)
			if err != nil {
				return err
			}
			rec.Entries = append(rec.Entries, encoded)
		}
	}

	line, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := os.OpenFile(d.walPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func (d *diskStorage) saveSnapshot(snapshot raftpb.Snapshot) error {
	if raft.IsEmptySnap(snapshot) {
		return nil
	}
	if err := d.ensureDir(); err != nil {
		return err
	}

	raw, err := snapshot.Marshal()
	if err != nil {
		return err
	}

	tmp := d.snapPath + ".tmp"

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, d.snapPath)
}

func (d *diskStorage) load() (bootState, error) {
	var state bootState

	if err := d.ensureDir(); err != nil {
		return state, err
	}

	if raw, err := os.ReadFile(d.snapPath); err == nil {
		if err := state.snapshot.Unmarshal(raw); err != nil {
			return state, fmt.Errorf("decode snapshot: %w", err)
		}
		if !raft.IsEmptySnap(state.snapshot) {
			state.hasState = true
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return state, err
	}

	f, err := os.Open(d.walPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 8*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rec walRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return state, fmt.Errorf("decode wal line: %w", err)
		}

		if rec.HardState != "" {
			hs, err := decodeHardState(rec.HardState)
			if err != nil {
				return state, fmt.Errorf("decode hard state: %w", err)
			}
			state.hard = hs
			state.hasState = true
		}

		for _, encoded := range rec.Entries {
			ent, err := decodeEntry(encoded)
			if err != nil {
				return state, fmt.Errorf("decode entry: %w", err)
			}
			state.entries = append(state.entries, ent)
			state.hasState = true
		}
	}

	if err := scanner.Err(); err != nil {
		return state, err
	}

	if !raft.IsEmptySnap(state.snapshot) {
		snapIdx := state.snapshot.Metadata.Index
		filtered := make([]raftpb.Entry, 0, len(state.entries))
		for _, ent := range state.entries {
			if ent.Index > snapIdx {
				filtered = append(filtered, ent)
			}
		}
		state.entries = filtered
	}

	return state, nil
}

type raftNode struct {
	id      uint64
	node    raft.Node
	storage *raft.MemoryStorage
	disk    *diskStorage

	mu           sync.RWMutex
	kv           map[string]string
	appliedIndex uint64
	stopped      bool

	confState         raftpb.ConfState
	lastSnapshotIndex uint64
	snapshotEvery     uint64
	snapshotCatchUp   uint64

	stopCh chan struct{}
	doneCh chan struct{}
}

type cluster struct {
	nodes map[uint64]*raftNode
}

func newPersistentCluster(ids []uint64, baseDir string) (*cluster, error) {
	peers := make([]raft.Peer, 0, len(ids))
	for _, id := range ids {
		peers = append(peers, raft.Peer{ID: id})
	}

	c := &cluster{
		nodes: make(map[uint64]*raftNode, len(ids)),
	}

	for _, id := range ids {
		n, err := newRaftNode(id, peers, baseDir)
		if err != nil {
			c.stopAll()
			return nil, err
		}
		c.nodes[id] = n
	}

	for _, n := range c.nodes {
		go n.run(c)
	}

	return c, nil
}

func newRaftNode(id uint64, peers []raft.Peer, baseDir string) (*raftNode, error) {
	disk := newDiskStorage(baseDir, id)
	boot, err := disk.load()
	if err != nil {
		return nil, fmt.Errorf("load node %d state: %w", id, err)
	}

	storage := raft.NewMemoryStorage()
	n := &raftNode{
		id:                id,
		storage:           storage,
		disk:              disk,
		kv:                make(map[string]string),
		snapshotEvery:     5,
		snapshotCatchUp:   2,
		stopCh:            make(chan struct{}),
		doneCh:            make(chan struct{}),
		lastSnapshotIndex: 0,
	}

	if !raft.IsEmptySnap(boot.snapshot) {
		if err := storage.ApplySnapshot(boot.snapshot); err != nil {
			return nil, fmt.Errorf("apply boot snapshot node %d: %w", id, err)
		}
		if err := n.restoreStateMachineFromSnapshot(boot.snapshot.Data); err != nil {
			return nil, fmt.Errorf("restore snapshot data node %d: %w", id, err)
		}
		n.lastSnapshotIndex = boot.snapshot.Metadata.Index
		n.confState = boot.snapshot.Metadata.ConfState
		if n.appliedIndex == 0 {
			n.appliedIndex = boot.snapshot.Metadata.Index
		}
	}

	if len(boot.entries) > 0 {
		if err := storage.Append(boot.entries); err != nil {
			return nil, fmt.Errorf("append boot entries node %d: %w", id, err)
		}
	}
	if !raft.IsEmptyHardState(boot.hard) {
		storage.SetHardState(boot.hard)
	}

	cfg := &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
		CheckQuorum:     true,
		PreVote:         true,
		Applied:         n.appliedIndex,
	}

	if boot.hasState {
		n.node = raft.RestartNode(cfg)
		log.Printf("[node %d] start with RestartNode (recovered from disk)", id)
	} else {
		n.node = raft.StartNode(cfg, peers)
		n.confState = raftpb.ConfState{Voters: peerIDs(peers)}
		log.Printf("[node %d] start with StartNode (brand new)", id)
	}

	return n, nil
}

func peerIDs(peers []raft.Peer) []uint64 {
	ids := make([]uint64, 0, len(peers))
	for _, p := range peers {
		ids = append(ids, p.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (n *raftNode) run(c *cluster) {
	defer close(n.doneCh)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.node.Tick()

		case rd := <-n.node.Ready():
			if err := n.persistReady(rd); err != nil {
				log.Printf("[node %d] persist ready failed: %v", n.id, err)
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				if err := n.storage.ApplySnapshot(rd.Snapshot); err != nil {
					log.Printf("[node %d] apply snapshot failed: %v", n.id, err)
				} else {
					if err := n.restoreStateMachineFromSnapshot(rd.Snapshot.Data); err != nil {
						log.Printf("[node %d] restore snapshot data failed: %v", n.id, err)
					}
					n.lastSnapshotIndex = rd.Snapshot.Metadata.Index
					n.confState = rd.Snapshot.Metadata.ConfState
				}
			}
			if !raft.IsEmptyHardState(rd.HardState) {
				n.storage.SetHardState(rd.HardState)
			}
			if err := n.storage.Append(rd.Entries); err != nil {
				log.Printf("[node %d] append entries failed: %v", n.id, err)
			}

			c.deliver(rd.Messages)
			n.applyCommitted(rd.CommittedEntries)
			n.node.Advance()

		case <-n.stopCh:
			n.node.Stop()
			return
		}
	}
}

func (n *raftNode) persistReady(rd raft.Ready) error {
	if !raft.IsEmptySnap(rd.Snapshot) {
		if err := n.disk.saveSnapshot(rd.Snapshot); err != nil {
			return err
		}
	}
	return n.disk.appendWAL(rd.HardState, rd.Entries)
}

func (n *raftNode) applyCommitted(entries []raftpb.Entry) {
	if len(entries) == 0 {
		return
	}

	for _, ent := range entries {
		n.mu.Lock()
		n.appliedIndex = ent.Index
		n.mu.Unlock()

		switch ent.Type {
		case raftpb.EntryNormal:
			if len(ent.Data) == 0 {
				continue
			}

			var cmd command
			if err := json.Unmarshal(ent.Data, &cmd); err != nil {
				log.Printf("[node %d] decode normal entry failed: %v", n.id, err)
				continue
			}

			n.mu.Lock()
			n.kv[cmd.Key] = cmd.Value
			n.mu.Unlock()

			log.Printf("[node %d] apply index=%d term=%d set %s=%s", n.id, ent.Index, ent.Term, cmd.Key, cmd.Value)

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			if err := cc.Unmarshal(ent.Data); err != nil {
				log.Printf("[node %d] decode conf change failed: %v", n.id, err)
				continue
			}
			cs := n.node.ApplyConfChange(cc)
			if cs != nil {
				n.confState = *cs
			}

		case raftpb.EntryConfChangeV2:
			var cc raftpb.ConfChangeV2
			if err := cc.Unmarshal(ent.Data); err != nil {
				log.Printf("[node %d] decode conf change v2 failed: %v", n.id, err)
				continue
			}
			cs := n.node.ApplyConfChange(cc)
			if cs != nil {
				n.confState = *cs
			}
		}
	}

	n.maybeSnapshot()
}

func (n *raftNode) maybeSnapshot() {
	n.mu.RLock()
	applied := n.appliedIndex
	n.mu.RUnlock()

	if applied == 0 {
		return
	}
	if applied <= n.lastSnapshotIndex || applied-n.lastSnapshotIndex < n.snapshotEvery {
		return
	}

	data, err := n.stateMachineSnapshotData()
	if err != nil {
		log.Printf("[node %d] marshal snapshot data failed: %v", n.id, err)
		return
	}

	snap, err := n.storage.CreateSnapshot(applied, &n.confState, data)
	if err != nil {
		if errors.Is(err, raft.ErrSnapOutOfDate) {
			return
		}
		log.Printf("[node %d] create snapshot failed: %v", n.id, err)
		return
	}

	if err := n.disk.saveSnapshot(snap); err != nil {
		log.Printf("[node %d] save snapshot failed: %v", n.id, err)
		return
	}
	n.lastSnapshotIndex = snap.Metadata.Index

	compactTo := uint64(1)
	if snap.Metadata.Index > n.snapshotCatchUp {
		compactTo = snap.Metadata.Index - n.snapshotCatchUp
	}
	if err := n.storage.Compact(compactTo); err != nil {
		if errors.Is(err, raft.ErrCompacted) || errors.Is(err, raft.ErrSnapOutOfDate) {
			return
		}
		log.Printf("[node %d] compact failed: %v", n.id, err)
		return
	}

	log.Printf("[node %d] local snapshot index=%d compactTo=%d", n.id, snap.Metadata.Index, compactTo)
}

func (n *raftNode) stateMachineSnapshotData() ([]byte, error) {
	n.mu.RLock()
	kvCopy := make(map[string]string, len(n.kv))
	for k, v := range n.kv {
		kvCopy[k] = v
	}
	applied := n.appliedIndex
	n.mu.RUnlock()

	return json.Marshal(stateMachineSnapshot{
		KV:           kvCopy,
		AppliedIndex: applied,
	})
}

func (n *raftNode) restoreStateMachineFromSnapshot(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var snap stateMachineSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	if snap.KV == nil {
		snap.KV = make(map[string]string)
	}

	n.mu.Lock()
	n.kv = snap.KV
	if snap.AppliedIndex > 0 {
		n.appliedIndex = snap.AppliedIndex
	}
	n.mu.Unlock()
	return nil
}

func (n *raftNode) stop() {
	n.mu.Lock()
	if n.stopped {
		n.mu.Unlock()
		return
	}
	n.stopped = true
	close(n.stopCh)
	n.mu.Unlock()

	<-n.doneCh
}

func (n *raftNode) isStopped() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stopped
}

func (n *raftNode) kvSnapshot() (uint64, map[string]string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	snapshot := make(map[string]string, len(n.kv))
	for k, v := range n.kv {
		snapshot[k] = v
	}
	return n.appliedIndex, snapshot
}

func (n *raftNode) getValue(key string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	v, ok := n.kv[key]
	return v, ok
}

func (c *cluster) deliver(msgs []raftpb.Message) {
	for _, msg := range msgs {
		target := c.nodes[msg.To]
		if target == nil || target.isStopped() {
			continue
		}

		if err := target.node.Step(context.Background(), msg); err != nil {
			log.Printf("deliver %s %d->%d failed: %v", msg.Type, msg.From, msg.To, err)
		}
	}
}

func (c *cluster) stopAll() {
	for _, n := range c.nodes {
		n.stop()
	}
}

func (c *cluster) stopNode(id uint64) error {
	n := c.nodes[id]
	if n == nil {
		return fmt.Errorf("node %d not found", id)
	}
	n.stop()
	return nil
}

func (c *cluster) activeNodeIDs() []uint64 {
	ids := make([]uint64, 0, len(c.nodes))
	for id, n := range c.nodes {
		if n.isStopped() {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (c *cluster) waitLeader(timeout time.Duration) (uint64, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, id := range c.activeNodeIDs() {
			status := c.nodes[id].node.Status()
			if status.RaftState == raft.StateLeader {
				return id, nil
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	return 0, errors.New("leader election timeout")
}

func (c *cluster) chooseFollower(leaderID uint64) (uint64, error) {
	for _, id := range c.activeNodeIDs() {
		if id == leaderID {
			continue
		}
		return id, nil
	}
	return 0, errors.New("no follower found")
}

func (c *cluster) proposeKV(leaderID uint64, key, value string) error {
	leader := c.nodes[leaderID]
	if leader == nil {
		return fmt.Errorf("leader node %d not found", leaderID)
	}
	if leader.isStopped() {
		return fmt.Errorf("leader node %d is stopped", leaderID)
	}

	data, err := json.Marshal(command{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return err
	}

	return leader.node.Propose(context.Background(), data)
}

func (c *cluster) waitValueOnActiveNodes(key, expect string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok := true
		for _, id := range c.activeNodeIDs() {
			v, found := c.nodes[id].getValue(key)
			if !found || v != expect {
				ok = false
				break
			}
		}
		if ok {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("key %s not converged to %s on active nodes", key, expect)
}

func (c *cluster) printState(title string) {
	log.Println(title)
	for _, id := range c.activeNodeIDs() {
		n := c.nodes[id]
		applied, kv := n.kvSnapshot()
		log.Printf("node=%d applied=%d kv={%s}", id, applied, formatKV(kv))
	}
}

func formatKV(kv map[string]string) string {
	if len(kv) == 0 {
		return ""
	}
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, kv[k]))
	}
	return strings.Join(parts, ", ")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	baseDir := filepath.Join(".", "data")
	if err := os.RemoveAll(baseDir); err != nil {
		log.Fatal(err)
	}

	ids := []uint64{1, 2, 3}

	log.Println("=== phase 1: start brand-new cluster ===")
	c1, err := newPersistentCluster(ids, baseDir)
	if err != nil {
		log.Fatal(err)
	}

	leader1, err := c1.waitLeader(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("phase 1 leader: node %d", leader1)

	initialWrites := []command{
		{Key: "x", Value: "10"},
		{Key: "y", Value: "20"},
		{Key: "user", Value: "alice"},
	}
	for _, w := range initialWrites {
		if err := c1.proposeKV(leader1, w.Key, w.Value); err != nil {
			log.Fatalf("phase 1 propose %s=%s failed: %v", w.Key, w.Value, err)
		}
		if err := c1.waitValueOnActiveNodes(w.Key, w.Value, 2*time.Second); err != nil {
			log.Fatal(err)
		}
	}
	c1.printState("phase 1 committed:")
	c1.stopAll()

	log.Println("=== phase 2: restart from WAL + snapshot ===")
	c2, err := newPersistentCluster(ids, baseDir)
	if err != nil {
		log.Fatal(err)
	}
	defer c2.stopAll()

	leader2, err := c2.waitLeader(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("phase 2 leader after restart: node %d", leader2)

	for _, w := range initialWrites {
		if err := c2.waitValueOnActiveNodes(w.Key, w.Value, 2*time.Second); err != nil {
			log.Fatalf("recovery check failed for %s=%s: %v", w.Key, w.Value, err)
		}
	}
	c2.printState("after restart recovered:")

	if err := c2.proposeKV(leader2, "z", "30"); err != nil {
		log.Fatalf("phase 2 propose z=30 failed: %v", err)
	}
	if err := c2.waitValueOnActiveNodes("z", "30", 2*time.Second); err != nil {
		log.Fatal(err)
	}
	c2.printState("after restart new write:")

	followerID, err := c2.chooseFollower(leader2)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stop follower node %d after restart", followerID)
	if err := c2.stopNode(followerID); err != nil {
		log.Fatal(err)
	}

	if err := c2.proposeKV(leader2, "k", "v"); err != nil {
		log.Fatalf("phase 2 propose k=v failed: %v", err)
	}
	if err := c2.waitValueOnActiveNodes("k", "v", 2*time.Second); err != nil {
		log.Fatal(err)
	}
	c2.printState("one follower down, majority still commits:")

	log.Println("demo finished")
}
