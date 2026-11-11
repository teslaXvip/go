package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

// Order 秒杀订单，通过 Kafka 传到消费端落库
type Order struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	GoodsID   int64     `json:"goods_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Producer 订单队列的生产端：把抢购成功的订单写入 Kafka
type Producer struct {
	writer  *kafka.Writer
	goodsID int64
	seq     int64
	base    int64
}

func NewProducer(brokers []string, topic string, goodsID int64) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
			// 按 key 哈希分区，同一商品固定落同一分区，保证消费顺序
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		goodsID: goodsID,
		// 订单号 = 启动时间戳 + 自增，跨进程重启也不会撞主键
		base: time.Now().UnixNano(),
	}
}

// Publish 写订单到 Kafka，key 用商品 ID
func (p *Producer) Publish(ctx context.Context, o Order) error {
	o.ID = p.base + atomic.AddInt64(&p.seq, 1)

	payload, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(o.GoodsID, 10)),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func decodeOrder(payload []byte) (Order, error) {
	var o Order
	if err := json.Unmarshal(payload, &o); err != nil {
		return Order{}, fmt.Errorf("解析订单失败: %w", err)
	}
	return o, nil
}
