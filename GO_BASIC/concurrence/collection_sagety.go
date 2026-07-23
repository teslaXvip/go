package concurrence

import (
	"fmt"
	"sync"
)

/*
数组、slice、struct允许并发修改(可能会脏写)，并发修改map可能会(不是一定会)发生fatal error,
recover()只能捕获panic，但不能捕获fatal error。
如果需要并发修改map请使用sync.Map
*/

var m = sync.Map{}

func CollectionSafety() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		m.Store("k1", "v1")
	}()

	go func() {
		defer wg.Done()
		m.Store("k1", "v2")
	}()

	wg.Wait()

	if v, exists := m.Load("k1"); exists {
		fmt.Println(v)
	}
}
