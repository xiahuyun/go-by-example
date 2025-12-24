package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// 定义任务类型常量
const (
	TypeEmailDelivery = "email:deliver"
	TypeImageResize   = "image:resize"
)

func main() {
	// 配置Redis连接
	redisAddr := asynq.RedisClientOpt{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "Troy@0403",      // Redis密码
	}

	// 创建任务客户端
	client := asynq.NewClient(redisAddr)
	defer client.Close()

	log.Println("[客户端] 开始发送任务...")

	// 1. 创建并发送邮件任务
	log.Println("[客户端] 发送邮件任务...")
	emailPayload := map[string]interface{}{
		"user_id": 12345,
		"subject": "欢迎使用Asynq任务队列",
	}
	emailPayloadBytes, err := json.Marshal(emailPayload)
	if err != nil {
		log.Fatalf("[客户端] 序列化邮件任务负载失败: %v", err)
	}

	emailTask := asynq.NewTask(TypeEmailDelivery, emailPayloadBytes)

	// 发送任务到默认队列
	info, err := client.Enqueue(emailTask)
	if err != nil {
		log.Fatalf("[客户端] 发送邮件任务失败: %v", err)
	}
	log.Printf("[客户端] 邮件任务发送成功: %s", info.ID)

	// 2. 创建并发送图片调整任务
	log.Println("[客户端] 发送图片调整任务...")
	imagePayload := map[string]interface{}{
		"image_url": "https://example.com/image.jpg",
	}
	imagePayloadBytes, err := json.Marshal(imagePayload)
	if err != nil {
		log.Fatalf("[客户端] 序列化图片任务负载失败: %v", err)
	}

	imageTask := asynq.NewTask(TypeImageResize, imagePayloadBytes)

	// 发送任务到默认队列
	info, err = client.Enqueue(imageTask)
	if err != nil {
		log.Fatalf("[客户端] 发送图片任务失败: %v", err)
	}
	log.Printf("[客户端] 图片任务发送成功: %s", info.ID)

	// 3. 创建并发送延迟任务 (5秒后执行)
	log.Println("[客户端] 发送延迟邮件任务...")
	delayEmailPayload := map[string]interface{}{
		"user_id": 67890,
		"subject": "这是一封延迟邮件",
	}
	delayEmailPayloadBytes, err := json.Marshal(delayEmailPayload)
	if err != nil {
		log.Fatalf("[客户端] 序列化延迟邮件任务负载失败: %v", err)
	}

	delayEmailTask := asynq.NewTask(TypeEmailDelivery, delayEmailPayloadBytes)

	// 发送任务，5秒后执行
	info, err = client.Enqueue(delayEmailTask, asynq.ProcessIn(10*time.Second))
	if err != nil {
		log.Fatalf("[客户端] 发送延迟邮件任务失败: %v", err)
	}
	log.Printf("[客户端] 延迟邮件任务发送成功: %s", info.ID)

	log.Println("[客户端] 所有任务发送完成！")
}
