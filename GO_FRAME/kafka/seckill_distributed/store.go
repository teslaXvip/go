package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	ErrOutOfStock     = errors.New("数据库库存不足")
	ErrStockNotInit   = errors.New("Redis 库存未初始化")
	ErrRateLimited    = errors.New("被令牌桶限流")
	ErrRedisSoldOut   = errors.New("Redis 库存已售罄")
	ErrOrderDuplicate = errors.New("订单已存在")
)

// Store 最终落库，对应图里最后一步「库存减1 → 数据库」
type Store interface {
	// DeductStockAndCreateOrder 数据库库存减 1 并写入订单，二者同一事务
	DeductStockAndCreateOrder(ctx context.Context, o Order) error
	// Stock 返回数据库当前库存
	Stock(ctx context.Context, goodsID int64) (int64, error)
	// OrderCount 返回数据库订单总数
	OrderCount(ctx context.Context, goodsID int64) (int64, error)
	// EnsureStock 商品库存记录不存在时才初始化，避免每次运行重置线上库存
	EnsureStock(ctx context.Context, goodsID, stock int64) (int64, error)
	Close() error
}

type seckillStock struct {
	GoodsID int64 `gorm:"primaryKey;autoIncrement:false"`
	Stock   int64
}

func (seckillStock) TableName() string { return "seckill_stock" }

type seckillOrder struct {
	ID        int64 `gorm:"primaryKey;autoIncrement:false"`
	UserID    int64
	GoodsID   int64
	CreatedAt time.Time
}

func (seckillOrder) TableName() string { return "seckill_order" }

type mysqlStore struct {
	db *gorm.DB
}

func newMySQLStore(dsn string) (Store, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&seckillStock{}, &seckillOrder{}); err != nil {
		return nil, err
	}
	return &mysqlStore{db: db}, nil
}

func (s *mysqlStore) DeductStockAndCreateOrder(ctx context.Context, o Order) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 带 stock > 0 条件的更新：即使消费端有多实例，数据库层也不会扣成负数
		res := tx.Model(&seckillStock{}).
			Where("goods_id = ? AND stock > 0", o.GoodsID).
			Update("stock", gorm.Expr("stock - 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrOutOfStock
		}
		return tx.Create(&seckillOrder{
			ID:        o.ID,
			UserID:    o.UserID,
			GoodsID:   o.GoodsID,
			CreatedAt: o.CreatedAt,
		}).Error
	})
}

func (s *mysqlStore) Stock(ctx context.Context, goodsID int64) (int64, error) {
	var row seckillStock
	err := s.db.WithContext(ctx).Where("goods_id = ?", goodsID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return row.Stock, err
}

func (s *mysqlStore) OrderCount(ctx context.Context, goodsID int64) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&seckillOrder{}).
		Where("goods_id = ?", goodsID).Count(&n).Error
	return n, err
}

func (s *mysqlStore) EnsureStock(ctx context.Context, goodsID, stock int64) (int64, error) {
	row := seckillStock{GoodsID: goodsID, Stock: stock}
	// 主键冲突说明已有记录，保留数据库现有库存
	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
	if err != nil {
		return 0, err
	}
	return s.Stock(ctx, goodsID)
}

func (s *mysqlStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// memoryStore 没有 MySQL 时的本地演示实现，语义与 mysqlStore 一致
type memoryStore struct {
	mu     sync.Mutex
	stock  map[int64]int64
	orders []Order
}

func newMemoryStore() *memoryStore {
	return &memoryStore{stock: make(map[int64]int64)}
}

func (s *memoryStore) DeductStockAndCreateOrder(_ context.Context, o Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stock[o.GoodsID] <= 0 {
		return ErrOutOfStock
	}
	s.stock[o.GoodsID]--
	s.orders = append(s.orders, o)
	return nil
}

func (s *memoryStore) Stock(_ context.Context, goodsID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stock[goodsID], nil
}

func (s *memoryStore) OrderCount(_ context.Context, goodsID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int64
	for _, o := range s.orders {
		if o.GoodsID == goodsID {
			n++
		}
	}
	return n, nil
}

func (s *memoryStore) EnsureStock(_ context.Context, goodsID, stock int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.stock[goodsID]; !ok {
		s.stock[goodsID] = stock
	}
	return s.stock[goodsID], nil
}

func (s *memoryStore) Close() error { return nil }
