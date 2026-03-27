package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("new etcd client faiiled with error: ", err)
	}
	defer cli.Close()

	// put data to etcd
	result, err := cli.Put(context.TODO(), "sample_key", "sample_value")
	if err != nil {
		switch err {
		case context.Canceled:
			log.Fatalf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Fatalf("client-side error: %v", err)
		default:
			log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
	}

	fmt.Println(result)

	// get data from etcd
	resp, err := cli.Get(context.TODO(), "sample_key")
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	// delete data from etcd
	res, err := cli.Delete(context.TODO(), "sample_key")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	// get data from etcd
	resp, err = cli.Get(context.TODO(), "sample_key")
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	// update date from etcd
	_, err = cli.Put(context.TODO(), "sample_key", "sample_value")
	if err != nil {
		log.Fatalf("failed to put data into etcd with error: %v", err)
	}
	resp, err = cli.Get(context.TODO(), "sample_key")
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	_, err = cli.Put(context.TODO(), "sample_key", "demo_value")
	if err != nil {
		log.Fatalf("failed to put data into etcd with error: %v", err)
	}
	resp, err = cli.Get(context.TODO(), "sample_key")
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	// watch data from etcd
	rch := cli.Watch(context.Background(), "sample_key")
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}
