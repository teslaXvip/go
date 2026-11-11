package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSeckillNoOversell(t *testing.T) {
	const (
		stock    = 100
		requests = 1000
		bucket   = 8
		queue    = 16
	)

	db := NewDB(stock)
	s := NewSeckill(stock, db, bucket, queue)

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
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			err := s.Do(context.Background(), userID)
			switch {
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

	if success != stock {
		t.Fatalf("成功数应为 %d，实际 %d", stock, success)
	}
	if soldOut != requests-stock {
		t.Fatalf("售罄数应为 %d，实际 %d", requests-stock, soldOut)
	}
	if aborted != 0 {
		t.Fatalf("不应有中断请求，实际 %d", aborted)
	}
	if left := atomic.LoadInt64(&s.stock); left != 0 {
		t.Fatalf("内存库存应扣完，实际剩余 %d", left)
	}

	dbStock, orderCount := db.Stats()
	if dbStock != 0 {
		t.Fatalf("数据库库存应扣完，实际剩余 %d", dbStock)
	}
	if orderCount != stock {
		t.Fatalf("数据库订单数应为 %d，实际 %d", stock, orderCount)
	}
}

func TestSeckillSoldOut(t *testing.T) {
	s := NewSeckill(0, NewDB(0), 1, 1)
	if err := s.Do(context.Background(), 1); !errors.Is(err, ErrSoldOut) {
		t.Fatalf("库存为 0 时应返回售罄，实际 %v", err)
	}
}
