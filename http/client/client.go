package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := http.Client{
		Timeout: time.Minute * 30,
		Transport: &http.Transport{
			DisableKeepAlives:     false,
			MaxIdleConns:          10,
			IdleConnTimeout:       time.Second * 10,
			ResponseHeaderTimeout: time.Minute * 2,
			ExpectContinueTimeout: time.Minute * 1,
			MaxIdleConnsPerHost:   10,
			DisableCompression:    true,
		},
	}

	// 发送GET请求到服务器的根路径"/"
	for i := 0; i < 2; i++ {
		resp, err := client.Get("http://localhost:12302/")
		if err != nil {
			// 错误处理
			fmt.Printf("Failed to send request: %v\n", err)
			return
		}

		// 读取响应体内容

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// 错误处理
			fmt.Printf("Failed to read response body: %v\n", err)
			return
		}

		// 打印服务器返回的响应内容
		fmt.Printf("Server response: %s\n", body)

		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v\n", err)
		}
	}
}
