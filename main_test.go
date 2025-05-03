package lock

import (
	"testing"
	"time"
)

// mockProvider 用于测试的模拟锁提供程序
type mockProvider struct {
	createCalled bool
	key          string
	ttl          time.Duration
	driver       *mockDriver
}

// mockDriver 用于测试的模拟锁驱动
type mockDriver struct {
	acquireCalled bool
	releaseCalled bool
	waitFlag      bool
}

func (p *mockProvider) Create(key string, ttl time.Duration) LockDriver {
	p.createCalled = true
	p.key = key
	p.ttl = ttl
	p.driver = &mockDriver{}
	return p.driver
}

func (d *mockDriver) Acquire(wait bool) bool {
	d.acquireCalled = true
	d.waitFlag = wait
	return true
}

func (d *mockDriver) Release() error {
	d.releaseCalled = true
	return nil
}

func TestNew(t *testing.T) {
	provider := &mockProvider{}
	manager := New(provider)

	if manager == nil {
		t.Fatal("New should return a non-nil Manager")
	}

	if manager.driver != provider {
		t.Fatal("Manager should store the provider")
	}
}

func TestManagerCreate(t *testing.T) {
	provider := &mockProvider{}
	manager := New(provider)

	key := "test-lock"
	ttl := time.Second * 30

	// 创建锁
	lock := manager.Create(key, ttl)

	if !provider.createCalled {
		t.Fatal("Manager.Create should call provider.Create")
	}

	if provider.key != key {
		t.Fatalf("Expected key %s, got %s", key, provider.key)
	}

	if provider.ttl != ttl {
		t.Fatalf("Expected TTL %v, got %v", ttl, provider.ttl)
	}

	if lock == nil {
		t.Fatal("Create should return a non-nil LockDriver")
	}

	// 确认返回的是正确的驱动实例
	if lock != provider.driver {
		t.Fatal("The returned lock should be the one created by provider")
	}
}
