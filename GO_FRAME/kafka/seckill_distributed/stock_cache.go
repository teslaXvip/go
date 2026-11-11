package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// deductLua 库存预扣的原子实现：只有 stock > 0 才减 1，保证 Redis 层不超卖。
var deductLua = redis.NewScript(`
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil then
    return -1
end
if stock <= 0 then
    return 0
end
redis.call('DECR', KEYS[1])
return 1
`)

// StockCache Redis 库存，对应图中「减1成功」环节
type StockCache struct {
	rdb *redis.Client
	key string
}

func NewStockCache(rdb *redis.Client, goodsID int64) *StockCache {
	return &StockCache{
		rdb: rdb,
		key: fmt.Sprintf("seckill:{%d}:stock", goodsID),
	}
}

// Init 把库存预热到 Redis，通常由活动开始前的定时任务或管理后台完成
func (s *StockCache) Init(ctx context.Context, stock int64) error {
	return s.rdb.Set(ctx, s.key, stock, 0).Err()
}

// Deduct 库存减 1，false 表示已售罄
func (s *StockCache) Deduct(ctx context.Context) (bool, error) {
	n, err := deductLua.Run(ctx, s.rdb, []string{s.key}).Int()
	if err != nil {
		return false, err
	}
	if n < 0 {
		return false, ErrStockNotInit
	}
	return n == 1, nil
}

// Refund 订单入队失败时回补库存，避免少卖
func (s *StockCache) Refund(ctx context.Context) error {
	return s.rdb.Incr(ctx, s.key).Err()
}

// Stock 读取当前 Redis 库存
func (s *StockCache) Stock(ctx context.Context) (int64, error) {
	return s.rdb.Get(ctx, s.key).Int64()
}
