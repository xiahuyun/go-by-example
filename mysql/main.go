package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 定义一个数据结构用于表示test表中的记录
 type TestData struct {
	ID    int
	Name  string
	Age   int
	Score float64
}

func main() {
	// 加载配置
	config := LoadConfig()

	// 数据库连接信息
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("成功连接到数据库")

	// 创建test表（如果不存在）
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS test (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		age INT NOT NULL,
		score FLOAT NOT NULL
	)
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("表test创建成功或已存在")

	// 生成测试数据
	data := generateTestData(config.NumRecords)

	// 执行批量插入
	startTime := time.Now()
	err = batchInsert(db, data, config.BatchSize)
	if err != nil {
		log.Fatalf("批量插入失败: %v", err)
	}
	elapsedTime := time.Since(startTime)

	fmt.Printf("成功插入 %d 条记录，耗时: %v\n", config.NumRecords, elapsedTime)
}

// 生成测试数据
func generateTestData(count int) []TestData {
	data := make([]TestData, count)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		data[i] = TestData{
			ID:    i + 1,
			Name:  fmt.Sprintf("User%d", i+1),
			Age:   r.Intn(50) + 18, // 18-67岁
			Score: r.Float64() * 100, // 0-100分
		}
	}

	return data
}

// 批量插入数据，支持事务和真正的批量插入
func batchInsert(db *sql.DB, data []TestData, batchSize int) error {
	// 准备事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 分批次处理
	for i := 0; i < len(data); i += batchSize {
		// 计算当前批次的结束索引
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		// 获取当前批次的数据
		batch := data[i:end]

		// 构建批量插入语句
		query := "INSERT INTO test (name, age, score) VALUES "
		params := make([]interface{}, 0, len(batch)*3)

		for j, record := range batch {
			if j > 0 {
				query += ", "
			}
			query += "(?, ?, ?)"
			params = append(params, record.Name, record.Age, record.Score)
		}

		// 执行批量插入
		_, err = tx.Exec(query, params...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("执行批量插入失败: %v\nSQL: %s", err, query)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}