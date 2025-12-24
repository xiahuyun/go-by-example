# Go MySQL 批量插入示例

这个项目演示了如何使用Go语言批量向MySQL数据库插入数据。通过配置文件和环境变量，你可以轻松调整数据库连接信息和插入参数。

## 功能特点

- 从环境变量加载数据库配置
- 支持事务和真正的批量插入（使用MySQL多行插入语法）
- 可配置的批量大小和记录数量
- 自动创建测试表（如果不存在）
- 生成随机测试数据

## 目录结构

```
.\n├── main.go       // 主程序入口
├── config.go     // 配置加载
├── go.mod        // Go模块定义
├── go.sum        // 依赖包列表
└── README.md     // 项目说明
```

## 配置说明

可以通过环境变量配置以下参数：

- `DB_HOST`: 数据库主机地址 (默认: 127.0.0.1)
- `DB_PORT`: 数据库端口 (默认: 3306)
- `DB_USER`: 数据库用户名 (默认: root)
- `DB_PASSWORD`: 数据库密码 (默认: password)
- `DB_NAME`: 数据库名称 (默认: testdb)
- `BATCH_SIZE`: 每批插入的记录数量 (默认: 1000)
- `NUM_RECORDS`: 要插入的总记录数量 (默认: 1000)

## 使用方法

1. 确保已安装Go和MySQL

2. 克隆或下载此项目

3. 安装依赖

```bash
cd go-by-example/mysql
```

4. 配置环境变量（可选）

```bash
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=testdb
export BATCH_SIZE=1000
export NUM_RECORDS=10000
```

5. 运行程序

```bash
go run .
```

## 性能说明

- 使用事务可以显著提高插入性能
- 批量插入的大小（BATCH_SIZE）会影响性能，过大或过小都可能导致性能下降
- 建议根据实际的数据库配置和网络环境调整BATCH_SIZE参数

## 注意事项

1. 确保MySQL服务已启动并可以访问
2. 确保数据库用户具有创建表和插入数据的权限
3. 对于大量数据插入，可能需要调整MySQL的配置参数，如`max_allowed_packet`
4. 本程序仅用于演示目的，生产环境中请根据实际需求进行修改和优化

## 许可证

本项目采用MIT许可证 - 详情请见LICENSE文件