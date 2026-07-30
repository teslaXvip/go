package concurrence

import (
	"context"
	"fmt"
	"time"
)

// 自定义类型，用于context.WithValue的key/value（避免key冲突）
// 官方建议不要用裸类型
type StringKey string
type StringValue string
type IntValue int

// InheritTimeout 演示子Context继承父Context超时，以更早到期为准
func InheritTimeout() {
	// 父context：1000ms超时
	parent, cancel1 := context.WithTimeout(context.Background(), time.Millisecond*1000)
	t0 := time.Now()
	defer cancel1()

	// 先休眠500ms，父context剩余有效期只剩500ms
	time.Sleep(500 * time.Millisecond)

	// 子context设置1000ms超时，但受父限制，最多只能再存活500ms
	child, cancel2 := context.WithTimeout(parent, time.Millisecond*1000)
	// 注释版本：子设置100ms超时，则100ms就到期
	// child, cancel2 := context.WithTimeout(parent, time.Millisecond*100)

	t1 := time.Now()
	defer cancel2()

	// 阻塞等待context结束
	<-child.Done()

	// 打印耗时
	fmt.Println(time.Since(t0).Milliseconds(), time.Since(t1).Milliseconds())
	fmt.Println(child.Err()) // 输出：context deadline exceeded
}

// RoutineID 演示 context.WithValue 传递协程元数据
func RoutineID() {
	for i := 0; i < 3; i++ {
		// 逐层包装携带value的context
		ctx := context.WithValue(context.Background(), StringKey("gid"), IntValue(i))
		ctx = context.WithValue(ctx, StringKey("owner"), StringValue("dqq"))

		// 启动goroutine，传递ctx
		go func(ctx context.Context) {
			// 类型断言取出gid
			if gid, ok := ctx.Value(StringKey("gid")).(IntValue); ok {
				fmt.Printf("本协程ID %d\n", gid)
			}
			// 类型断言取出owner
			if owner, ok := ctx.Value(StringKey("owner")).(StringValue); ok {
				fmt.Printf("owner %s\n", owner)
			}
		}(ctx)
	}
	// 等待goroutine执行完成
	time.Sleep(time.Second)
}
