package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// WebSocketClient 表示一个 WebSocket 客户端
type WebSocketClient struct {
	Conn *websocket.Conn
}

// NewWebSocketClient 创建一个新的 WebSocket 客户端实例并连接到服务端
func NewWebSocketClient(serverAddr string) (*WebSocketClient, error) {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Client RemoteAddr: %s\n", conn.RemoteAddr().String())
	fmt.Printf("Client LocalAddr: %s\n", conn.LocalAddr().String())

	return &WebSocketClient{
		Conn: conn,
	}, nil
}

// SendMessage 发送消息到服务端
func (c *WebSocketClient) SendMessage(message []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, message)
}

// ReadMessage 从服务端读取消息
func (c *WebSocketClient) ReadMessage() ([]byte, error) {
	_, message, err := c.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Close 关闭连接
func (c *WebSocketClient) Close() error {
	return c.Conn.Close()
}

// main 启动客户端
func main() {
	// 创建客户端并连接到服务端
	client, err := NewWebSocketClient("localhost:8081")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 客户端发送消息
	if err := client.SendMessage([]byte("Hello, WebSocket!")); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// 客户端接收消息
	message, err := client.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read message: %v", err)
	}
	log.Printf("Client received: %s", message)

	// 持续读取消息
	for {
		message, err := client.ReadMessage()
		if err != nil {
			log.Printf("Failed to read message: %v", err)
			break
		}
		log.Printf("Client received: %s", message)
	}
}
