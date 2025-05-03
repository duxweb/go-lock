# go-lock

go-lock 是一个用于Go应用程序的分布式锁库，提供了简单而强大的API来实现分布式锁功能。
go-lock is a distributed lock library for Go applications, providing a simple yet powerful API for distributed locking.

## 特性 | Features

- 支持多种锁实现（内存、Redis）| Support for multiple lock implementations (Memory, Redis)
- 简洁的接口，易于使用 | Clean interface, easy to use
- 支持锁超时和自动释放 | Support for lock timeout and auto-release
- 可等待和非等待模式 | Waiting and non-waiting modes
- 高度可扩展的架构，易于添加新的锁实现 | Highly extensible architecture, easy to add new lock implementations

## 安装 | Installation

```bash
go get github.com/duxweb/go-lock
```

## 快速开始 | Quick Start

### 使用内存锁（单机应用）| Using Memory Lock (Single Application)

```go
package main

import (
	"fmt"
	"time"

	"github.com/duxweb/go-lock"
	"github.com/duxweb/go-lock/drivers"
)

func main() {
	// 创建内存锁驱动 | Create memory lock driver
	driver := drivers.NewMemoryDriver()

	// 创建锁管理器 | Create lock manager
	manager := lock.New(driver)

	// 创建一个锁 | Create a lock
	myLock := manager.Create("my-resource", time.Second*10)

	// 尝试获取锁 | Try to acquire the lock
	if myLock.Acquire(false) {
		fmt.Println("锁获取成功 | Lock acquired successfully")

		// 执行受保护的操作 | Execute protected operations
		// ...

		// 释放锁 | Release the lock
		myLock.Release()
	} else {
		fmt.Println("无法获取锁，资源正在被使用 | Cannot acquire lock, resource is in use")
	}
}
```

### 使用Redis分布式锁 | Using Redis Distributed Lock

```go
package main

import (
	"fmt"
	"time"

	"github.com/duxweb/go-lock"
	"github.com/duxweb/go-lock/drivers"
)

func main() {
	// 创建Redis锁驱动（默认连接localhost:6379）| Create Redis lock driver (default connection to localhost:6379)
	driver, err := drivers.NewRedisDriver(&drivers.RedisOptions{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Prefix:   "my-app:",
	})
	if err != nil {
		panic(err)
	}

	// 创建锁管理器 | Create lock manager
	manager := lock.New(driver)

	// 创建一个锁，超时时间为10秒 | Create a lock with 10 seconds timeout
	myLock := manager.Create("my-resource", time.Second*10)

	// 尝试获取锁，如果不可用则等待 | Try to acquire the lock, wait if not available
	if myLock.Acquire(true) {
		fmt.Println("锁获取成功 | Lock acquired successfully")

		// 执行受保护的操作 | Execute protected operations
		// ...

		// 释放锁 | Release the lock
		myLock.Release()
	} else {
		fmt.Println("无法获取锁 | Cannot acquire lock")
	}
}
```

## 锁驱动 | Lock Drivers

### 内存锁 | Memory Lock

适用于单机应用和单进程内的锁定。不适用于分布式场景。
Suitable for single-machine applications and locking within a single process. Not suitable for distributed scenarios.

```go
driver := drivers.NewMemoryDriver()
```

### Redis锁 | Redis Lock

适用于分布式应用程序，基于Redis实现的分布式锁。
Suitable for distributed applications, based on Redis distributed lock implementation.

```go
driver, err := drivers.NewRedisDriver(&drivers.RedisOptions{
	Addr:     "localhost:6379",
	Password: "your-password",
	DB:       0,
	Prefix:   "lock-prefix:",
	// 可选配置 | Optional configuration
	PoolSize:     10,
	MinIdleConns: 5,
})
```

## 完整示例 | Complete Examples

完整的示例可在 `examples` 目录下找到：
Complete examples can be found in the `examples` directory:

- 内存锁示例 | Memory lock example: [examples/memory/main.go](examples/memory/main.go)
- Redis锁示例 | Redis lock example: [examples/redis/main.go](examples/redis/main.go)

## 测试 | Testing

项目的测试文件位于每个驱动的实现目录中：
Test files are located in each driver's implementation directory:

### 运行所有测试 | Run All Tests

```bash
go test ./...
```

### 运行内存锁测试 | Run Memory Lock Tests

```bash
go test ./drivers -run TestMemoryLock
```

### 运行Redis锁测试 | Run Redis Lock Tests

默认会尝试连接 localhost:6379，无需额外配置：
By default, it will try to connect to localhost:6379, no additional configuration required:

```bash
go test ./drivers -run TestRedisLock
```

如需指定其他Redis服务器：
To specify another Redis server:

```bash
REDIS_ADDR=192.168.1.100:6379 REDIS_PASSWORD=your-password go test ./drivers -run TestRedisLock
```

### 测试覆盖率 | Test Coverage

核心代码覆盖率为 100%（主包）和 98.5%（驱动包）：
Core code coverage is 100% (main package) and 98.5% (drivers package):

```
github.com/duxweb/go-lock/drivers/memory.go:18: NewMemoryDriver 100.0%
github.com/duxweb/go-lock/drivers/memory.go:26: Create          100.0%
github.com/duxweb/go-lock/drivers/memory.go:52: Acquire         100.0%
github.com/duxweb/go-lock/drivers/memory.go:77: Release         100.0%
github.com/duxweb/go-lock/drivers/redis.go:37:  NewRedisDriver  100.0%
github.com/duxweb/go-lock/drivers/redis.go:67:  Create          100.0%
github.com/duxweb/go-lock/drivers/redis.go:89:  getKey          100.0%
github.com/duxweb/go-lock/drivers/redis.go:98:  Acquire         100.0%
github.com/duxweb/go-lock/drivers/redis.go:124: Release         91.7%
github.com/duxweb/go-lock/main.go:15:                   New             100.0%
github.com/duxweb/go-lock/main.go:23:                   Create          100.0%
```

生成覆盖率报告：
Generate coverage report:

```bash
go test ./drivers -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

## 许可证 | License

[MIT](LICENSE)