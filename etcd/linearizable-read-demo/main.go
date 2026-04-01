package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type notifier struct {
	c   chan struct{}
	err error
}

func newNotifier() *notifier {
	return &notifier{c: make(chan struct{})}
}

func (n *notifier) notify(err error) {
	n.err = err
	close(n.c)
}

type applyWait struct {
	mu   sync.Mutex
	last uint64
	m    map[uint64]chan struct{}
}

func newApplyWait() *applyWait {
	return &applyWait{m: make(map[uint64]chan struct{})}
}

func (w *applyWait) Wait(deadline uint64) <-chan struct{} {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.last >= deadline {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	ch := w.m[deadline]
	if ch == nil {
		ch = make(chan struct{})
		w.m[deadline] = ch
	}
	return ch
}

func (w *applyWait) Trigger(deadline uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.last = deadline
	for d, ch := range w.m {
		if d <= deadline {
			delete(w.m, d)
			close(ch)
		}
	}
}

type Server struct {
	readMu       sync.RWMutex
	readwaitc    chan struct{}
	readNotifier *notifier
	stopping     chan struct{}

	applyWait    *applyWait
	appliedIndex atomic.Uint64
	reqID        atomic.Uint64
}

func newServer() *Server {
	return &Server{
		readwaitc:    make(chan struct{}, 1),
		readNotifier: newNotifier(),
		stopping:     make(chan struct{}),
		applyWait:    newApplyWait(),
	}
}

// 每个请求各自调用：拿到当前 batch 的 notifier，然后等待它被 loop 统一唤醒。
func (s *Server) linearizableReadNotify(ctx context.Context, name string) error {
	s.readMu.RLock()
	nc := s.readNotifier
	s.readMu.RUnlock()

	select {
	case s.readwaitc <- struct{}{}:
	default:
	}

	logf("%s: waiting on its own notify handle", name)
	select {
	case <-nc.c:
		logf("%s: unblocked (batch done)", name)
		return nc.err
	case <-ctx.Done():
		return ctx.Err()
	case <-s.stopping:
		return fmt.Errorf("server stopping")
	}
}

// 单个共享 loop：按批次处理读请求。
func (s *Server) linearizableReadLoop() {
	for {
		select {
		case <-s.readwaitc:
		case <-s.stopping:
			return
		}

		reqID := s.reqID.Add(1)
		nextnr := newNotifier()
		s.readMu.Lock()
		nr := s.readNotifier
		s.readNotifier = nextnr
		s.readMu.Unlock()

		// 模拟从 leader 拿到 ReadIndex，固定为 10，强制先等待 apply。
		confirmedIndex := uint64(10)
		applied := s.appliedIndex.Load()
		logf("loop #%d: confirmedIndex=%d, applied=%d", reqID, confirmedIndex, applied)

		if applied < confirmedIndex {
			logf("loop #%d: blocked at applyWait.Wait(%d)", reqID, confirmedIndex)
			select {
			case <-s.applyWait.Wait(confirmedIndex):
			case <-s.stopping:
				return
			}
			logf("loop #%d: applyWait released", reqID)
		}

		nr.notify(nil)
		logf("loop #%d: notified one batch of readers", reqID)
	}
}

func logf(format string, args ...any) {
	fmt.Printf("%s | %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
}

func main() {
	s := newServer()
	go s.linearizableReadLoop()

	// 模拟 apply 线程慢慢推进 appliedIndex。
	go func() {
		for i := uint64(1); i <= 10; i++ {
			time.Sleep(250 * time.Millisecond)
			s.appliedIndex.Store(i)
			s.applyWait.Trigger(i)
			logf("apply: Trigger(%d)", i)
		}
	}()

	// 第一批请求：会一起被同一轮 loop 放行。
	for i := 1; i <= 3; i++ {
		idx := i
		go func() {
			name := fmt.Sprintf("read-A%d", idx)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.linearizableReadNotify(ctx, name); err != nil {
				logf("%s: err=%v", name, err)
			}
		}()
		time.Sleep(80 * time.Millisecond)
	}

	// 稍后到来的第二批请求：会挂在 next notifier，等下一轮 loop。
	time.Sleep(700 * time.Millisecond)
	for i := 1; i <= 2; i++ {
		idx := i
		go func() {
			name := fmt.Sprintf("read-B%d", idx)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.linearizableReadNotify(ctx, name); err != nil {
				logf("%s: err=%v", name, err)
			}
		}()
		time.Sleep(60 * time.Millisecond)
	}

	time.Sleep(4 * time.Second)
	close(s.stopping)
}

