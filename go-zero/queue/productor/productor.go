package main

import (
	"context"
	"log"

	"github.com/zeromicro/go-queue/kq"
)

func main() {
	pusher := kq.NewPusher([]string{
		"127.0.0.1:9092",
	}, "test-topic",
		kq.WithSyncPush())

	if err := pusher.Push(context.Background(), "foo"); err != nil {
		log.Fatal(err)
	}

	// async push to wait for the message to be processed
	// time.Sleep(5 * time.Second)
}
