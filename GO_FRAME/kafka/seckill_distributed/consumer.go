package main

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
)

// Consumer 订单队列的消费端：单 goroutine 顺序消费，把「并行写库」变成「串行写库」，
// 落库压力被 Kafka 削峰，数据库不会被瞬时并发打垮。
type Consumer struct {
	reader    *kafka.Reader
	store     Store
	processed atomic.Int64
	failed    atomic.Int64
	done      chan struct{}
}

func NewConsumer(brokers []string, topic, groupID string, store Store) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10 << 20,
			CommitInterval: 0, // 手动提交，处理完再提交 offset
			StartOffset:    kafka.FirstOffset,
		}),
		store: store,
		done:  make(chan struct{}),
	}
}

// Run 阻塞式消费，ctx 取消后退出
func (c *Consumer) Run(ctx context.Context) {
	defer close(c.done)
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[consumer] 拉取消息失败: %v", err)
			continue
		}

		if err := processOrder(ctx, c.store, msg.Value); err != nil {
			// 库存不足这类业务失败重试也没用，记日志后照常提交，避免卡住队列
			c.failed.Add(1)
			log.Printf("[consumer] 处理订单失败: %v", err)
		} else {
			c.processed.Add(1)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil && ctx.Err() == nil {
			log.Printf("[consumer] 提交 offset 失败: %v", err)
		}
	}
}

func (c *Consumer) Wait() {
	<-c.done
}

func (c *Consumer) Processed() int64 {
	return c.processed.Load()
}

func (c *Consumer) Failed() int64 {
	return c.failed.Load()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

// processOrder 单条订单落库：数据库库存减 1 + 订单入库，同一事务内完成
func processOrder(ctx context.Context, store Store, payload []byte) error {
	o, err := decodeOrder(payload)
	if err != nil {
		return err
	}
	return store.DeductStockAndCreateOrder(ctx, o)
}
