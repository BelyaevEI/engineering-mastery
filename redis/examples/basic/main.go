package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Basic operations with Redis using go-redis library
func main() {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password by default
		DB:       0,  // default DB
	})

	ctx := context.Background()

	// Ping server to check connection
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis successfully")

	// ========== STRING OPERATIONS ==========

	// SET command - set a key with value
	key := "user:1"
	err = rdb.Set(ctx, key, "John Doe", 0).Err()
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	fmt.Printf("SET %s: John Doe\n", key)

	// GET command - get value by key
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("GET %s: %s\n", key, val)

	// GETSET command - set new value and return old one
	oldVal, err := rdb.GetSet(ctx, key, "Jane Smith").Result()
	if err != nil {
		log.Fatalf("Failed to getset key: %v", err)
	}
	fmt.Printf("GETSET %s: old value was '%s', new value is 'Jane Smith'\n", key, oldVal)

	// EXPIRE command - set key expiration (TTL)
	err = rdb.Expire(ctx, key, 5*time.Second).Err()
	if err != nil {
		log.Fatalf("Failed to set expiration: %v", err)
	}
	fmt.Printf("SET expiration for %s to 5 seconds\n", key)

	// TTL command - check remaining time to live
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		log.Fatalf("Failed to get TTL: %v", err)
	}
	fmt.Printf("TTL for %s: %v\n", key, ttl)

	// Exists command - check if key exists
	exists, err := rdb.Exists(ctx, key).Result()
	fmt.Printf("Key %s exists: %v\n", key, exists)

	// DEL command - delete key
	err = rdb.Del(ctx, key).Err()
	if err != nil {
		log.Fatalf("Failed to delete key: %v", err)
	}
	fmt.Printf("Deleted key %s\n", key)

	// ========== LIST OPERATIONS ==========

	listKey := "messages"
	// LPUSH - add element to the beginning of list
	err = rdb.LPush(ctx, listKey, "third", "second", "first").Err()
	if err != nil {
		log.Fatalf("Failed to lpush: %v", err)
	}
	fmt.Printf("LPUSH %s: added elements\n", listKey)

	// LRANGE - get range of elements
	elements, err := rdb.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		log.Fatalf("Failed to lrange: %v", err)
	}
	fmt.Printf("LRANGE %s 0 -1: %v\n", listKey, elements)

	// LPOP - remove and get first element
	val, err = rdb.LPop(ctx, listKey).Result()
	if err != nil {
		log.Fatalf("Failed to lpop: %v", err)
	}
	fmt.Printf("LPOP %s: removed '%s'\n", listKey, val)

	// ========== HASH OPERATIONS ==========

	hashKey := "user:profile"
	// HSET - set hash field
	err = rdb.HSet(ctx, hashKey, map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}).Err()
	if err != nil {
		log.Fatalf("Failed to hset: %v", err)
	}
	fmt.Printf("HSET %s: set user profile\n", hashKey)

	// HGET - get single field value
	name, err := rdb.HGet(ctx, hashKey, "name").Result()
	if err != nil {
		log.Fatalf("Failed to hget: %v", err)
	}
	fmt.Printf("HGET %s name: %s\n", hashKey, name)

	// HGETALL - get all fields and values
	allFields, err := rdb.HGetAll(ctx, hashKey).Result()
	if err != nil {
		log.Fatalf("Failed to hgetall: %v", err)
	}
	fmt.Printf("HGETALL %s: %v\n", hashKey, allFields)

	// HDEL - delete field
	err = rdb.HDel(ctx, hashKey, "age").Err()
	if err != nil {
		log.Fatalf("Failed to hdel: %v", err)
	}
	fmt.Printf("HDEL %s age: deleted age field\n", hashKey)

	// ========== SET OPERATIONS ==========

	setKey := "tags"
	// SADD - add member to set
	err = rdb.SAdd(ctx, setKey, "golang", "redis", "database", "redis").Err()
	if err != nil {
		log.Fatalf("Failed to sadd: %v", err)
	}
	fmt.Printf("SADD %s: added tags (redis added twice, only stored once)\n", setKey)

	// SMEMBERS - get all members
	members, err := rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		log.Fatalf("Failed to smembers: %v", err)
	}
	fmt.Printf("SMEMBERS %s: %v\n", setKey, members)

	// SISMEMBER - check if member exists
	isMember, err := rdb.SIsMember(ctx, setKey, "golang").Result()
	fmt.Printf("SISMEMBER %s golang: %v\n", setKey, isMember)

	// SCARD - get set cardinality
	card, err := rdb.SCard(ctx, setKey).Result()
	fmt.Printf("SCARD %s: %d members\n", setKey, card)

	// ========== SORTED SET OPERATIONS ==========

	zsetKey := "scores"
	// ZADD - add member with score (go-redis v9: variadic redis.Z arguments)
	_, err = rdb.ZAdd(ctx, zsetKey,
		redis.Z{Score: 100, Member: "user1"},
		redis.Z{Score: 85, Member: "user2"},
		redis.Z{Score: 95, Member: "user3"},
		redis.Z{Score: 100, Member: "user4"},
	).Result()
	if err != nil {
		log.Fatalf("Failed to zadd: %v", err)
	}
	fmt.Printf("ZADD %s: added scores\n", zsetKey)

	// ZRANGE - get members by range
	zrange, err := rdb.ZRange(ctx, zsetKey, 0, -1).Result()
	if err != nil {
		log.Fatalf("Failed to zrange: %v", err)
	}
	fmt.Printf("ZRANGE %s 0 -1: %v\n", zsetKey, zrange)

	// ZRANGEBYSCORE - get members by score range
	zrangeByScore, err := rdb.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Min:    "90",
		Max:    "100",
		Offset: 0,
		Count:  10,
	}).Result()
	if err != nil {
		log.Fatalf("Failed to zrangebyscore: %v", err)
	}
	fmt.Printf("ZRANGEBYSCORE %s 90-100: %v\n", zsetKey, zrangeByScore)

	// ZREM - remove member
	err = rdb.ZRem(ctx, zsetKey, "user2").Err()
	if err != nil {
		log.Fatalf("Failed to zrem: %v", err)
	}
	fmt.Printf("ZREM %s user2: removed user2\n", zsetKey)

	// ========== BULK OPERATIONS ==========

	// Pipeline - execute multiple commands in one batch
	pipe := rdb.TxPipeline()
	pipe.Set(ctx, "pipe:1", "value1", 0)
	pipe.Set(ctx, "pipe:2", "value2", 0)
	pipe.Get(ctx, "pipe:1")
	pipe.Incr(ctx, "counter")
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to execute pipeline: %v", err)
	}
	fmt.Println("Pipeline executed successfully")

	// ========== SCAN OPERATIONS ==========

	// Scan - iterate over keys (non-blocking)
	// go-redis v9: Scan returns keys, cursor, error
	var cursor uint64
	var keys []string
	for {
		keys, cursor, err = rdb.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("Scan found keys: %v\n", keys)
		if cursor == 0 {
			break
		}
	}

	// ========== COUNTER OPERATIONS ==========

	counterKey := "visits"
	// INCR - increment by 1 (returns int64 in go-redis v9)
	incremented, err := rdb.Incr(ctx, counterKey).Result()
	if err != nil {
		log.Fatalf("Failed to incr: %v", err)
	}
	fmt.Printf("INCR %s: %d\n", counterKey, incremented)

	// INCRBY - increment by N (returns int64 in go-redis v9)
	incrementedBy, err := rdb.IncrBy(ctx, counterKey, 10).Result()
	if err != nil {
		log.Fatalf("Failed to incrby: %v", err)
	}
	fmt.Printf("INCRBY %s 10: %d\n", counterKey, incrementedBy)

	// DECR - decrement by 1 (returns int64 in go-redis v9)
	decremented, err := rdb.Decr(ctx, counterKey).Result()
	if err != nil {
		log.Fatalf("Failed to decr: %v", err)
	}
	fmt.Printf("DECR %s: %d\n", counterKey, decremented)

	// ========== USAGE OF CLIENT ==========

	// Close the connection when done
	err = rdb.Close()
	if err != nil {
		log.Printf("Failed to close connection: %v", err)
	}
	fmt.Println("Connection closed")
}
