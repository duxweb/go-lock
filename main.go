package lock

import (
	"time"
)

// Manager 锁管理器
// Manager lock manager
type Manager struct {
	driver LockProvider
}

// New 创建一个锁管理器
// New creates a lock manager.
func New(driver LockProvider) *Manager {
	return &Manager{
		driver: driver,
	}
}

// Create 创建一个锁
// Create creates a lock.
func (m *Manager) Create(key string, ttl time.Duration) LockDriver {
	return m.driver.Create(key, ttl)
}
