package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "hello"

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatalf("连接 rabbitmq 失败: %v", err)
	}
	defer conn.Close()

	// Ctrl+C / SIGTERM 时优雅退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := receive(ctx, conn, id); err != nil {
				log.Printf("[%d] 消费异常退出: %v", id, err)
			}
		}(i)
	}
	wg.Wait()
}

func receive(ctx context.Context, conn *amqp.Connection, flag int) error {
	// 每个消费者使用独立 channel，amqp.Channel 不是并发安全的
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 消费者先启动时队列可能还不存在，声明一次保持与生产者一致
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	// 未设置 Qos 时 broker 会把消息全部推给客户端，这里限制为每次只投递一条
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	consume, err := ch.Consume(
		queueName, "", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%d] 收到退出信号，停止消费", flag)
			return nil
		case delivery, ok := <-consume:
			if !ok {
				return nil
			}
			log.Printf("[%d] receive message [%s]", flag, delivery.Body)
			if err := delivery.Ack(false); err != nil {
				log.Printf("[%d] ack 失败: %v", flag, err)
			}
		}
	}
}
