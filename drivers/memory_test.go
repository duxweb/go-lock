package drivers

import (
	"testing"
	"time"

	"github.com/duxweb/go-lock"
)

func TestMemoryLock(t *testing.T) {
	// 创建内存锁驱动
	driver := NewMemoryDriver()

	// 创建锁管理器
	manager := lock.New(driver)

	// 测试基本锁获取和释放
	t.Run("Basic Lock Operations", func(t *testing.T) {
		lockKey := "test-lock"
		ttl := time.Second * 10

		lock1 := manager.Create(lockKey, ttl)

		// 获取锁应该成功
		if !lock1.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		// 重复获取锁应该失败
		lock2 := manager.Create(lockKey, ttl)
		if lock2.Acquire(false) {
			t.Fatal("Should not be able to acquire a held lock")
		}

		// 释放锁
		if err := lock1.Release(); err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		// 现在应该能获取锁了 (添加短暂延迟确保锁释放生效)
		time.Sleep(10 * time.Millisecond)
		if !lock2.Acquire(false) {
			t.Fatal("Should be able to acquire after release")
		}
	})

	// 测试锁超时
	t.Run("Lock Timeout", func(t *testing.T) {
		lockKey := "test-lock-timeout"
		ttl := time.Millisecond * 100

		lock1 := manager.Create(lockKey, ttl)

		// 获取锁应该成功
		if !lock1.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		// 等待超时
		time.Sleep(ttl + time.Millisecond*50)

		// 超时后应该能获取锁
		lock2 := manager.Create(lockKey, ttl)
		if !lock2.Acquire(false) {
			t.Fatal("Should be able to acquire after timeout")
		}
	})

	// 测试等待获取锁
	t.Run("Wait For Lock", func(t *testing.T) {
		lockKey := "test-lock-wait"
		ttl := time.Millisecond * 300

		lock1 := manager.Create(lockKey, ttl)

		// 获取锁应该成功
		if !lock1.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		// 同步通道，用于确认测试完成
		done := make(chan struct{})

		// 在另一个goroutine中等待获取锁
		go func() {
			lock2 := manager.Create(lockKey, ttl)

			start := time.Now()
			// 等待获取锁
			success := lock2.Acquire(true)
			elapsed := time.Since(start)

			if !success {
				t.Error("Should be able to acquire after waiting")
			}

			// 检查时间是否合理，至少应该等待一小段时间
			if elapsed < time.Millisecond*10 {
				t.Errorf("Lock acquisition too fast: %v, expected at least some delay", elapsed)
			}

			close(done)
		}()

		// 等待一段时间后释放锁
		time.Sleep(time.Millisecond * 50)
		if err := lock1.Release(); err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		// 等待goroutine完成
		select {
		case <-done:
			// Test passed
		case <-time.After(time.Second):
			t.Fatal("Test timed out")
		}
	})

	// 测试Release方法的特殊情况
	t.Run("Release Edge Cases", func(t *testing.T) {
		// 测试释放不存在的锁
		nonExistentLock := manager.Create("non-existent-lock", time.Second)
		if err := nonExistentLock.Release(); err != nil {
			t.Fatalf("Releasing non-existent lock should not error: %v", err)
		}

		// 测试释放未锁定的锁
		lockKey := "unlocked-lock"
		lock := manager.Create(lockKey, time.Second)

		// 先获取锁，然后释放
		if !lock.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		// 第一次释放
		if err := lock.Release(); err != nil {
			t.Fatalf("First release should succeed: %v", err)
		}

		// 再次释放同一把锁
		if err := lock.Release(); err != nil {
			t.Fatalf("Second release of the same lock should not error: %v", err)
		}
	})
}
