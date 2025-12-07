package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 基准测试：服务端通过HTTP头关闭连接而客户端保持长连接的场景
func BenchmarkServerClosesConnection(b *testing.B) {
	b.ReportAllocs()
	b.SetParallelism(1) // 串行执行

	// 创建一个每次请求后关闭连接的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 主动关闭连接
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	defer server.Close()

	// 客户端配置为尝试保持长连接
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false, // 客户端尝试保持长连接
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	b.ResetTimer()
	maxIterations := 50
	for i := 0; i < b.N && i < maxIterations; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()

		// 短暂暂停
		time.Sleep(50 * time.Millisecond)
	}
}

// 启动测试服务器，模拟真实HTTP服务
func startTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
}

// 基准测试：使用连接池的HTTP客户端
func BenchmarkWithConnectionPool(b *testing.B) {
	// 为了避免端口耗尽，我们使用非常保守的设置
	b.ReportAllocs()
	b.SetParallelism(1) // 串行执行

	server := startTestServer()

	defer server.Close()

	// 简单的连接池配置
	transport := &http.Transport{
		MaxIdleConns:        5, // 非常保守的连接数
		MaxIdleConnsPerHost: 2, // 每主机最多2个连接
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	b.ResetTimer()
	// 限制最大迭代次数，避免端口耗尽
	maxIterations := 100
	for i := 0; i < b.N && i < maxIterations; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()

		// 每次请求后都暂停，给系统时间释放资源
		time.Sleep(20 * time.Millisecond)
	}
}

// 基准测试：不使用连接池的HTTP客户端（禁用连接复用）
func BenchmarkWithoutConnectionPool(b *testing.B) {
	// 为了避免端口耗尽，使用极其保守的设置
	b.ReportAllocs()
	b.SetParallelism(1) // 串行执行

	server := startTestServer()

	defer server.Close()

	// 禁用连接复用
	transport := &http.Transport{
		DisableKeepAlives: true, // 禁用连接复用
		MaxIdleConns:      1,    // 最小连接数
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second, // 更长的超时时间
	}

	b.ResetTimer()
	// 限制最大迭代次数，避免端口耗尽
	maxIterations := 20
	for i := 0; i < b.N && i < maxIterations; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()

		// 更长的暂停时间，确保系统完全释放端口
		time.Sleep(100 * time.Millisecond)
	}
}
