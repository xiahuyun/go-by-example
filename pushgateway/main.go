package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

// ==================== 自定义指标定义 ====================
// 定义需要暴露的指标（示例：计数器、仪表盘）
var (
	myCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",                     // 指标名（符合 Prometheus 规范）
			Help: "Total number of HTTP requests processed", // 指标描述
			ConstLabels: prometheus.Labels{ // 固定标签（可选）
				"app": "my_exporter",
			},
		},
	)

	myGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage of the exporter process",
			ConstLabels: prometheus.Labels{
				"app": "my_exporter",
			},
		},
	)
)

// ==================== 初始化指标注册 ====================
func init() {
	// 将自定义指标注册到 Prometheus 默认注册器
	prometheus.MustRegister(myCounter, myGauge)
}

// ==================== 模拟业务逻辑（更新指标） ====================
// 示例：每隔 5 秒模拟更新指标值
func updateMetrics() {
	for {
		// 模拟 HTTP 请求计数（每 5 秒增加 2 次）
		myCounter.Add(2)

		// 模拟内存使用（随机值，实际可替换为真实采集逻辑）
		myGauge.Set(float64(randInt(100, 500))) // 假设内存使用在 100-500MB 之间

		time.Sleep(5 * time.Second)
	}
}

// 辅助函数：生成随机整数（仅示例用）
func randInt(min, max int) int {
	// 实际项目中建议使用 math/rand 包（需初始化种子）
	return min + (max-min)/2
}

// ==================== 启动 HTTP 服务（供 Prometheus Pull） ====================

func startHTTPServer() {
	// 暴露 /metrics 端点，供 Prometheus 主动拉取指标
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Exporter HTTP server started on port %d")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// ==================== 推送指标到 Pushgateway ====================
func pushToPushgateway(pushgatewayURL, jobName, instanceName string) {
	// 创建 Pusher（指定 Pushgateway 地址）
	pusher := push.New(pushgatewayURL, jobName)

	// 添加实例标签（可选，但推荐）
	// pusher = pusher.Instance(instanceName)

	// 执行推送（将当前注册器中的所有指标推送到 Pushgateway）
	err := pusher.Push()
	if err != nil {
		log.Printf("Failed to push metrics to Pushgateway: %v", err)
		return
	}
	log.Println("Successfully pushed metrics to Pushgateway")
}

// ==================== 主函数 ====================
func main() {
	// 启动 HTTP 服务（供 Prometheus Pull）
	// go startHTTPServer()

	// 启动指标更新协程（模拟业务逻辑）
	go updateMetrics()

	// 定时推送指标到 Pushgateway（示例：每 10 秒推送一次）
	pushgatewayURL := "http://localhost:9091" // Pushgateway 地址（需与实际运行一致）
	jobName := "hxia_custom_exporter"         // 任务名称（Prometheus 拉取时的标识）
	instanceName := "exporter-instance-01"    // 实例名称（区分不同部署实例）

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pushToPushgateway(pushgatewayURL, jobName, instanceName)
	}
}
