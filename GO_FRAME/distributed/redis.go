package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func StringValue(ctx context.Context, client *redis.Client) {
	key := "name"
	value := "张三"

	defer client.Del(ctx, key)

	err := client.Set(ctx, key, value, 0).Err()
	CheckError(err)

	client.Expire(ctx, key, 3*time.Second)
	time.Sleep(2 * time.Second)

	v2, err := client.Get(ctx, key).Result()
	CheckError(err)
	fmt.Println(v2)
}

func DeleteKey(ctx context.Context, client *redis.Client) {
	n, err := client.Del(ctx, "not_exissts").Result()
	if err == nil {
		fmt.Printf("删除%d个key\n", n)
	}
}

type Student struct {
	Id   int
	Name string
}

func WriteStudent2Redis(client *redis.Client, stu *Student) error {
	if stu == nil {
		return nil
	}
	key := "STU_" + strconv.Itoa(stu.Id)
	v, err := json.Marshal(stu)
	if err != nil {
		return err
	}
	err = client.Set(context.Background(), key, v, 5*time.Minute).Err()
	return err
}

func GetStudentFromRedis(client *redis.Client, sid int) *Student {
	key := "STU_" + strconv.Itoa(sid)
	v, err := client.Get(context.Background(), key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Println(err)
		}
		return nil
	}
	var stu Student
	err = json.Unmarshal([]byte(v), &stu)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &stu
}
