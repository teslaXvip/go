package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "hello"

// 优雅关闭窗口期：收到停止信号后，最多给 3 秒处理存量消息
const shutdownGracePeriod = 3 * time.Second

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatalf("连接 rabbitmq 失败: %v", err)
	}
	defer conn.Close()

	// 捕获终止信号
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

	// 阻塞等待信号
	<-ctx.Done()
	log.Println("收到停止信号，开始优雅关闭，预留", shutdownGracePeriod)

	// 构造带超时的关闭上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer shutdownCancel()

	// 开协程等消费协程全部完成
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Println("所有消费协程正常完成，优雅退出成功")
	case <-shutdownCtx.Done():
		log.Println("优雅关闭超时！还有消费协程未完成，强制退出")
	}

	log.Println("main 退出")
}

func receive(ctx context.Context, conn *amqp.Connection, flag int) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() {
		if err := ch.Close(); err != nil {
			log.Printf("[%d] channel 关闭异常: %v", flag, err)
		}
	}()

	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}
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
			log.Printf("[%d] 收到退出信号，停止消费循环", flag)
			return nil
		case delivery, ok := <-consume:
			if !ok {
				log.Printf("[%d] consume 通道被关闭", flag)
				return nil
			}
			log.Printf("[%d] receive message [%s]", flag, delivery.Body)

			// 这里可以把业务处理逻辑写在这里
			// 如果业务处理很耗时，也建议业务内部监听 ctx.Done()
			if err := delivery.Ack(false); err != nil {
				log.Printf("[%d] ack 失败: %v", flag, err)
			}
		}
	}
}
