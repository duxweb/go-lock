package drivers

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/duxweb/go-lock"
)

// RedisProvider Redis锁驱动提供者
// RedisProvider Redis lock provider
type RedisProvider struct {
	client *redis.Client
	prefix string
}

// RedisOptions Redis锁配置选项
// RedisOptions Redis lock configuration options
type RedisOptions struct {
	Client       *redis.Client
	Addr         string
	Password     string
	DB           int
	Prefix       string
	PoolSize     int
	MinIdleConns int
}

// NewRedisDriver 创建Redis锁驱动
// NewRedisDriver creates a Redis lock driver
func NewRedisDriver(opts *RedisOptions) (*RedisProvider, error) {
	var client *redis.Client

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if opts.Client != nil {
		client = opts.Client
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         opts.Addr,
			Password:     opts.Password,
			DB:           opts.DB,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
		})
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis: %w", err)
	}

	return &RedisProvider{
		client: client,
		prefix: opts.Prefix,
	}, nil
}

// Create 创建一个Redis锁
// Create creates a Redis lock
func (r *RedisProvider) Create(key string, ttl time.Duration) lock.LockDriver {
	return &RedisLockDriver{
		key:        key,
		ttl:        ttl,
		client:     r.client,
		prefix:     r.prefix,
		lockValues: &sync.Map{},
	}
}

// RedisLockDriver Redis锁驱动实例
// RedisLockDriver Redis lock driver instance
type RedisLockDriver struct {
	key        string
	ttl        time.Duration
	client     *redis.Client
	lockValues *sync.Map
	prefix     string
}

// getKey 获取带前缀的键名
// getKey gets the key with prefix
func (r *RedisLockDriver) getKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + ":" + key
}

// Acquire 尝试获取一个锁
// Acquire attempts to acquire a lock
func (r *RedisLockDriver) Acquire(wait bool) bool {
	lockValue := uuid.New().String()
	redisKey := r.getKey(r.key)
	deadline := time.Now().Add(30 * time.Second)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		cmd := r.client.SetNX(ctx, redisKey, lockValue, r.ttl)
		err := cmd.Err()
		cancel()

		if err == nil && cmd.Val() {
			r.lockValues.Store(redisKey, lockValue)
			return true
		}

		if !wait || time.Now().After(deadline) {
			return false
		}

		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	}
}

// Release 释放一个锁
// Release releases a lock
func (r *RedisLockDriver) Release() error {
	redisKey := r.getKey(r.key)

	lockValue, ok := r.lockValues.Load(redisKey)
	if !ok {
		return nil
	}

	defer r.lockValues.Delete(redisKey)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	_, err := r.client.Eval(ctx, script, []string{redisKey}, lockValue).Result()
	if err != nil {
		return err
	}

	return nil
}
