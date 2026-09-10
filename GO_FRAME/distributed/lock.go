package distributed

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func TryLock(rc *redis.Client, key string, expire time.Duration) bool {
	cmd := rc.SetNX(context.Background(), key, "value随意", expire)
	if cmd.Err() != nil {
		return false
	} else {
		// 没有错误，返回redis SetNX真实返回值 true/false
		return cmd.Val()
		// 1. key 不存在 → 设置成功，Redis 返回 `1` → `cmd.Val()` 返回 `true`，拿到锁
		// 2. key 已经存在 → 设置失败，Redis 返回 `0` → `cmd.Val()` 返回 `false`，抢锁失败
	}
}

func ReleaseLock(rc *redis.Client, key string) {
	rc.Del(context.Background(), key)
}
