package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"etcd-grpc-roundrobin/echo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type echoServer struct {
	id string
}

func (s *echoServer) Echo(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(fmt.Sprintf("server=%s reply=%s", s.id, req.Value)), nil
}

func serve(addr string, id string, wg *sync.WaitGroup) {
	defer wg.Done()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s failed: %v", addr, err)
	}

	s := grpc.NewServer()
	echo.RegisterServiceServer(s, &echoServer{id: id})

	log.Printf("gRPC server %s listening on %s", id, addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve %s failed: %v", addr, err)
	}
}

func main() {
	var portsArg string
	flag.StringVar(&portsArg, "ports", "50051,50052", "comma-separated server ports")
	flag.Parse()

	parts := strings.Split(portsArg, ",")
	if len(parts) == 0 {
		log.Fatal("no ports configured")
	}

	var wg sync.WaitGroup
	for i, p := range parts {
		port := strings.TrimSpace(p)
		if port == "" {
			continue
		}
		wg.Add(1)
		addr := "127.0.0.1:" + port
		id := fmt.Sprintf("node-%d", i+1)
		go serve(addr, id, &wg)
	}

	log.Printf("servers are running, press Ctrl+C to exit")

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	<-sigC
	log.Printf("received shutdown signal")
}
