package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Sentinel for high availability
// Sentinel provides monitoring, failover, and configuration services
func main() {
	// Sentinel configuration
	// In production, you'd have multiple sentinels monitoring your Redis instances
	sentinelAddrs := []string{
		"localhost:26379",
		"localhost:26380",
		"localhost:26381",
	}

	// Redis master name (configured in sentinel.conf)
	masterName := "mymaster"

	fmt.Println("=== Redis Sentinel Example ===")

	// Create client for Sentinel
	// Note: go-redis uses the same API for both direct Redis and Sentinel
	// It automatically handles master discovery and failover
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      "", // Optional: password for Redis
		DB:            0,
	})

	ctx := context.Background()

	defer rdb.Close()

	// Test connection
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Printf("Warning: Cannot connect to Sentinel/Redis: %v", err)
		fmt.Println("Note: This example requires a running Sentinel setup.")
		fmt.Println("For a working example, start Redis and Sentinel locally.")
		fmt.Println("\nTo set up Redis Sentinel:")
		fmt.Println("1. Start Redis instances on ports 7000, 7001, 7002")
		fmt.Println("2. Start Sentinel instances on ports 26379, 26380, 26381")
		fmt.Println("3. Configure sentinel.conf with: sentinel monitor mymaster 127.0.0.1 7000 2")
	} else {
		fmt.Println("Connected to Redis via Sentinel")
	}

	// ========== BASIC REDIS OPERATIONS WITH SENTINEL ==========

	fmt.Println("\n=== Basic Redis Operations ===")

	if err == nil {
		// Set a value
		err = rdb.Set(ctx, "ha:test", "hello-sentinel", 0).Err()
		if err != nil {
			log.Fatalf("SET failed: %v", err)
		}
		fmt.Println("Set ha:test = hello-sentinel")

		// Read the value back (should work even if failover occurred)
		val, err := rdb.Get(ctx, "ha:test").Result()
		if err != nil {
			log.Fatalf("GET failed: %v", err)
		}
		fmt.Printf("Retrieved ha:test = %s\n", val)

		// Cleanup
		rdb.Del(ctx, "ha:test")
	}

	fmt.Println("\n=== Sentinel Configuration Notes ===")
	fmt.Println(`
Sentinel is a high availability solution for Redis that provides:
  1. Monitoring: Checks if master and slave instances are working
  2. Automatic Failover: Promotes a slave to master when master fails
  3. Configuration Provider: Clients query Sentinel for current master address
  4. Notification: Notifies admins about Redis instance issues

Using go-redis with Sentinel:
  - Use redis.NewFailoverClient() with MasterName and SentinelAddrs
  - The client automatically discovers the current master
  - On failover, the client reconnects to the new master

Typical Sentinel setup:
  - 3+ Sentinel instances (odd number for quorum)
  - 1 Redis master
  - 2+ Redis slaves

Configuration example in sentinel.conf:
  sentinel monitor mymaster 127.0.0.1 6379 2
  sentinel down-after-mymaster 30000
  sentinel failover-timeout mymaster 180000
  sentinel parallel-syncs mymaster 1
`)
}