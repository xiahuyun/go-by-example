package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	_ "google.golang.org/grpc/balancer/roundrobin"

	"etcd-grpc-roundrobin/echo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

func main() {
	var addrsArg string
	var totalCalls int
	flag.StringVar(&addrsArg, "addrs", "127.0.0.1:50051,127.0.0.1:50052", "comma-separated gRPC backend addresses")
	flag.IntVar(&totalCalls, "n", 8, "number of requests")
	flag.Parse()

	rawAddrs := strings.Split(addrsArg, ",")
	addrs := make([]resolver.Address, 0, len(rawAddrs))
	for _, a := range rawAddrs {
		addr := strings.TrimSpace(a)
		if addr == "" {
			continue
		}
		addrs = append(addrs, resolver.Address{Addr: addr})
	}
	if len(addrs) == 0 {
		log.Fatal("no backend addresses provided")
	}

	r := manual.NewBuilderWithScheme("manual-test")
	r.InitialState(resolver.State{Addresses: addrs})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"manual-test:///echo.service.local",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(r),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("round_robin client is ready")
	for i := 1; i <= totalCalls; i++ {
		callCtx, callCancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := echo.CallEcho(callCtx, conn, fmt.Sprintf("hello-%d", i))
		callCancel()
		if err != nil {
			log.Printf("call %d failed: %v", i, err)
			continue
		}
		fmt.Printf("call %02d -> %s\n", i, resp)
		time.Sleep(200 * time.Millisecond)
	}
}
