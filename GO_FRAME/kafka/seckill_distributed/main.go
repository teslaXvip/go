package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// App 秒杀主链路：令牌桶 → Redis 库存减 1 → Kafka 订单队列 → 异步落库
type App struct {
	tokens   *TokenBucket
	stock    *StockCache
	producer *Producer
}

// Do 处理一次秒杀请求
func (a *App) Do(ctx context.Context, userID int64) error {
	// 1. Redis 令牌桶：原子 LIMIT，挡住绝大部分无效流量
	ok, err := a.tokens.Allow(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRateLimited
	}

	// 2. Redis 库存减 1：Lua 原子执行，不会超卖
	ok, err = a.stock.Deduct(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRedisSoldOut
	}

	// 3. 订单写入 Kafka，后面的落库交给消费端串行处理
	order := Order{
		UserID:    userID,
		GoodsID:   a.producer.goodsID,
		CreatedAt: time.Now(),
	}
	if err := a.producer.Publish(ctx, order); err != nil {
		// 入队失败就把预扣的库存还回去，避免少卖
		if refundErr := a.stock.Refund(ctx); refundErr != nil {
			log.Printf("[seckill] 回补 Redis 库存失败: %v", refundErr)
		}
		return err
	}
	return nil
}

func main() {
	cfg := LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("连接 Redis(%s) 失败: %v", cfg.RedisAddr, err)
	}

	store := openStore(cfg)
	defer store.Close()

	// 数据库库存为准，预热到 Redis；Redis 只做拦截，数据库才是最终账本
	dbStock, err := store.EnsureStock(ctx, cfg.GoodsID, cfg.Stock)
	if err != nil {
		log.Fatalf("初始化数据库库存失败: %v", err)
	}
	stock := NewStockCache(rdb, cfg.GoodsID)
	if err := stock.Init(ctx, dbStock); err != nil {
		log.Fatalf("预热 Redis 库存失败: %v", err)
	}

	producer := NewProducer(cfg.KafkaBrokers, cfg.Topic, cfg.GoodsID)
	defer producer.Close()

	consumer := NewConsumer(cfg.KafkaBrokers, cfg.Topic, cfg.GroupID, store)
	go consumer.Run(ctx)
	defer consumer.Close()

	app := &App{
		tokens:   NewTokenBucket(rdb, cfg.GoodsID, cfg.TokenCapacity, cfg.TokenRate),
		stock:    stock,
		producer: producer,
	}

	var (
		wg        sync.WaitGroup
		limited   int64
		soldOut   int64
		publishKo int64
		accepted  int64
	)

	start := time.Now()
	for i := 0; i < cfg.Requests; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			switch err := app.Do(ctx, userID); {
			case err == nil:
				atomic.AddInt64(&accepted, 1)
			case errors.Is(err, ErrRateLimited):
				atomic.AddInt64(&limited, 1)
			case errors.Is(err, ErrRedisSoldOut):
				atomic.AddInt64(&soldOut, 1)
			default:
				atomic.AddInt64(&publishKo, 1)
				log.Printf("[seckill] user=%d 下单失败: %v", userID, err)
			}
		}(int64(i + 1))
	}
	wg.Wait()
	cost := time.Since(start)

	// 等 Kafka 消费端把队列消化完再统计
	if !waitConsumed(ctx, consumer, accepted, 20*time.Second) {
		log.Printf("[seckill] 等待消费完成超时，已消费 %d 条", consumer.Processed())
	}

	redisStock, _ := stock.Stock(ctx)
	finalStock, _ := store.Stock(ctx, cfg.GoodsID)
	totalOrders, _ := store.OrderCount(ctx, cfg.GoodsID)

	log.Printf("抢购耗时: %v", cost)
	log.Printf("请求总数: %d，入队成功: %d，被限流: %d，Redis 售罄: %d，入队失败: %d",
		cfg.Requests, accepted, limited, soldOut, publishKo)
	log.Printf("消费成功: %d，消费失败: %d", consumer.Processed(), consumer.Failed())
	log.Printf("Redis 剩余库存: %d，数据库剩余库存: %d，数据库订单总数: %d",
		redisStock, finalStock, totalOrders)
}

// waitConsumed 轮询等待消费端追上入队数量
func waitConsumed(ctx context.Context, c *Consumer, accepted int64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.Processed()+c.Failed() >= accepted {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(50 * time.Millisecond):
		}
	}
	return false
}

// openStore 优先用 MySQL，连不上则退回内存实现，保证本机可先跑通链路
func openStore(cfg Config) Store {
	store, err := newMySQLStore(cfg.MySQLDSN)
	if err != nil {
		log.Printf("[seckill] 连接 MySQL 失败，改用内存存储（仅演示）: %v", err)
		return newMemoryStore()
	}
	return store
}
