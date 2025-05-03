package main

import (
	"fmt"
	"os"
	"time"

	"github.com/duxweb/go-lock"
	"github.com/duxweb/go-lock/drivers"
)

func main() {
	// 默认使用 localhost:6379
	redisAddr := "localhost:6379"
	redisPassword := ""

	// 允许通过环境变量覆盖
	if envAddr := os.Getenv("REDIS_ADDR"); envAddr != "" {
		redisAddr = envAddr
	}
	if envPass := os.Getenv("REDIS_PASSWORD"); envPass != "" {
		redisPassword = envPass
	}

	// 创建Redis锁驱动
	driver, err := drivers.NewRedisDriver(&drivers.RedisOptions{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
		Prefix:   "example-lock:",
	})
	if err != nil {
		fmt.Printf("创建Redis锁驱动失败: %v\n", err)
		return
	}

	// 创建锁管理器
	manager := lock.New(driver)

	// 演示基本锁功能
	fmt.Println("===== 基本锁功能演示 =====")
	demonstrateBasicLocking(manager)

	// 演示分布式锁（在多个进程间运行此例子可以看到效果）
	fmt.Println("\n===== 分布式锁演示 =====")
	demonstrateDistributedLocking(manager)
}

func demonstrateBasicLocking(manager *lock.Manager) {
	lockKey := "my-resource"
	ttl := time.Second * 10

	// 创建锁
	lock1 := manager.Create(lockKey, ttl)

	// 尝试获取锁
	fmt.Println("尝试获取锁...")
	if lock1.Acquire(false) {
		fmt.Println("成功获取锁！")
		defer func() {
			if err := lock1.Release(); err != nil {
				fmt.Printf("释放锁失败: %v\n", err)
				return
			}
			fmt.Println("锁已释放")
		}()

		// 模拟执行受保护的操作
		fmt.Println("执行受保护的操作...")
		time.Sleep(time.Second * 2)

		// 创建另一个锁实例，尝试同时获取锁
		fmt.Println("创建第二个锁实例...")
		lock2 := manager.Create(lockKey, ttl)

		fmt.Println("第二个锁实例尝试获取锁...")
		if lock2.Acquire(false) {
			fmt.Println("第二个锁实例获取锁成功 - 这不应该发生！")
		} else {
			fmt.Println("第二个锁实例获取锁失败，锁已被占用")
		}
	} else {
		fmt.Println("无法获取锁")
	}
}

func demonstrateDistributedLocking(manager *lock.Manager) {
	lockKey := "distributed-resource"
	ttl := time.Second * 30

	// 创建锁
	lock := manager.Create(lockKey, ttl)

	// 尝试获取锁
	fmt.Println("尝试获取分布式锁...")
	if lock.Acquire(false) {
		fmt.Println("成功获取分布式锁！")
		fmt.Printf("锁 '%s' 已被此进程持有，PID: %d\n", lockKey, os.Getpid())

		fmt.Println("模拟长时间运行的任务...")
		fmt.Println("您可以在另一个终端窗口运行相同的示例，它将无法获取锁")
		fmt.Println("按Ctrl+C终止此进程来释放锁")

		// 保持锁持有状态，直到用户中断程序
		for {
			time.Sleep(time.Second * 5)
			fmt.Printf("进程 %d 仍然持有锁 '%s'...\n", os.Getpid(), lockKey)
		}
	} else {
		fmt.Println("无法获取分布式锁，可能被其他进程持有")
		fmt.Println("尝试等待获取锁...")

		// 尝试等待获取锁
		if lock.Acquire(true) {
			fmt.Println("成功等待并获取分布式锁！")
			fmt.Printf("锁 '%s' 现在被此进程持有，PID: %d\n", lockKey, os.Getpid())

			// 持有锁5秒
			time.Sleep(time.Second * 5)

			// 释放锁
			if err := lock.Release(); err != nil {
				fmt.Printf("释放锁失败: %v\n", err)
				return
			}
			fmt.Println("锁已释放")
		} else {
			fmt.Println("等待获取锁失败")
		}
	}
}
