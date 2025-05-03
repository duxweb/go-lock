package lock

import (
	"time"
)

// LockProvider 分布式锁提供者
// LockProvider is the interface for a lock provider.
type LockProvider interface {
	Create(key string, ttl time.Duration) LockDriver
}

// LockDriver 分布式锁驱动
// LockDriver is the interface for a lock driver.
type LockDriver interface {

	// Acquire 尝试获取一个锁
	// Acquire attempts to acquire a lock.
	Acquire(wait bool) bool

	// Release 释放一个锁
	// Release releases a lock.
	Release() error
}
