package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// 创建 Redis 客户端
func createClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址
		Password: "",               // 无密码
		DB:       0,                // 默认数据库
	})

	return client
}

// 发布消息函数
func publish(client *redis.Client, channel string, message string) error {
	err := client.Publish(ctx, channel, message).Err()
	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}
	fmt.Printf("发布消息到频道 %s: %s\n", channel, message)
	return nil
}

// 订阅消息函数
func subscribe(client *redis.Client, channel string) {
	pubsub := client.Subscribe(ctx, channel)

	// 等待订阅成功
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Fatalf("订阅失败: %v", err)
	}

	fmt.Printf("已订阅频道: %s\n", channel)

	// 监听消息
	ch := pubsub.Channel()

	// 持续接收消息
	for msg := range ch {
		fmt.Printf("接收到来自频道 %s 的消息: %s\n", msg.Channel, msg.Payload)
		time.Sleep(1 * time.Second)
	}
}

func subscribe1(client *redis.Client, channel string) {
	pubsub := client.Subscribe(ctx, channel)

	// 等待订阅成功
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Fatalf("订阅失败: %v", err)
	}

	fmt.Printf("已订阅频道: %s\n", channel)

	// 监听消息
	ch := pubsub.Channel()

	// 持续接收消息
	for msg := range ch {
		fmt.Printf("subscribe1 接收到来自频道 %s 的消息: %s\n", msg.Channel, msg.Payload)
		//time.Sleep(1 * time.Second)
	}
}

func main() {
	// 创建 Redis 客户端
	client := createClient()

	// 测试连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}

	fmt.Println("成功连接到 Redis 服务器")

	// 创建一个订阅者 goroutine
	go subscribe(client, "news")

	time.Sleep(1 * time.Second)

	go subscribe1(client, "news")

	// 等待订阅者准备就绪
	// 重要性: 确保订阅者goroutine有足够时间完成订阅过程
	// 没有这个延迟，可能导致发布的消息在订阅者准备好前发送，造成消息丢失
	time.Sleep(1 * time.Second)

	// 发布几条消息
	for i := 1; i <= 5; i++ {
		message := fmt.Sprintf("这是第 %d 条新闻", i)
		if err := publish(client, "news", message); err != nil {
			log.Printf("发布消息错误: %v", err)
		}
		//time.Sleep(2 * time.Second)
	}

	fmt.Println("发布完成")

	// 等待消息处理完成
	time.Sleep(10 * time.Second)
}
