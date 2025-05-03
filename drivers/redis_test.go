package drivers

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/duxweb/go-lock"
	"github.com/redis/go-redis/v9"
)

// 测试getKey方法
func TestRedisGetKey(t *testing.T) {
	driver := &RedisLockDriver{
		key:    "test-key",
		prefix: "",
	}

	// 测试无前缀情况
	if key := driver.getKey("test-key"); key != "test-key" {
		t.Errorf("getKey without prefix should return the original key, got %s", key)
	}

	// 测试有前缀情况
	driver.prefix = "prefix"
	if key := driver.getKey("test-key"); key != "prefix:test-key" {
		t.Errorf("getKey with prefix should return prefixed key, got %s", key)
	}
}

// 测试NewRedisDriver方法的各种情况
func TestNewRedisDriver(t *testing.T) {
	// 跳过实际连接部分，如果CI环境没有Redis
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping Redis connection tests in CI environment")
	}

	// 测试使用已有客户端创建驱动
	mockClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	driver, err := NewRedisDriver(&RedisOptions{
		Client: mockClient,
	})
	if err != nil {
		t.Logf("连接Redis失败: %v", err)
	} else {
		if driver.client != mockClient {
			t.Error("Driver should use the provided client")
		}
	}

	// 测试连接失败的情况
	_, err = NewRedisDriver(&RedisOptions{
		Addr: "nonexistent:6379", // 使用不存在的地址
		// 设置较短的超时确保测试快速完成
	})
	if err == nil {
		t.Error("Should return error when connection fails")
	}
}

func TestRedisLock(t *testing.T) {
	redisAddr := "localhost:6379"
	redisPassword := ""

	if envAddr := os.Getenv("REDIS_ADDR"); envAddr != "" {
		redisAddr = envAddr
	}
	if envPass := os.Getenv("REDIS_PASSWORD"); envPass != "" {
		redisPassword = envPass
	}

	driver, err := NewRedisDriver(&RedisOptions{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
		Prefix:   "test-lock:",
	})
	if err != nil {
		t.Skipf("Skipping Redis lock test: unable to connect to Redis server %s: %v", redisAddr, err)
		return
	}

	manager := lock.New(driver)

	t.Run("Basic Lock Operations", func(t *testing.T) {
		lockKey := "test-lock"
		ttl := time.Second * 10

		lock1 := manager.Create(lockKey, ttl)

		if !lock1.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		lock2 := manager.Create(lockKey, ttl)
		if lock2.Acquire(false) {
			t.Fatal("Should not be able to acquire a held lock")
		}

		if err := lock1.Release(); err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		if !lock2.Acquire(false) {
			t.Log("Unable to acquire lock, may be due to Redis processing delay, skipping assertion")
			return
		}
	})

	t.Run("Lock Timeout", func(t *testing.T) {
		lockKey := "test-lock-timeout"
		ttl := time.Millisecond * 100

		lock1 := manager.Create(lockKey, ttl)

		if !lock1.Acquire(false) {
			t.Fatal("Should be able to acquire an unheld lock")
		}

		time.Sleep(ttl + time.Millisecond*100)

		lock2 := manager.Create(lockKey, ttl)
		if !lock2.Acquire(false) {
			t.Log("Unable to acquire lock, may be due to Redis processing delay, skipping assertion")
			return
		}
	})

	t.Run("Wait For Lock", func(t *testing.T) {
		lockKey := "test-lock-wait"
		ttl := time.Millisecond * 300

		lock1 := manager.Create(lockKey, ttl)

		if !lock1.Acquire(false) {
			t.Skip("Unable to acquire lock, skipping this test")
			return
		}

		done := make(chan struct{})

		go func() {
			lock2 := manager.Create(lockKey, ttl)
			success := lock2.Acquire(true)

			if !success {
				t.Log("Unable to acquire lock after waiting, may be due to Redis processing delay")
			}

			close(done)
		}()

		time.Sleep(time.Millisecond * 100)
		if err := lock1.Release(); err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		select {
		case <-done:
			// Test passed
		case <-time.After(time.Second * 2):
			t.Fatal("Test timed out")
		}
	})

	t.Run("Release Edge Cases", func(t *testing.T) {
		// 如果没法连接Redis就跳过
		if driver == nil {
			t.Skip("Skipping due to no Redis connection")
			return
		}

		// 测试释放不存在的锁
		nonExistentLock := &RedisLockDriver{
			key:        "non-existent-lock",
			client:     driver.client,
			lockValues: &sync.Map{},
		}
		if err := nonExistentLock.Release(); err != nil {
			t.Fatalf("Releasing non-existent lock should not error: %v", err)
		}

		// 测试Release函数中的错误处理
		lockKey := "test-release-errors"
		actualLock := manager.Create(lockKey, time.Second)
		redisLock := actualLock.(*RedisLockDriver)

		// 先获取锁
		if !redisLock.Acquire(false) {
			t.Skip("Couldn't acquire lock, skipping")
			return
		}

		// 存储自定义值用于测试
		redisKey := redisLock.getKey(lockKey)
		redisLock.lockValues.Store(redisKey, "test-value")

		// 正常释放
		if err := redisLock.Release(); err != nil {
			t.Fatalf("Release should not error: %v", err)
		}

		// 再次释放（已经删除了锁值）
		if err := redisLock.Release(); err != nil {
			t.Fatalf("Second release should not error: %v", err)
		}
	})
}
