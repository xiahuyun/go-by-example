package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// 定义任务类型常量
const (
	TypeEmailDelivery = "email:deliver"
	TypeImageResize   = "image:resize"
)

// 邮件任务负载结构
type EmailDeliveryPayload struct {
	UserID  int64  `json:"user_id"`
	Subject string `json:"subject"`
}

// 图片调整任务负载结构
type ImageResizePayload struct {
	ImageURL string `json:"image_url"`
}

// 任务处理函数
func handleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	// 从任务中获取参数
	var payload EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("[任务处理] 发送邮件给用户 %d: %s", payload.UserID, payload.Subject)
	// 模拟发送邮件的延迟
	time.Sleep(2 * time.Second)
	log.Printf("[任务处理] 邮件发送成功给用户 %d", payload.UserID)
	return nil
}

func handleImageResizeTask(ctx context.Context, t *asynq.Task) error {
	// 从任务中获取参数
	var payload ImageResizePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("[任务处理] 调整图片大小: %s", payload.ImageURL)
	// 模拟图片处理的延迟
	time.Sleep(3 * time.Second)
	log.Printf("[任务处理] 图片调整完成: %s", payload.ImageURL)
	return nil
}

func main() {
	// 配置Redis连接
	redisAddr := asynq.RedisClientOpt{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "Troy@0403",      // Redis密码
	}

	// 创建任务服务器
	srv := asynq.NewServer(
		redisAddr,
		asynq.Config{
			Concurrency: 10, // 并发处理任务数
		},
	)

	// 创建任务处理器映射
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailDelivery, handleEmailDeliveryTask)
	mux.HandleFunc(TypeImageResize, handleImageResizeTask)

	// 启动任务服务器
	log.Println("[服务器] 启动Asynq任务服务器...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("[服务器] 启动失败: %v", err)
	}
}
