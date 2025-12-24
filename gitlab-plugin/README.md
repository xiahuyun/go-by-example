# GitLab Plugin

这是一个用于访问 GitLab API 获取项目信息的 Go 程序。

## 功能

- 通过 GitLab API 在 https://gitlab-ce.alauda.cn/ 上搜索并获取名为 `hxia-test` 的项目
- 显示项目详细信息，包括 ID、名称、描述、命名空间等
- 获取第一个匹配项目的完整详细信息并保存到 JSON 文件
- 从环境变量读取 GitLab 访问令牌，确保安全性

## 前提条件

- Go 1.20 或更高版本
- GitLab 访问令牌（具有读取项目信息的权限）

## 使用方法

### 1. 设置环境变量

首先，设置 GitLab 访问令牌：

```bash
export GITLAB_TOKEN=your_personal_access_token_here
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 编译和运行

```bash
go build -o gitlab-plugin
./gitlab-plugin
```

## 项目结构

- `main.go`: 主程序文件，包含 API 调用逻辑
- `go.mod`: Go 模块定义文件
- `gitlab_project_response.json`: 运行后生成的文件，包含项目的完整响应数据

## 注意事项

- 确保访问令牌具有足够的权限（至少需要 `read_api` 范围）
- 如果 API 响应失败，请检查令牌权限和网络连接
- 程序默认搜索名称包含 `hxia-test` 的项目

## 错误处理

程序会在以下情况输出错误信息：
- 未设置 GITLAB_TOKEN 环境变量
- JSON 解码失败
- HTTP 请求发送失败
- API 响应状态码不为 200 OK
- 未找到匹配的项目

## 示例输出

成功获取项目后，将输出以下信息：

```
找到 1 个匹配的项目:
========================================
项目 1:
ID: 12345
名称: hxia-test
描述: 测试项目
命名空间: username (username)
完整路径: username/hxia-test
项目URL: https://gitlab-ce.alauda.cn/username/hxia-test
创建时间: 2024-01-01T12:00:00Z
最后活动时间: 2024-01-01T12:00:00Z
----------------------------------------

特定项目的完整响应已保存到 gitlab_project_response.json
```