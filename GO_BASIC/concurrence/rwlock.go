package concurrence

import (
	"fmt"
	"sync"
	"time"
)

var (
	mu sync.RWMutex
)

// 读锁可重入：同一个goroutine可以多次获取同一把读锁
func ReentrantRLock(n int) {
	mu.RLock()
	defer mu.RUnlock()
	fmt.Println(n)
	if n > 0 {
		ReentrantRLock(n - 1)
	}
	time.Sleep(1 * time.Second)
}

// 写锁不可重入，递归调用会死锁
func ReentrantWLock(n int) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Println(n)
	if n > 0 {
		ReentrantWLock(n - 1)
	}
	time.Sleep(1 * time.Second)
}

// 获取写锁时，其他协程不能获取读锁、也不能获取写锁
func WLockExclusion() {
	mu.Lock() // 获取写锁
	defer mu.Unlock()
	go func() {
		mu.RLock()
		defer mu.RUnlock()
		fmt.Println("子协程也获得了读锁")
	}()

	go func() {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("子协程也获得了写锁")
	}()
	time.Sleep(1 * time.Second)
}

// 获取读锁时：允许其他协程获取读锁；阻塞其他协程获取写锁
func RLockExclusion() {
	mu.RLock() // 获取读锁
	defer mu.RUnlock()
	go func() {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("子协程也获得了写锁")
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		mu.RLock()
		defer mu.RUnlock()
		fmt.Println("子协程也获得了读锁")
	}()
	time.Sleep(1 * time.Second)
}

/*
# 核心知识点总结（sync.RWMutex 读写锁）

## 1. 基础互斥规则（重中之重）

`sync.RWMutex` 分为 **读锁 (RLock/RUnlock)** 和 **写锁 (Lock/Unlock)**

1. ✅ **读读共享**：多个 goroutine 可以同时持有读锁
2. ❌ **读写互斥**：持有读锁时，其他 goroutine 拿不到写锁；持有写锁时，其他 goroutine 拿不到读锁
3. ❌ **写写互斥**：同一时间只能有一个 goroutine 持有写锁

>
> 通俗理解：读可以并发；写必须独占；读和写互相排斥。

## 2. 可重入特性（容易踩坑）

1. **读锁 RLock：同一 goroutine 支持可重入**
同一个协程递归多次调用 `RLock()` 不会死锁，但是**必须调用同等次数 `RUnlock()` 释放**。

>
> ⚠️ 注意：同一个 goroutine**先拿读锁，再尝试拿写锁，会死锁！**
> 原因：写锁需要等待所有读锁释放，而当前协程自己持有读锁，永久等待。

2. **写锁 Lock：不可重入**
同一个 goroutine 递归多次调用 `Lock()` **直接死锁**（示例`ReentrantWLock`运行卡死）。

## 3. 两个演示函数运行现象

### ① `WLockExclusion()`

主线程拿到**写锁**后启动两个子协程：

- 子协程尝试 `mu.RLock()`（读锁）→ **阻塞**
- 子协程尝试 `mu.Lock()`（写锁）→ **阻塞**
输出不会打印任何子协程日志，主线程 Sleep 结束释放写锁后，子协程才能继续执行。

### ② `RLockExclusion()`

主线程拿到**读锁**后启动两个子协程：

- 子协程尝试 `mu.Lock()`（写锁）→ **阻塞**
- 子协程尝试 `mu.RLock()`（读锁）→ **成功获取**，正常打印日志

## 4. 使用场景与注意事项

✅ **适用场景：读多写少**（大量并发读取、少量修改）；读多写少时性能远优于普通`sync.Mutex`
❌ **写多读少不推荐**：大量写竞争场景，读写锁优势不明显

⚠️ 避坑清单

1. 不要在读锁范围内尝试获取写锁 → 死锁
2. 写锁不可递归重入
3. 锁、解锁成对出现，建议统一用`defer`释放锁
4. 不要跨 goroutine 传递锁（一个协程加锁、另一个协程解锁，会 panic）

## 5. 和 sync.Mutex 的简单对比

- `sync.Mutex`：完全互斥，不管读写，同一时间只能一个协程执行
- `sync.RWMutex`：区分读写，读读并发，读写 / 写写互斥

如果你需要，我可以把每个测试函数运行后的输出结果和死锁原因单独拆解说明。
*/
