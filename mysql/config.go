package main

import (
	"log"
	"os"
	"strconv"
)

// Config 存储应用程序配置
 type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	BatchSize  int
	NumRecords int
}

// LoadConfig 从环境变量加载配置
func LoadConfig() Config {
	// 从环境变量获取配置，如果不存在则使用默认值
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnvAsInt("DB_PORT", 3306)
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "testdb")
	batchSize := getEnvAsInt("BATCH_SIZE", 1000)
	numRecords := getEnvAsInt("NUM_RECORDS", 1000)

	return Config{
		DBHost:     host,
		DBPort:     port,
		DBUser:     user,
		DBPassword: password,
		DBName:     dbName,
		BatchSize:  batchSize,
		NumRecords: numRecords,
	}
}

// getEnv 从环境变量获取字符串值，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 从环境变量获取整数值，如果不存在或解析失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("解析环境变量 %s 失败: %v, 使用默认值 %d\n", key, err, defaultValue)
		return defaultValue
	}

	return value
}