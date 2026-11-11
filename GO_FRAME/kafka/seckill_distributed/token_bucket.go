package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// tokenBucketLua 令牌桶的原子实现：
// 1. 按距上次请求的时间差以 rate 补充令牌，最多补到 capacity；
// 2. 令牌足够则扣 1 并放行，否则拒绝。
// 整个「读-补-扣」在一段 Lua 里完成，Redis 单线程执行，天然支持分布式并发。
var tokenBucketLua = redis.NewScript(`
local tokens = tonumber(redis.call('HGET', KEYS[1], 'tokens'))
local ts = tonumber(redis.call('HGET', KEYS[1], 'ts'))
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

if tokens == nil then
    tokens = capacity
    ts = now
end

-- 按毫秒时间差补充令牌
local elapsed = math.max(0, now - tonumber(ts)) / 1000.0
tokens = math.min(capacity, tokens + elapsed * rate)

if tokens < 1 then
    redis.call('HSET', KEYS[1], 'tokens', tokens, 'ts', now)
    return 0
end

redis.call('HSET', KEYS[1], 'tokens', tokens - 1, 'ts', now)
return 1
`)

// TokenBucket 分布式令牌桶，对应图中的「令牌桶」环节
type TokenBucket struct {
	rdb      *redis.Client
	key      string
	capacity int
	rate     float64
}

func NewTokenBucket(rdb *redis.Client, goodsID int64, capacity int, rate float64) *TokenBucket {
	return &TokenBucket{
		rdb:      rdb,
		key:      fmt.Sprintf("seckill:{%d}:token", goodsID),
		capacity: capacity,
		rate:     rate,
	}
}

// Allow 尝试取一个令牌，false 表示被限流
func (b *TokenBucket) Allow(ctx context.Context) (bool, error) {
	n, err := tokenBucketLua.Run(ctx, b.rdb, []string{b.key},
		b.capacity, b.rate, time.Now().UnixMilli()).Int()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
