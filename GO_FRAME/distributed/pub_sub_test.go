package distributed_test

import (
	"context"
	"testing"
	"time"

	"github.com/teslaXvip/go/go_frame/distributed"

	"github.com/redis/go-redis/v9"
)

// TestPubSub 集成测试，需要本地启动redis:127.0.0.1:6379
func TestPubSub(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	ctx := context.Background()

	// ping 校验连通性
	pingCmd := client.Ping(ctx)
	if err := pingCmd.Err(); err != nil {
		t.Fatalf("redis connect failed: %v", err)
	}

	distributed.PubSub(ctx, client)

	// 等待所有goroutine完成输出
	time.Sleep(2 * time.Second)
}

// TestPublish 单独测试publish函数
func TestPublish(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, "publisher_name", "test_pub")

	distributed.Publish(ctx, client, "test_channel", "hello redis pubsub")
}

// TestSubscribe 单独测试subscribe函数
func TestSubscribe(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, "subscriber_name", "test_sub")

	go distributed.Subscribe(ctx, client, []string{"test_channel"})
	time.Sleep(500 * time.Millisecond)

	// 发一条测试消息
	distributed.Publish(baseCtx, client, "test_channel", "test message content")
	time.Sleep(1 * time.Second)
}

/*
# 执行全部测试
go test -v ./distributed  pub_sub_test.go pub_sub.go

# 只跑主流程
go test -v ./distributed -run TestPubSub

*/
