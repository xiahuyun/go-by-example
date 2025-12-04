package main

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func main() {
	// 1. 配置生产者
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true          // 确保成功送达的消息会被返回
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有同步副本确认

	// 2. 创建同步生产者
	producers, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalln("Failed to create producer:", err)
	}
	defer producers.Close() // 确保程序退出前关闭生产者

	// 3. 构造并发送消息
	topic := "test-topic"
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder("Hello, Kafka from Go!"),
	}

	partition, offset, err := producers.SendMessage(msg)
	if err != nil {
		log.Fatalln("Failed to send message:", err)
	}

	// 4. 打印发送结果
	fmt.Printf("Message sent successfully! Partition: %d, Offset: %d\n", partition, offset)
}
