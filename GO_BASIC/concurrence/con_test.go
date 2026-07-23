package concurrence_test

import (
	"sync"
	"testing"

	"go_basic/concurrence"
)

func TestCollectionSafety(t *testing.T) {

}

func TestConcurrentMap(t *testing.T) {
	cm1 := concurrence.NewConcurrenceMap[string, int](50)
	cm1.Store("张三", 18)
	if v, exists := cm1.Load("张三"); !exists {
		t.Fail()
	} else {
		if v != 18 {
			t.Fail()
		}
	}
	if _, exists := cm1.Load("李四"); exists {
		t.Fail()
	}

	cm2 := concurrence.NewConcurrenceMap[int, bool](50)
	cm2.Store(18, true)
	if v, exists := cm2.Load(18); !exists {
		t.Fail()
	} else {
		if v != true {
			t.Fail()
		}
	}

	const P = 10
	wg := sync.WaitGroup{}
	wg.Add(P)
	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cm2.Store(j, true)
			}
		}()
	}
	wg.Wait()
}
