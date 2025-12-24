# Asynq 异步任务队列示例

这是一个使用 `github.com/hibiken/asynq` 包实现的异步任务队列示例项目。

## 功能特性

- 支持多种任务类型（邮件发送、图片处理）
- 实时任务和延迟任务支持
- 任务并发处理
- Redis 作为后端存储

## 环境要求

- Go 1.18+
- Redis 服务器（默认监听 localhost:6379）

## 快速开始

### 1. 启动 Redis 服务器

确保 Redis 服务器正在运行：

```bash
redis-server
```

### 2. 启动任务服务器

在一个终端窗口中运行：

```bash
go run server.go
```

### 3. 发送任务

在另一个终端窗口中运行：

```bash
go run client.go
```

## 项目结构

- `server.go` - 任务服务器，负责处理异步任务
- `client.go` - 任务客户端，负责发送异步任务
- `go.mod` - Go 模块依赖配置
- `go.sum` - 依赖校验文件

## 任务类型

### 邮件发送任务 (`email:deliver`)
- 参数：user_id, subject, body
- 模拟发送邮件的异步处理

### 图片调整任务 (`image:resize`)
- 参数：image_url, width, height
- 模拟图片大小调整的异步处理

## 配置说明

- Redis 地址：默认 `localhost:6379`
- 并发处理数：默认 10

## 扩展开发

1. 添加新的任务类型常量
2. 实现对应的任务处理函数
3. 在服务器的任务处理器映射中注册新函数
4. 在客户端中创建并发送新类型的任务

## 依赖包

- github.com/hibiken/asynq v0.25.1
