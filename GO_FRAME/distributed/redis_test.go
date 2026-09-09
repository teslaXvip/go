package distributed_test

import (
	"context"
	"go_frame/distributed"
	"log/slog"
	"testing"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Username: "",
		Password: "",
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("connect to redis failed", "error", err)
	} else {
		slog.Info("connect to redis")
	}
}

func TestStringValue(t *testing.T) {
	distributed.StringValue(context.Background(), client)
}

// TestStructValue 测试结构体写入Redis和读取
func TestStructValue(t *testing.T) {
	// 初始化redis客户端，根据你的实际连接信息修改
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	// 先测试连通性
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("redis connect failed: %v", err)
	}

	// 构造student对象
	stu := &distributed.Student{Id: 1, Name: "大乔乔"}
	// 写入redis
	err = distributed.WriteStudent2Redis(client, stu)
	if err != nil {
		t.Fatalf("WriteStudent2Redis failed: %v", err)
	}

	// 从redis读取
	stu2 := distributed.GetStudentFromRedis(client, 1)
	if stu2 == nil {
		t.Fatal("GetStudentFromRedis return nil")
	}

	// 断言id相等
	if stu2.Id != stu.Id {
		t.Fail()
	}
	// 额外断言name，校验完整数据
	if stu2.Name != stu.Name {
		t.Fail()
	}
}
