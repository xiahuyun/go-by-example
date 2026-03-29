// server.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)

	addr := conn.RemoteAddr().String()
	fmt.Println("Connected to Websocket Server:", addr)

	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	go testServerConnection(conn)

	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		mu.Lock()
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
		mu.Unlock()
	}
}

func testServerConnection(conn *websocket.Conn) {
	for {
		mu.Lock()
		if err := conn.WriteMessage(websocket.TextMessage, []byte("hello test server connection")); err != nil {
			log.Println("Error during message writing:", err)
		}
		mu.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}
