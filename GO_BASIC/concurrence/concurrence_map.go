package concurrence

import "sync"

type ConcurrenceMap[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

func NewConcurrenceMap[K comparable, V any](cap int) *ConcurrenceMap[K, V] {
	return &ConcurrenceMap[K, V]{
		data: make(map[K]V, cap),
	}
}

func (cm *ConcurrenceMap[K, V]) Store(key K, value V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.data[key] = value
}

func (cm *ConcurrenceMap[K, V]) Load(key K) (value V, exists bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	value, exists = cm.data[key]
	return
}
