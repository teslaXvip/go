package concurrence

import (
	"fmt"
	"math/rand"
	"time"
)

func ListenMultiWay() {
	// 创建两个带缓冲channel
	ch1 := make(chan int, 1000)
	ch2 := make(chan byte, 1000)

	// 第一个goroutine：随机间隔向ch1写入int随机数
	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			ch1 <- rand.Int()
		}
	}()

	// 第二个goroutine：随机间隔向ch2写入byte随机数
	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			ch2 <- byte(rand.Int())
		}
	}()

AB: // 外层循环标签，用于break跳出多层循环
	for {
		time.Sleep(time.Second)
		// select多路监听多个channel
		select {
		case v1 := <-ch1:
			fmt.Printf("v1=%d\n", v1)
		case v2 := <-ch2:
			fmt.Printf("v2=%d\n", v2)
			// 如果ch2读到的值小于40，跳出AB标签的for循环，结束监听
			if v2 < 40 {
				break AB
			}
		default:
			// 没有channel可读取时执行default，非阻塞
			fmt.Println("default")
		}
	}

	// 兜底无阻塞读取ch1：退出瞬间如果ch1刚写入数据，做最后打印
	// 用select+default实现不阻塞，避免goroutine挂住
	select {
	case v1 := <-ch1:
		fmt.Printf("at last v1=%d\n", v1)
	default:
	}
}
