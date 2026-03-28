package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ServerConfig controls startup and runtime behavior.
type ServerConfig struct {
	StartupDelay time.Duration
	TickInterval time.Duration
	RunDuration  time.Duration
	FailStartup  bool
}

// AsyncServer is a small mock server that mimics etcd's async lifecycle channels.
type AsyncServer struct {
	readyCh chan struct{}
	stopCh  chan struct{}
	errCh   chan error

	closeFn context.CancelFunc
	once    sync.Once
}

func StartAsyncServer(cfg ServerConfig) (*AsyncServer, error) {
	if cfg.StartupDelay < 0 {
		return nil, errors.New("startup delay cannot be negative")
	}
	if cfg.TickInterval <= 0 {
		return nil, errors.New("tick interval must be > 0")
	}
	if cfg.RunDuration <= 0 {
		return nil, errors.New("run duration must be > 0")
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv := &AsyncServer{
		readyCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
		errCh:   make(chan error, 1),
		closeFn: cancel,
	}

	go srv.run(ctx, cfg)
	return srv, nil
}

func (s *AsyncServer) run(ctx context.Context, cfg ServerConfig) {
	startTimer := time.NewTimer(cfg.StartupDelay)
	defer startTimer.Stop()

	select {
	case <-ctx.Done():
		s.finish(nil)
		return
	case <-startTimer.C:
	}

	if cfg.FailStartup {
		s.finish(errors.New("startup failed: simulated error"))
		return
	}

	log.Println("server: ready")
	close(s.readyCh)

	ticker := time.NewTicker(cfg.TickInterval)
	defer ticker.Stop()

	stopTimer := time.NewTimer(cfg.RunDuration)
	defer stopTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			s.finish(nil)
			return
		case <-ticker.C:
			log.Println("server: handling background task")
		case <-stopTimer.C:
			log.Println("server: run duration reached, graceful shutdown")
			s.finish(nil)
			return
		}
	}
}

func (s *AsyncServer) finish(err error) {
	s.once.Do(func() {
		if err != nil {
			s.errCh <- err
		}
		close(s.stopCh)
		close(s.errCh)
	})
}

func (s *AsyncServer) Close() {
	s.closeFn()
}

func (s *AsyncServer) ReadyNotify() <-chan struct{} {
	return s.readyCh
}

func (s *AsyncServer) StopNotify() <-chan struct{} {
	return s.stopCh
}

func (s *AsyncServer) Err() <-chan error {
	return s.errCh
}

func registerInterruptHandler(closeFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("signal received: shutting down")
		closeFn()
	}()
}

// startAsyncServer mirrors etcd's startEtcd style:
// Start -> register close hook -> wait for Ready/Stop -> return StopNotify + Err channels.
func startAsyncServer(cfg ServerConfig) (<-chan struct{}, <-chan error, error) {
	srv, err := StartAsyncServer(cfg)
	if err != nil {
		return nil, nil, err
	}
	registerInterruptHandler(srv.Close)

	select {
	case <-srv.ReadyNotify():
		log.Println("launcher: server joined and is ready")
	case <-srv.StopNotify():
		log.Println("launcher: startup aborted")
	}

	return srv.StopNotify(), srv.Err(), nil
}

func main() {
	cfg := ServerConfig{
		StartupDelay: 1200 * time.Millisecond,
		TickInterval: 1 * time.Second,
		RunDuration:  5 * time.Second,
		FailStartup:  false,
	}

	stopCh, errCh, err := startAsyncServer(cfg)
	if err != nil {
		log.Fatalf("failed to start async server: %v", err)
	}

	select {
	case err, ok := <-errCh:
		if !ok {
			errCh = nil
			log.Println("launcher: server error channel closed without error, assuming normal shutdown")
			return
		}
		if err != nil {
			log.Printf("launcher: server error: %v", err)
		}
	case <-stopCh:
		log.Println("launcher: server stopped")
	}
}
