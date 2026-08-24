package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Cluster for distributed data storage
// Cluster provides automatic sharding and high availability
func main() {
	// Cluster configuration
	// In production, you'd have multiple nodes across different machines
	clusterAddrs := []string{
		"localhost:7000",
		"localhost:7001",
		"localhost:7002",
		"localhost:7003",
		"localhost:7004",
		"localhost:7005",
	}

	fmt.Println("=== Redis Cluster Example ===")

	// Create client for Redis Cluster
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: clusterAddrs,
		// Optional: Cluster-specific settings
		// DialTimeout:  5 * time.Second,
		// ReadTimeout:  3 * time.Second,
		// WriteTimeout: 3 * time.Second,
		// MaxRedirects: 3,
		// Password: "",
	})

	ctx := context.Background()

	defer rdb.Close()

	// Test connection
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Printf("Warning: Cannot connect to Cluster: %v", err)
		fmt.Println("Note: This example requires a running Redis Cluster.")
		fmt.Println("For a working example, start a Redis Cluster locally.")
		fmt.Println("\nTo create a test cluster:")
		fmt.Println("1. Start 6 Redis instances on ports 7000-7005")
		fmt.Println("2. Use redis-cli --cluster create 127.0.0.1:7000-7005")
	} else {
		fmt.Println("Connected to Redis Cluster")
	}

	// ========== CLUSTER INFO ==========

	fmt.Println("\n1. Cluster Information")

	info, err := rdb.ClusterInfo(ctx).Result()
	if err != nil {
		log.Printf("ClusterInfo failed: %v", err)
	} else {
		fmt.Println(info)
	}

	fmt.Println("\n2. Cluster Nodes")

	// Get information about all nodes
	nodes, err := rdb.ClusterNodes(ctx).Result()
	if err != nil {
		log.Printf("ClusterNodes failed: %v", err)
	} else {
		fmt.Println(nodes)
	}

	fmt.Println("\n3. Cluster Slots")

	// Get slot coverage information
	// Each key is mapped to a slot (0-16383)
	slots, err := rdb.ClusterSlots(ctx).Result()
	if err != nil {
		log.Printf("ClusterSlots failed: %v", err)
	} else {
		fmt.Printf("Number of slot ranges: %d\n", len(slots))
		for _, slotInfo := range slots {
			fmt.Printf("  Slots %d-%d handled by %s\n",
				slotInfo.Start, slotInfo.End, slotInfo.Nodes[0].Addr)
		}
	}

	// ========== CLUSTER KEYSLOT ==========

	fmt.Println("\n4. Key Slot Calculation")

	// Redis Cluster uses hash slots (0-16383)
	// Keys are mapped to slots using: slot = CRC16(key) mod 16384

	testKeys := []string{"user:1", "user:2", "user:3", "order:1", "order:2"}

	for _, key := range testKeys {
		slot := rdb.ClusterKeySlot(ctx, key)
		fmt.Printf("Key '%s' -> Slot %d\n", key, slot)
	}

	// Keys with the same hashtag (inside {}) go to the same slot
	// This enables multi-key operations across keys
	hashtagKeys := []string{"user:100{order}", "user:200{order}", "user:300{order}"}
	fmt.Println("\nKeys with same hashtag:")
	for _, key := range hashtagKeys {
		slot := rdb.ClusterKeySlot(ctx, key)
		fmt.Printf("  Key '%s' -> Slot %d (same slot!)\n", key, slot)
	}

	// ========== CLUSTER OPERATIONS ==========

	fmt.Println("\n=== Cluster Operations ===")

	if err == nil {
		// Basic operations work the same as with single Redis
		key := "cluster:test"
		err = rdb.Set(ctx, key, "cluster-value", 0).Err()
		if err != nil {
			log.Fatalf("SET failed: %v", err)
		}
		fmt.Printf("Set %s = cluster-value\n", key)

		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			log.Fatalf("GET failed: %v", err)
		}
		fmt.Printf("Retrieved %s = %s\n", key, val)

		// Multi-key operations work only if keys are in the same slot
		// Use {hashtag} to ensure keys go to the same slot

		// This works because keys share the same hashtag
		multiKey1 := "user:1{hash}"
		multiKey2 := "user:2{hash}"

		rdb.MSet(ctx, map[string]interface{}{
			multiKey1: "value1",
			multiKey2: "value2",
		})

		values, err := rdb.MGet(ctx, multiKey1, multiKey2).Result()
		if err != nil {
			log.Printf("MGET failed: %v", err)
		} else {
			fmt.Printf("MGET results: %v\n", values)
		}

		fmt.Println("\n5. Cluster Failover (Manual)")

		fmt.Println(`To manually trigger a failover, use redis-cli:
  redis-cli -c -h <host> -p <port> cluster failover

Or connect to a slave and run:
  CLUSTER FAILOVER
`)

		// Cleanup
		rdb.Del(ctx, key, multiKey1, multiKey2)
	}

	fmt.Println("\n=== Cluster Configuration Notes ===")
	fmt.Println(`
Redis Cluster features:
  - Automatic sharding across 16384 hash slots
  - Built-in replication (each shard has master + slaves)
  - Automatic failover when master fails
  - No proxy layer needed (clients connect directly to nodes)

Typical cluster setup:
  - Minimum 6 nodes (3 masters + 3 slaves)
  - Each master handles ~5461 slots (16384 / 3)
  - Slaves provide redundancy

Creating a cluster (using redis-cli):
  redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
            127.0.0.1:7002 127.0.0.1:7003 \
            127.0.0.1:7004 127.0.0.1:7005 --cluster-replicas 1

Client configuration in go-redis:
  - Use redis.NewClusterClient() for cluster mode
  - Client handles node discovery automatically
  - Automatic key-to-slot mapping
  - Automatic redirect handling

Key hashing for multi-key operations:
  Keys inside {} braces use only the braces' content for slot calculation
  Example: user:{100}:order and user:{100}:profile are in the same slot
`)
}
