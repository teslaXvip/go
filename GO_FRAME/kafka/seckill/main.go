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
)

const (
	defaultStock  = 100 // 初始库存
	defaultBucket = 10  // 令牌桶容量：同一时刻最多 10 个请求进入扣减环节
	defaultQueue  = 100 // 订单队列缓冲
	requestTotal  = 1000
	goodsID       = 1
)

var (
	ErrSoldOut  = errors.New("商品已售罄")
	ErrShutdown = errors.New("服务已关闭")
)

// Order 秒杀订单
type Order struct {
	ID      int64
	UserID  int64
	GoodsID int64
	Created time.Time
}

// DB 单进程内模拟数据库：库存表 + 订单表
type DB struct {
	mu     sync.Mutex
	stock  int64
	orders []Order
}

func NewDB(stock int64) *DB {
	return &DB{stock: stock}
}

// SaveOrder 对应图中的最后一步：订单落库、数据库库存减 1
func (d *DB) SaveOrder(o Order) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stock--
	d.orders = append(d.orders, o)
}

// Stats 返回数据库当前库存与订单数
func (d *DB) Stats() (stock int64, orderCount int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stock, len(d.orders)
}

// Seckill 单进程秒杀：令牌桶 channel → 内存扣减 → 订单队列 channel → 数据库
type Seckill struct {
	stock  int64         // 内存库存，预扣减
	tokens chan struct{} // 令牌桶
	orders chan Order    // 订单队列
	seq    int64         // 订单号生成器
	db     *DB
}

func NewSeckill(stock int64, db *DB, bucketSize, queueSize int) *Seckill {
	return &Seckill{
		stock:  stock,
		tokens: make(chan struct{}, bucketSize),
		orders: make(chan Order, queueSize),
		db:     db,
	}
}

// Do 处理一次秒杀请求
func (s *Seckill) Do(ctx context.Context, userID int64) error {
	// 1. 令牌桶：拿不到令牌就在 channel 上排队，限制进入扣减环节的并发数
	if !s.acquireToken(ctx) {
		return ErrShutdown
	}
	defer s.releaseToken()

	// 2. 内存库存减 1，CAS 保证不超卖
	if !s.deductStock() {
		return ErrSoldOut
	}

	order := Order{
		ID:      atomic.AddInt64(&s.seq, 1),
		UserID:  userID,
		GoodsID: goodsID,
		Created: time.Now(),
	}

	// 3. 订单写入队列，由消费者异步落库
	select {
	case s.orders <- order:
		return nil
	case <-ctx.Done():
		atomic.AddInt64(&s.stock, 1) // 入队失败，回补内存库存
		return ErrShutdown
	}
}

// Consume 消费订单队列并落库，队列关闭后返回
func (s *Seckill) Consume() {
	for order := range s.orders {
		s.db.SaveOrder(order)
	}
}

// Close 关闭订单队列，通知消费者退出
func (s *Seckill) Close() {
	close(s.orders)
}

func (s *Seckill) acquireToken(ctx context.Context) bool {
	select {
	case s.tokens <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (s *Seckill) releaseToken() {
	<-s.tokens
}

func (s *Seckill) deductStock() bool {
	for {
		cur := atomic.LoadInt64(&s.stock)
		if cur <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt64(&s.stock, cur, cur-1) {
			return true
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := NewDB(defaultStock)
	s := NewSeckill(defaultStock, db, defaultBucket, defaultQueue)

	// 消费者：把订单队列中的订单落库（数据库库存减 1）
	var consumerWG sync.WaitGroup
	consumerWG.Add(1)
	go func() {
		defer consumerWG.Done()
		s.Consume()
	}()

	var (
		wg      sync.WaitGroup
		success int64
		soldOut int64
		aborted int64
	)

	start := time.Now()
	for i := 0; i < requestTotal; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			switch err := s.Do(ctx, userID); {
			case err == nil:
				atomic.AddInt64(&success, 1)
			case errors.Is(err, ErrSoldOut):
				atomic.AddInt64(&soldOut, 1)
			default:
				atomic.AddInt64(&aborted, 1)
			}
		}(int64(i + 1))
	}
	wg.Wait()

	s.Close()
	consumerWG.Wait()

	stock, orderCount := db.Stats()
	log.Printf("耗时: %v", time.Since(start))
	log.Printf("请求总数: %d，成功: %d，售罄: %d，中断: %d", requestTotal, success, soldOut, aborted)
	log.Printf("内存剩余库存: %d，数据库剩余库存: %d，数据库订单数: %d",
		atomic.LoadInt64(&s.stock), stock, orderCount)
}
