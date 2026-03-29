package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	Conn *websocket.Conn
	ID   string
}

// WebSocketServer 管理 WebSocket 连接
type WebSocketServer struct {
	clients    map[string]*Client
	clientsMux sync.RWMutex
	upgrader   websocket.Upgrader
}

// NewWebSocketServer 创建一个新的 WebSocket 服务端实例
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients: make(map[string]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的连接，生产环境应该限制
			},
		},
	}
}

// HandleConnection 处理新的 WebSocket 连接
func (s *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 生成客户端 ID（这里使用简单的时间戳，实际应用中应该使用更唯一的标识）
	clientID := r.RemoteAddr
	fmt.Printf("Client RemoteAddr: %s\n", conn.RemoteAddr().String())
	fmt.Printf("Client LocalAddr: %s\n", conn.LocalAddr().String())

	client := &Client{
		Conn: conn,
		ID:   clientID,
	}

	// 注册客户端
	s.clientsMux.Lock()
	s.clients[clientID] = client
	s.clientsMux.Unlock()

	log.Printf("Client connected: %s", clientID)

	// 处理消息
	go s.handleMessages(client)
}

// handleMessages 处理客户端发送的消息
func (s *WebSocketServer) handleMessages(client *Client) {
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients, client.ID)
		s.clientsMux.Unlock()
		client.Conn.Close()
		log.Printf("Client disconnected: %s", client.ID)
	}()

	for {
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from %s: %s", client.ID, message)

		// 回显消息
		if err := client.Conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

// GetClient 获取指定 ID 的客户端连接
func (s *WebSocketServer) GetClient(clientID string) *Client {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	return s.clients[clientID]
}

// GetAllClients 获取所有客户端连接
func (s *WebSocketServer) GetAllClients() map[string]*Client {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	clients := make(map[string]*Client, len(s.clients))
	for id, client := range s.clients {
		clients[id] = client
	}
	return clients
}

// SendToClient 向指定客户端发送消息
func (s *WebSocketServer) SendToClient(clientID string, message []byte) error {
	client := s.GetClient(clientID)
	if client == nil {
		return nil
	}
	return client.Conn.WriteMessage(websocket.TextMessage, message)
}

// Start 启动 WebSocket 服务
func (s *WebSocketServer) Start(addr string) error {
	http.HandleFunc("/ws", s.HandleConnection)
	log.Printf("WebSocket server started on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// main 启动服务端
func main() {
	server := NewWebSocketServer()

	// 启动一个 goroutine，定期向所有客户端发送消息
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			// 获取所有客户端连接
			clients := server.GetAllClients()
			if len(clients) > 0 {
				log.Printf("Sending message to all %d clients...", len(clients))
				for id := range clients {
					if err := server.SendToClient(id, []byte("Message from server goroutine!")); err != nil {
						log.Printf("Failed to send message to client %s: %v", id, err)
					}
				}
			}
		}
	}()

	if err := server.Start(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
