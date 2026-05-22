# go-by-example 项目目录说明

这个仓库是一个按主题拆分的 Go 示例集合，覆盖并发、存储、中间件、可观测性、Kubernetes、分布式一致性等常见场景。

## 顶层目录用途

| 目录 | 用途 | 关键内容 |
| --- | --- | --- |
| `actor` | 使用 `oklog/run` 管理 goroutine 生命周期与退出协调 | `main.go` |
| `ants-pool` | 对比原生 goroutine 与 `ants` 协程池的资源开销与吞吐 | `ants_test.go`, `main.go` |
| `asynq` | 基于 Redis 的异步任务队列示例（生产者/消费者） | `client/`, `server/`, `README.md` |
| `chunkenc` | 演示 Prometheus TSDB 的 XOR chunk 压缩与遍历 | `main.go` |
| `client-go` | Kubernetes `client-go` 用法示例：RESTClient、List、Informer、Indexer、Webhook | `RESTClient/`, `informer/`, `webhook/` |
| `context` | `context` 取消、超时等协程控制示例 | `cancel/`, `timeout/` |
| `cow` | 并发配置更新示例（Copy-On-Write 思路实验） | `main.go` |
| `envoy` | Envoy 静态/动态配置样例（YAML） | `envoy-*.yaml` |
| `es` | Elasticsearch 客户端调用示例（含 ILM 等操作） | `main.go` |
| `etcd` | etcd 相关专题示例集合（客户端、存储引擎、Raft、MVCC、线性读流程等） | `basic/`, `mvcc/`, `raft/`, `linearizable-read-demo/` 等 |
| `excel` | 通过 HTTP 导出 Excel 文件示例（`excelize`） | `main.go` |
| `exporter` | 自定义 Prometheus Exporter（暴露 `/metrics`） | `main.go` |
| `fsync` | `write + fsync` 持久化语义演示 | `fsync.go` |
| `gin` | Gin 路由与中间件示例 | `demo1/`, `demo2/`, `demo3/` |
| `gitlab-plugin` | 访问 GitLab API 查询项目并导出信息 | `main.go`, `README.md` |
| `go-zero` | go-zero 生态示例（缓存、日志、队列、短链、延迟任务等） | `cache/`, `queue/`, `shorturl/` 等 |
| `goplantuml` | Go 类型关系示例（用于 PlantUML 结构图） | `pkg/pkg.go`, `test.puml` |
| `gorm` | GORM + MySQL 的迁移与 CRUD 示例 | `demo1/`, `demo2/` |
| `grpc` | 基础 gRPC 服务端/客户端示例（含 proto） | `server/`, `client/`, `proto/` |
| `hex` | 字符串十六进制编码与截断补齐示例 | `main.go` |
| `http` | HTTP client/server、chunked 传输、watch 示例 | `client/`, `server/`, `chunked/`, `watch/` |
| `jwt` | JWT 生成、签发、解析与校验示例 | `main.go` |
| `kafka` | Kafka 生产者/消费者示例（Sarama） | `productor/`, `consumer/` |
| `ladon` | ORY Ladon 鉴权策略示例 | `main.go` |
| `lease` | Kubernetes Lease Leader Election 示例 | `main.go` |
| `limiter` | go-zero API + gRPC 限流/并发控制服务示例 | `api/`, `rpc/` |
| `loadbalancer` | 轮询与加权轮询负载均衡算法示例 | `main.go` |
| `mmap` | mmap 文件映射、随机读取性能测试、虚拟内存实验 | `main.go`, `test/`, `test1/` |
| `mongo` | MongoDB 增删改查示例（官方驱动） | `main.go` |
| `murmur` | Murmur3 哈希与分片取模示例 | `main.go` |
| `mysql` | MySQL 批量插入示例（配置化参数） | `main.go`, `config.go`, `README.md` |
| `priority-queue` | 基于 `container/heap` 的优先级队列实现示例 | `main.go` |
| `prometheus` | Prometheus/Alertmanager/Grafana 本地运行与采集示例 | `README.md`, `02/main.go` |
| `pushgateway` | 指标本地暴露 + 推送到 Pushgateway 的示例 | `main.go` |
| `redis` | Redis Pub/Sub 与 Pipeline 示例 | `pubsub/`, `pipeline/`, `README.md` |
| `ring-buffer` | `k8s.io/utils/buffer` 环形缓冲区示例 | `main.go` |
| `ristretto-cache` | Ristretto 本地缓存与 Redis 二级缓存示例 | `basic/`, `redis/`, `README.md` |
| `sync` | `sync.Cond`、`sync.Pool` 等并发原语示例 | `cond/`, `cond2/`, `pool/` |
| `task-scheduler` | 定时任务调度示例（`robfig/cron`） | `cron/main.go` |
| `time` | 定时器/时钟抽象示例（`k8s.io/utils/clock`） | `ticker/main.go` |
| `uuid` | 请求幂等设计演进示例（含 UUID request id） | `uuid1/`, `uuid2/`, `uuid3/` |
| `verify` | etcd `verify.Assert` 断言失败/快速失败示例 | `verify.go`, `README.md` |
| `wait` | 轮询与指数退避重试示例（`k8s wait`） | `backoff/`, `poll/` |
| `wal` | 简化版 WAL（Write-Ahead Log）写入与恢复示例 | `main.go` |
| `websocket` | WebSocket v1/v2 示例（回显、连接管理、广播） | `v1/`, `v2/` |
| `zookeeper` | ZooKeeper 节点 CRUD 与 Watch 事件示例 | `main.go` |

## etcd 子项目细分

| 子目录 | 用途 |
| --- | --- |
| `etcd/async-start` | 模拟 etcd 异步启动生命周期（ready/stop/error channel） |
| `etcd/basic` | etcd 客户端基础操作（Put/Get/Delete/Watch） |
| `etcd/b tree` | B-Tree 与 LLRB 的基准测试结果与脚本 |
| `etcd/b tree example` | 基于 `google/btree` 的 CRUD 例子 |
| `etcd/b+ tree` | etcd bbolt 快照可视化与 B+Tree 辅助分析工具 |
| `etcd/bbolt` | 用 bbolt 模拟 KV 存储与 CRUD |
| `etcd/etcd-bbolt` | 直接读取 etcd 底层 bbolt 数据并解析 mvccpb KV |
| `etcd/grpc` | gRPC 多节点 + round-robin 访问示例 |
| `etcd/linearizable-read-demo` | 用最小模型模拟 linearizableReadNotify/linearizableReadLoop/applyWait 协作流程 |
| `etcd/mvcc` | etcd MVCC 历史版本读取、CAS、按修订 watch 示例 |
| `etcd/raft` | 基于 `go.etcd.io/raft/v3` 的最小多节点 raft 演示 |
| `etcd/raft-wal-restart` | 带 WAL+快照落盘与重启恢复的 raft 演示 |

## 说明

- 这是示例仓库，部分目录偏实验性质，代码风格与工程化程度可能不完全一致。
- 少量目录以配置文件或素材为主（例如 `envoy`），可按需补充可运行示例。
