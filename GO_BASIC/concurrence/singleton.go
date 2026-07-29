package concurrence

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
)

// config 配置结构体
type config struct {
	Password      string
	ServerAddress string
}

var (
	// 优化点2: 使用 atomic.Pointer 保证内存可见性，避免其他 goroutine 读到半初始化状态
	cfg atomic.Pointer[config]
	mu1 sync.Mutex
	// 优化点3: 初始化失败时的退避缓存，避免反复加锁+读文件
	initErr     error
	initErrTime time.Time
)

// 优化点5: 配置路径通过参数注入，避免硬编码相对路径
func GetConfig(configPath string) *config {
	// 优化点2: 使用 atomic 读取，保证跨 goroutine 可见性
	if cfg.Load() != nil {
		return cfg.Load()
	}

	mu1.Lock()
	defer mu1.Unlock()

	// 双重检查
	if cfg.Load() != nil {
		return cfg.Load()
	}

	// 优化点3: 如果最近初始化失败且未超过退避时间，直接返回错误，避免反复重试
	if initErr != nil && time.Since(initErrTime) < 5*time.Second {
		slog.Error("配置初始化近期失败，退避中", "error", initErr, "retry_after", 5*time.Second-time.Since(initErrTime))
		return nil
	}

	vp := viper.New()
	// 优化点5: 使用参数注入的路径，而非硬编码相对路径
	vp.AddConfigPath(configPath)
	vp.SetConfigName("mysql")
	vp.SetConfigType("yaml")

	// 优化点4: 使用 slog 替代 fmt.Println，支持日志级别和结构化输出
	slog.Info("解析配置文件")
	if err := vp.ReadInConfig(); err != nil {
		slog.Error("配置文件读取失败", "error", err)
		// 优化点3: 缓存初始化错误及时间，实现退避策略
		initErr = err
		initErrTime = time.Now()
		return nil
	}

	c := &config{
		Password:      vp.GetString("gift.pass"),
		ServerAddress: vp.GetString("gift.host"),
	}
	// 优化点2: 使用 atomic.Store 保证 happens-before 关系，其他 goroutine 能看到完整初始化的对象
	cfg.Store(c)
	// 优化点3: 初始化成功后清除错误缓存
	initErr = nil
	return c
}
