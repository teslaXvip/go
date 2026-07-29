package concurrence

import (
	"context"
	"sync"
)

type GoroutineLimit struct {
	ch chan struct{}
	wg sync.WaitGroup
}

func NewGoroutineLimit(n int) *GoroutineLimit {
	return &GoroutineLimit{
		ch: make(chan struct{}, n),
	}
}

func (g *GoroutineLimit) Run(ctx context.Context, f func()) error {
	select {
	case g.ch <- struct{}{}:
		/*
			- 发送操作 ：向 g.ch （缓冲 channel）写入一个空结构体
			- g.ch 容量为 n ，当缓冲区 未满 时，写入立即成功，进入此分支
			- 意义： 拿到了一个协程槽位 ，可以启动新协程
		*/
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			f()
			<-g.ch
		}()
		return nil
	case <-ctx.Done():
		/*
			- 接收操作 ：从 ctx.Done() 读取
			- ctx.Done() 在 context 被 取消或超时 时才会返回值，之前一直阻塞
			- 意义： 外部通知放弃等待 ，不再提交任务
		*/
		return ctx.Err()
	}
}

func (g *GoroutineLimit) Wait() {
	g.wg.Wait()
}

/*
调用方示例：
limiter := NewGoroutineLimit(100)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

for range 10000 {
    if err := limiter.Run(ctx, work); err != nil {
        break // 超时或取消
    }
}
limiter.Wait() // 可靠等待所有任务完成
*/
