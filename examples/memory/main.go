package main

import (
	"fmt"
	"time"

	"github.com/duxweb/go-lock"
	"github.com/duxweb/go-lock/drivers"
)

func main() {
	// 创建内存锁驱动
	driver := drivers.NewMemoryDriver()

	// 创建锁管理器
	manager := lock.New(driver)

	// 演示基本锁功能
	fmt.Println("===== 基本锁功能演示 =====")
	demonstrateBasicLocking(manager)

	// 演示并发锁竞争
	fmt.Println("\n===== 并发锁竞争演示 =====")
	demonstrateConcurrentLocking(manager)
}

func demonstrateBasicLocking(manager *lock.Manager) {
	lockKey := "my-resource"
	ttl := time.Second * 5

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

func demonstrateConcurrentLocking(manager *lock.Manager) {
	lockKey := "shared-resource"
	ttl := time.Second * 3

	// 创建并启动多个 worker，它们会尝试获取同一把锁
	for i := 1; i <= 3; i++ {
		workerID := i
		go func() {
			lock := manager.Create(lockKey, ttl)
			fmt.Printf("工作线程 %d 尝试获取锁...\n", workerID)

			// 尝试获取锁，如果获取不到会等待
			if lock.Acquire(true) {
				fmt.Printf("工作线程 %d 成功获取锁\n", workerID)

				// 模拟处理工作
				fmt.Printf("工作线程 %d 正在执行工作...\n", workerID)
				time.Sleep(time.Second)

				// 完成后释放锁
				if err := lock.Release(); err != nil {
					fmt.Printf("工作线程 %d 释放锁失败: %v\n", workerID, err)
					return
				}
				fmt.Printf("工作线程 %d 完成工作并释放锁\n", workerID)
			} else {
				fmt.Printf("工作线程 %d 无法获取锁\n", workerID)
			}
		}()
	}

	// 等待所有 worker 完成
	time.Sleep(time.Second * 10)
}
