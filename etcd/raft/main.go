package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

type raftNode struct {
	id      uint64
	node    raft.Node
	storage *raft.MemoryStorage

	mu           sync.RWMutex
	kv           map[string]string
	appliedIndex uint64
	stopped      bool

	stopCh chan struct{}
	doneCh chan struct{}
}

type cluster struct {
	nodes map[uint64]*raftNode
}

func newCluster(ids []uint64) *cluster {
	peers := make([]raft.Peer, 0, len(ids))
	for _, id := range ids {
		peers = append(peers, raft.Peer{ID: id})
	}

	c := &cluster{
		nodes: make(map[uint64]*raftNode, len(ids)),
	}

	for _, id := range ids {
		storage := raft.NewMemoryStorage()
		cfg := &raft.Config{
			ID:              id,
			ElectionTick:    10,
			HeartbeatTick:   1,
			Storage:         storage,
			MaxSizePerMsg:   1024 * 1024,
			MaxInflightMsgs: 256,
			CheckQuorum:     true,
			PreVote:         true,
		}

		log.Printf("start node %d with peers %v", id, peers)
		rn := raft.StartNode(cfg, peers)
		log.Printf("node %d start with peers %v", id, peers)
		n := &raftNode{
			id:      id,
			node:    rn,
			storage: storage,
			kv:      make(map[string]string),
			stopCh:  make(chan struct{}),
			doneCh:  make(chan struct{}),
		}
		c.nodes[id] = n
	}

	for _, n := range c.nodes {
		go n.run(c)
	}

	return c
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
			log.Printf("[node %d] receive ready: commit=%d, entries=%d, messages=%d, snapshot=%v, hardstate=%v",
				n.id, len(rd.CommittedEntries), len(rd.Entries), len(rd.Messages), !raft.IsEmptySnap(rd.Snapshot), !raft.IsEmptyHardState(rd.HardState))
			if !raft.IsEmptySnap(rd.Snapshot) {
				if err := n.storage.ApplySnapshot(rd.Snapshot); err != nil {
					log.Printf("[node %d] apply snapshot failed: %v", n.id, err)
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

func (n *raftNode) applyCommitted(entries []raftpb.Entry) {
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
				log.Printf("[node %d] decode entry failed at index=%d: %v", n.id, ent.Index, err)
				continue
			}

			n.mu.Lock()
			n.kv[cmd.Key] = cmd.Value
			n.mu.Unlock()

			log.Printf("[node %d] apply index=%d term=%d set %s=%s", n.id, ent.Index, ent.Term, cmd.Key, cmd.Value)

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			if err := cc.Unmarshal(ent.Data); err != nil {
				log.Printf("[node %d] unmarshal conf change failed: %v", n.id, err)
				continue
			}
			n.node.ApplyConfChange(cc)
			log.Printf("[node %d] apply conf change: %+v", n.id, cc)
		}
	}
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

func (n *raftNode) getValue(key string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	v, ok := n.kv[key]
	return v, ok
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

func (c *cluster) deliver(msgs []raftpb.Message) {
	for _, msg := range msgs {
		to := c.nodes[msg.To]
		if to == nil || to.isStopped() {
			continue
		}

		if err := to.node.Step(context.Background(), msg); err != nil {
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
			st := c.nodes[id].node.Status()
			if st.RaftState == raft.StateLeader {
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
		st := c.nodes[id].node.Status()
		if st.RaftState != raft.StateLeader {
			return id, nil
		}
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

	log.Printf("propose %s=%s to leader node %d", key, value, leaderID)
	return leader.node.Propose(context.Background(), data)
}

func (c *cluster) waitValueOnActiveNodes(key, expect string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		allMatch := true
		for _, id := range c.activeNodeIDs() {
			n := c.nodes[id]
			got, ok := n.getValue(key)
			if !ok || got != expect {
				allMatch = false
				break
			}
		}

		if allMatch {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("value %s=%s not replicated to all active nodes in %s", key, expect, timeout)
}

func (c *cluster) printState() {
	for _, id := range c.activeNodeIDs() {
		n := c.nodes[id]
		applied, kv := n.kvSnapshot()
		log.Printf("node=%d appliedIndex=%d kv={%s}", id, applied, formatKV(kv))
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

	c := newCluster([]uint64{1, 2, 3})
	defer c.stopAll()

	leaderID, err := c.waitLeader(4 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("leader elected: node %d", leaderID)

	initialProposals := []command{
		{Key: "x", Value: "10"},
		{Key: "y", Value: "20"},
	}
	for _, p := range initialProposals {
		if err := c.proposeKV(leaderID, p.Key, p.Value); err != nil {
			log.Fatalf("propose %s=%s failed: %v", p.Key, p.Value, err)
		}
		if err := c.waitValueOnActiveNodes(p.Key, p.Value, 2*time.Second); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("after two proposals:")
	c.printState()

	followerID, err := c.chooseFollower(leaderID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stop one follower node %d (simulate node down)", followerID)
	if err := c.stopNode(followerID); err != nil {
		log.Fatal(err)
	}

	if err := c.proposeKV(leaderID, "z", "30"); err != nil {
		log.Fatalf("propose z=30 failed: %v", err)
	}
	if err := c.waitValueOnActiveNodes("z", "30", 2*time.Second); err != nil {
		log.Fatal(err)
	}

	log.Println("with one node down, remaining majority still commits:")
	c.printState()
}
