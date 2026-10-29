package main

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "hello"

func main() {
	// 连接 rabbitmq，默认账号guest/guest
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	if err != nil {
		log.Fatalf("连接 rabbitmq 失败: %v", err)
	}
	defer conn.Close()

	// 声明队列
	declareCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("创建 channel 失败: %v", err)
	}
	_, err = declareCh.QueueDeclare(queueName, true, false, false, false, nil)
	declareCh.Close()
	if err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}

	// 发消息：并发发送时每个 goroutine 使用独立 channel，amqp.Channel 不是并发安全的
	messages := []string{"Hello RabbitMQ!", "zx", "我是消息"}

	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg string) {
			defer wg.Done()
			if err := send(conn, msg); err != nil {
				log.Printf("消息发送失败 [%s]: %v", msg, err)
			}
		}(msg)
	}
	wg.Wait()
}

func send(conn *amqp.Connection, msg string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			// 队列是 durable 的，消息也要标持久化，否则 broker 重启后消息丢失
			DeliveryMode: amqp.Persistent,
			Body:         []byte(msg),
		})
	if err != nil {
		return err
	}

	log.Printf("消息发送成功: %s", msg)
	return nil
}
