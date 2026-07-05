package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.etcd.io/etcd/server/v3/storage/wal/walpb"

	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

type commit struct {
	data       []string
	applyDoneC chan<- struct{}
}

type raftNode struct {
	proposeC    <-chan string
	confChangeC <-chan raftpb.ConfChange
	commitC     chan<- *commit
	errorC      chan<- error

	id          int
	peers       []string
	join        bool
	waldir      string
	snapdir     string
	getSnapshot func() ([]byte, error)
	snapCount   uint64
	stopc       chan struct{}
	httpstopc   chan struct{}
	httpdonec   chan struct{}

	logger *zap.Logger

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter

	wal *wal.WAL

	node        raft.Node
	raftStorage *raft.MemoryStorage
	transport   *rafthttp.Transport

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64
}

var defaultSnapshotCount uint64 = 10000

func newRaftNode(
	id int,
	peers []string,
	join bool,
	getSnapshot func() ([]byte, error),
	proposeC chan string,
	confChangeC chan raftpb.ConfChange,
) (chan<- *commit, <-chan error, <-chan struct{}) {
	commitC := make(chan *commit)
	errorC := make(chan error)

	rn := &raftNode{
		proposeC:    proposeC,
		commitC:     commitC,
		errorC:      errorC,
		id:          id,
		peers:       peers,
		join:        join,
		waldir:      fmt.Sprintf("metcd-%d", id),
		snapdir:     fmt.Sprintf("metcd-%d-snap", id),
		getSnapshot: getSnapshot,
		snapCount:   defaultSnapshotCount,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),

		logger: zap.NewExample(),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
	}

	go rn.startRaft(confChangeC)
	return commitC, errorC, rn.snapshotterReady
}

func (rn *raftNode) startRaft(confChangeC chan raftpb.ConfChange) {
	if !fileutil.Exist(rn.snapdir) {
		if err := os.Mkdir(rn.snapdir, 0o750); err != nil {
			log.Fatalf("raftnode: cannot create dir for snapshot (%v)", err)
		}
	}

	rn.snapshotter = snap.New(zap.NewExample(), rn.snapdir)

	oldwal := wal.Exist(rn.waldir)
	rn.wal = rn.replayWAL()

	// signal replay has finished
	rn.snapshotterReady <- rn.snapshotter

	rpeers := make([]raft.Peer, len(rn.peers))
	for i := range rn.peers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:                        uint64(rn.id),
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   rn.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}

	if oldwal || rn.join {
		rn.node = raft.RestartNode(c)
	} else {
		rn.node = raft.StartNode(c, rpeers)
	}

	rn.transport = &rafthttp.Transport{
		Logger:      rn.logger,
		ID:          types.ID(rn.id),
		ClusterID:   0x1000,
		Raft:        rn,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(rn.id)),
		ErrorC:      make(chan error),
	}

	rn.transport.Start()
	for i := range rn.peers {
		if i+1 != rn.id {
			rn.transport.AddPeer(types.ID(i+1), []string{rn.peers[i]})
		}
	}

	go rn.serveRaft()
	go rn.serveChannels()
}

func (rn *raftNode) replayWAL() *wal.WAL {
	log.Printf("replaying WAL of member %d\n", rn.id)
	snapshot := rn.loadSnapshot()
	w := rn.openWAL(snapshot)
	_, state, ents, err := w.ReadAll()
	if err != nil {
		log.Fatalf("raftnode: failed to read WAL (%v)", err)
	}
	rn.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		rn.raftStorage.ApplySnapshot(*snapshot)
	}
	rn.raftStorage.SetHardState(state)
	rn.raftStorage.Append(ents)

	return w
}

func (rn *raftNode) loadSnapshot() *raftpb.Snapshot {
	if wal.Exist(rn.waldir) {
		walSnaps, err := wal.ValidSnapshotEntries(rn.logger, rn.waldir)
		if err != nil {
			log.Fatalf("raftnode: failed to list snapshot entries (%v)", err)
		}
		snapshot, err := rn.snapshotter.LoadNewestAvailable(walSnaps)
		if err != nil {
			log.Fatalf("raftnode: failed to load snapshot (%v)", err)
		}
		return snapshot
	}
	return &raftpb.Snapshot{}
}

