package main

import (
	"go_basic/concurrence"
)

func main() {
	// 测试入口，按需打开
	// concurrence.Atomic()
	// ReentrantRLock(2)
	// ReentrantWLock(2)
	// WLockExclusion()
	// RLockExclusion()
	// concurrence.CollectionSafety()
	// concurrence.CutDownLatch()
	// concurrence.Broadcast()
	// concurrence.DealMassFile("data/biz_log")
	concurrence.ListenMultiWay()
}
