package concurrence

import (
	"fmt"
	"time"
)

// 通知其它人
func Broadcast() {
	ch := make(chan struct{})

	const P = 3

	for v := range P {
		go func() {
			<-ch
			fmt.Printf("%d 出发了\n", v)
		}()
	}

	time.Sleep(2 * time.Second)
	fmt.Println("大伙出发了")
	close(ch) // 广播
	time.Sleep(time.Second)
}

// 等其它人都完成后，我再执行
// 相当于 wait
func CutDownLatch() {
	const P = 3
	ch := make(chan struct{}, P)

	for v := range P {
		go func() {
			time.Sleep(time.Duration(v) * time.Second)
			fmt.Printf("%d 完成工作了\n", v)
			ch <- struct{}{}
		}()
	}

	for i := 0; i < P; i++ {
		<-ch
	}

	fmt.Println("其它人都执行完毕,我要开始了")
}
