package concurrence

import (
	"fmt"
	"math/rand"
	"time"
)

func ListenMultiWay() {
	ch1 := make(chan int, 1000)
	ch2 := make(chan byte, 1000)

	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			ch1 <- rand.Intn(1000) // 限定范围，避免打印超大数
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			ch2 <- byte(rand.Intn(256)) // 语义更清晰
		}
	}()

	done := make(chan struct{})
	for {
		select {
		case v1 := <-ch1:
			fmt.Printf("v1=%d\n", v1)
		case v2 := <-ch2:
			fmt.Printf("v2=%d\n", v2)
			if v2 < 40 {
				close(done) // 通知退出
			}
		case <-done:
			// 兜底无阻塞读取
			select {
			case v1 := <-ch1:
				fmt.Printf("at last v1=%d\n", v1)
			default:
			}
			return
		}
	}
}
