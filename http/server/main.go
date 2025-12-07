package main

import (
	"fmt"
	"net/http"
)

// 定义处理函数，处理根路径"/"的请求
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 向客户端写入响应内容
	fmt.Fprint(w, "Hello I am server 3")
}

func main() {
	// 创建自定义的ServeMux实例
	mux := http.NewServeMux()

	// 在mux上注册路由和处理函数
	mux.HandleFunc("/", helloHandler)

	// 启动服务器并监听12302端口，使用自定义的mux作为handler
	fmt.Println("Server is listening on port 12302...")
	if err := http.ListenAndServe(":12302", mux); err != nil {
		// 错误处理
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
