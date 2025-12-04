package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {
	// 1. 创建消费者
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalln("Failed to create consumer:", err)
	}
	defer consumer.Close()

	// 2. 订阅指定主题的特定分区（这里订阅分区 0）
	topic := "test-topic"
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalln("Failed to create partition consumer:", err)
	}
	defer partitionConsumer.Close()

	// 3. 设置信号监听，优雅退出
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 4. 循环消费消息
	fmt.Println("Consumer started. Press Ctrl+C to exit.")
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Received message: Key=%s, Value=%s, Offset=%d\n", string(msg.Key), string(msg.Value), msg.Offset)
		case <-sigchan:
			fmt.Println("\nInterrupt is received, shutting down...")
			return
		}
	}
}