func (rn *raftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rn.waldir) {
		if err := os.Mkdir(rn.waldir, 0o750); err != nil {
			log.Fatalf("raftnode: cannot create dir for WAL (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), rn.waldir, nil)
		if err != nil {
			log.Fatalf("raftnode: failed to create WAL (%v)", err)
		}
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index = snapshot.Metadata.Index
		walsnap.Term = snapshot.Metadata.Term
	}
	log.Printf("loading WAL at term %d and index %d\n", walsnap.Term, walsnap.Index)
	w, err := wal.Open(zap.NewExample(), rn.waldir, walsnap)
	if err != nil {
		log.Fatalf("raftnode: failed to open WAL (%v)", err)
	}
	return w
}

func (rn *raftNode) serveRaft() {
	url, err := url.Parse(rn.peers[rn.id-1])
	if err != nil {
		log.Fatalf("raftnode: failed to parse peer URL (%v)", err)
	}

	ln, err := newStoppableListener(url.Host, rn.httpstopc)
	if err != nil {
		log.Fatalf("raftnode: failed to create listener (%v)", err)
	}

	err = (&http.Server{Handler: rn.transport.Handler()}).Serve(ln)
	select {
	case <-rn.httpstopc:
	default:
		log.Fatalf("raftnode: HTTP server error (%v)", err)
	}
	close(rn.httpdonec)
}

func (rn *raftNode) serveChannels() {
	snap, err := rn.raftStorage.Snapshot()
	if err != nil {
		log.Fatalf("raftnode: failed to get snapshot (%v)", err)
	}
	rn.confState = snap.Metadata.ConfState
	rn.snapshotIndex = snap.Metadata.Index
	rn.appliedIndex = snap.Metadata.Index

	defer rn.wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		confChangeCount := uint64(0)

		for rn.proposeC != nil && rn.confChangeC != nil {
			select {
			case prop, ok := <-rn.proposeC:
				if !ok {
					rn.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					rn.node.Propose(context.TODO(), []byte(prop))
				}
			case conf, ok := <-rn.confChangeC:
				if !ok {
					rn.confChangeC = nil
				} else {
					confChangeCount++
					conf.ID = confChangeCount
					rn.node.ProposeConfChange(context.TODO(), conf)
				}
			}
		}
		close(rn.stopc)
	}()

	for {
		select {
		case <-ticker.C:
			rn.node.Tick()
		case rd := <-rn.node.Ready():
			if !raft.IsEmptySnap(rd.Snapshot) {
				rn.saveSnap(rd.Snapshot)
			}
			rn.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				rn.raftStorage.ApplySnapshot(rd.Snapshot)
				rn.publishSnapshot(rd.Snapshot)
			}
		case <-rn.stopc:
			rn.stop()
			return
		}
	}
}

func (rn *raftNode) saveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index:     snap.Metadata.Index,
		Term:      snap.Metadata.Term,
		ConfState: &snap.Metadata.ConfState,
	}

	if err := rn.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	if err := rn.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}

	return rn.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (rn *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("publishing snapshot at index %d", rn.snapshotIndex)
	defer log.Printf("finished publishing snapshot at index %d", rn.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rn.snapshotIndex {
		log.Fatalf("snapshot index %d is older than current snapshot index %d", snapshotToSave.Metadata.Index, rn.snapshotIndex)
		return
	}

	rn.commitC <- nil

	rn.confState = snapshotToSave.Metadata.ConfState
	rn.snapshotIndex = snapshotToSave.Metadata.Index
	rn.appliedIndex = snapshotToSave.Metadata.Index
}

func (rn *raftNode) stop() {
	rn.stopHTTP()
	close(rn.commitC)
	close(rn.errorC)
	rn.node.Stop()
}

func (rn *raftNode) stopHTTP() {
	rn.transport.Stop()
	close(rn.httpstopc)
	<-rn.httpdonec
}
