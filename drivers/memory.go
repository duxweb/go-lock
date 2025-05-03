package drivers

import (
	"sync"
	"time"

	"github.com/duxweb/go-lock"
)

// MemoryProvider 内存锁驱动
// MemoryProvider memory lock driver
type MemoryProvider struct {
	locks sync.Map
}

// NewMemoryDriver 创建内存锁驱动
// NewMemoryDriver creates a memory lock driver
func NewMemoryDriver() *MemoryProvider {
	return &MemoryProvider{
		locks: sync.Map{},
	}
}

// Create 创建一个内存锁
// Create creates a memory lock
func (r *MemoryProvider) Create(key string, ttl time.Duration) lock.LockDriver {
	return &MemoryLockDriver{
		provider: r,
		key:      key,
		ttl:      ttl,
	}
}

// MemoryLockDriver 内存锁驱动实例
// MemoryLockDriver memory lock driver instance
type MemoryLockDriver struct {
	provider *MemoryProvider
	key      string
	ttl      time.Duration
}

// memLock 内存锁数据结构
// memLock memory lock data structure
type memLock struct {
	mutex   sync.Mutex
	locked  bool
	timeout time.Time
}

// Acquire 尝试获取一个锁
// Acquire attempts to acquire a lock
func (r *MemoryLockDriver) Acquire(wait bool) bool {
	for {

		actual, _ := r.provider.locks.LoadOrStore(r.key, &memLock{})
		lock := actual.(*memLock)

		lock.mutex.Lock()
		now := time.Now()
		if !lock.locked || now.After(lock.timeout) {
			lock.locked = true
			lock.timeout = now.Add(r.ttl)
			lock.mutex.Unlock()
			return true
		}
		lock.mutex.Unlock()

		if !wait {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Release 释放一个锁
// Release releases a lock
func (r *MemoryLockDriver) Release() error {
	actual, ok := r.provider.locks.Load(r.key)
	if !ok {
		return nil
	}
	lock := actual.(*memLock)
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	if !lock.locked {
		return nil
	}
	lock.locked = false
	return nil
}
