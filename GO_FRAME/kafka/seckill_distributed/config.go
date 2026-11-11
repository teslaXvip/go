package main

import (
	"os"
	"strconv"
	"strings"
)

// Config 秒杀压测参数，全部可用环境变量覆盖
type Config struct {
	RedisAddr    string   // Redis 地址，令牌桶与库存都放在这里
	KafkaBrokers []string // Kafka broker 列表
	Topic        string   // 订单队列 topic
	GroupID      string   // 消费者组
	MySQLDSN     string   // 最终落库的 MySQL

	GoodsID       int64   // 商品 ID
	Stock         int64   // 初始库存
	TokenCapacity int     // 令牌桶容量
	TokenRate     float64 // 每秒补充的令牌数
	Requests      int     // 压测并发请求数
}

func LoadConfig() Config {
	return Config{
		RedisAddr:     envStr("SECKILL_REDIS_ADDR", "127.0.0.1:6379"),
		KafkaBrokers:  strings.Split(envStr("SECKILL_KAFKA_BROKERS", "127.0.0.1:9092"), ","),
		Topic:         envStr("SECKILL_KAFKA_TOPIC", "seckill-orders"),
		GroupID:       envStr("SECKILL_KAFKA_GROUP", "seckill-order-consumer"),
		MySQLDSN:      envStr("SECKILL_MYSQL_DSN", "tester:123456@tcp(127.0.0.1:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local"),
		GoodsID:       int64(envInt("SECKILL_GOODS_ID", 1)),
		Stock:         int64(envInt("SECKILL_STOCK", 100)),
		TokenCapacity: envInt("SECKILL_TOKEN_CAPACITY", 200),
		TokenRate:     envFloat("SECKILL_TOKEN_RATE", 1000),
		Requests:      envInt("SECKILL_REQUESTS", 5000),
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
