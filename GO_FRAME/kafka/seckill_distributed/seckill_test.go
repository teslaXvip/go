package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 并发落库不超卖：100 库存面对 1000 条订单，只能成功 100 条
func TestMemoryStoreNoOversell(t *testing.T) {
	const (
		goodsID  = int64(1)
		stock    = int64(100)
		requests = 1000
	)

	store := newMemoryStore()
	if _, err := store.EnsureStock(context.Background(), goodsID, stock); err != nil {
		t.Fatalf("初始化库存失败: %v", err)
	}

	var (
		wg      sync.WaitGroup
		success int64
		refused int64
	)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			err := store.DeductStockAndCreateOrder(context.Background(), Order{
				ID:        userID,
				UserID:    userID,
				GoodsID:   goodsID,
				CreatedAt: time.Now(),
			})
			switch {
			case err == nil:
				atomic.AddInt64(&success, 1)
			case errors.Is(err, ErrOutOfStock):
				atomic.AddInt64(&refused, 1)
			default:
				t.Errorf("非预期错误: %v", err)
			}
		}(int64(i + 1))
	}
	wg.Wait()

	if success != stock {
		t.Fatalf("成功数应为 %d，实际 %d", stock, success)
	}
	if refused != requests-stock {
		t.Fatalf("库存不足数应为 %d，实际 %d", requests-stock, refused)
	}
	if left, _ := store.Stock(context.Background(), goodsID); left != 0 {
		t.Fatalf("库存应扣完，实际剩余 %d", left)
	}
	if n, _ := store.OrderCount(context.Background(), goodsID); n != stock {
		t.Fatalf("订单数应为 %d，实际 %d", stock, n)
	}
}

// EnsureStock 幂等：已有库存时不能被重置
func TestEnsureStockKeepsExisting(t *testing.T) {
	store := newMemoryStore()
	ctx := context.Background()

	if _, err := store.EnsureStock(ctx, 1, 100); err != nil {
		t.Fatal(err)
	}
	if err := store.DeductStockAndCreateOrder(ctx, Order{ID: 1, UserID: 1, GoodsID: 1}); err != nil {
		t.Fatal(err)
	}
	got, err := store.EnsureStock(ctx, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got != 99 {
		t.Fatalf("库存应保持 99，实际 %d", got)
	}
}

// 消费端单条处理：正常订单落库、库存为 0 时拒绝、脏消息返回解析错误
func TestProcessOrder(t *testing.T) {
	store := newMemoryStore()
	ctx := context.Background()
	if _, err := store.EnsureStock(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(Order{ID: 1, UserID: 7, GoodsID: 1, CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	if err := processOrder(ctx, store, payload); err != nil {
		t.Fatalf("正常订单应成功: %v", err)
	}
	if err := processOrder(ctx, store, payload); !errors.Is(err, ErrOutOfStock) {
		t.Fatalf("库存为 0 时应返回 ErrOutOfStock，实际 %v", err)
	}
	if err := processOrder(ctx, store, []byte("not-json")); err == nil {
		t.Fatal("非法消息应返回解析错误")
	}

	if n, _ := store.OrderCount(ctx, 1); n != 1 {
		t.Fatalf("订单数应为 1，实际 %d", n)
	}
}
