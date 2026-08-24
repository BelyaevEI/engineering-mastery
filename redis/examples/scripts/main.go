package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Lua scripting
// Lua scripts are executed atomically - no other command can run in the middle
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	defer rdb.Close()

	// Clean up
	rdb.Del(ctx, "script:test")

	fmt.Println("=== Redis Lua Scripting ===")

	// ========== BASIC SCRIPT EXECUTION ==========

	fmt.Println("\n1. Basic Script with EVAL")

	// Simple script: increment a key and return new value
	script := `
		redis.call('INCR', KEYS[1])
		return redis.call('GET', KEYS[1])
	`

	// Execute script
	result, err := rdb.Eval(ctx, script, []string{"script:test"}).Result()
	if err != nil {
		log.Fatalf("EVAL failed: %v", err)
	}
	fmt.Printf("Result: %v\n", result)

	// Run again
	result, err = rdb.Eval(ctx, script, []string{"script:test"}).Result()
	fmt.Printf("Result (second run): %v\n", result)

	// ========== SCRIPT WITH ARGUMENTS ==========

	fmt.Println("\n2. Script with ARGV (Arguments)")

	// Script that accepts arguments via ARGV
	// ARGV[1] = increment value
	// ARGV[2] = message to log
	multiScript := `
		local current = redis.call('GET', KEYS[1])
		if current == false then
			current = 0
		end
		local new_val = current + tonumber(ARGV[1])
		redis.call('SET', KEYS[1], new_val)
		return {new_val, ARGV[2]}
	`

	result, err = rdb.Eval(ctx, multiScript, []string{"script:test2"}, 5, "Hello").Result()
	if err != nil {
		log.Fatalf("EVAL with args failed: %v", err)
	}
	fmt.Printf("Result: %v (type: %T)\n", result, result)

	// ========== SCRIPT CACHE & SCRIPT LOAD ==========

	fmt.Println("\n3. Script Caching with SCRIPT LOAD")

	// Load script and get SHA1 hash
	sha, err := rdb.ScriptLoad(ctx, multiScript).Result()
	if err != nil {
		log.Fatalf("ScriptLoad failed: %v", err)
	}
	fmt.Printf("Script SHA: %s\n", sha)

	// Execute by SHA (faster, no script body transmission)
	result, err = rdb.EvalSha(ctx, sha, []string{"script:test2"}, 10, "World").Result()
	if err != nil {
		log.Fatalf("EVALSHA failed: %v", err)
	}
	fmt.Printf("Result via EVALSHA: %v\n", result)

	// Check if script exists
	exists, err := rdb.ScriptExists(ctx, sha).Result()
	fmt.Printf("Script exists: %v\n", exists)

	// ========== ATOMIC OPERATIONS: RATE LIMITER ==========

	fmt.Println("\n4. Real-World Example: Rate Limiter")

	rateLimitScript := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])

		-- Get current count
		local current = redis.call('GET', key)
		if current == false then
			current = 0
		end

		-- Check if limit exceeded
		if tonumber(current) >= limit then
			return 0
		end

		-- Increment and set expiration
		current = redis.call('INCR', key)
		if current == 1 then
			redis.call('EXPIRE', key, window)
		end

		return current
	`

	// Test rate limiter
	key := "ratelimit:user:123"
	limit := 3
	window := 10 // seconds

	fmt.Printf("Rate limit: %d requests per %d seconds\n", limit, window)

	for i := 1; i <= 5; i++ {
		result, err := rdb.Eval(ctx, rateLimitScript, []string{key}, limit, window).Result()
		if err != nil {
			log.Fatalf("Rate limit script failed: %v", err)
		}
		if result.(int64) == 0 {
			fmt.Printf("Request %d: REJECTED (rate limit exceeded)\n", i)
		} else {
			fmt.Printf("Request %d: ALLOWED (count: %d)\n", i, result)
		}
	}

	fmt.Println()

	// ========== ATOMIC OPERATIONS: LOCK ==========

	fmt.Println("5. Real-World Example: Distributed Lock")

	lockScript := `
		local key = KEYS[1]
		local token = ARGV[1]
		local ttl = tonumber(ARGV[2])

		-- SETNX: only set if not exists
		local result = redis.call('SET', key, token, 'NX', 'PX', ttl)
		if result == 'OK' then
			return 1
		end
		return 0
	`

	unlockScript := `
		local key = KEYS[1]
		local token = ARGV[1]

		-- Only unlock if token matches (prevents deleting others' locks)
		if redis.call('GET', key) == token then
			return redis.call('DEL', key)
		end
		return 0
	`

	// Try to acquire lock
	token := "my-lock-token-" + fmt.Sprintf("%d", time.Now().UnixNano())
	ttl := 5000 // 5 seconds in milliseconds

	acquired, err := rdb.Eval(ctx, lockScript, []string{"my:lock"}, token, ttl).Result()
	if err != nil {
		log.Fatalf("Lock script failed: %v", err)
	}
	fmt.Printf("Lock acquired: %v\n", acquired == int64(1))

	// Try to acquire another lock (should fail)
	acquired2, err := rdb.Eval(ctx, lockScript, []string{"my:lock"}, "other-token", ttl).Result()
	if err != nil {
		log.Fatalf("Lock script failed: %v", err)
	}
	fmt.Printf("Second lock attempt: %v (should be 0 - locked)\n", acquired2)

	// Release lock
	released, err := rdb.Eval(ctx, unlockScript, []string{"my:lock"}, token).Result()
	if err != nil {
		log.Fatalf("Unlock script failed: %v", err)
	}
	fmt.Printf("Lock released: %v\n", released == int64(1))

	// ========== ATOMIC OPERATIONS: COUNTER WITH CONDITIONAL ==========

	fmt.Println("\n6. Atomic Counter with Conditionals")

	conditionalScript := `
		local key = KEYS[1]
		local increment = tonumber(ARGV[1])
		local min = tonumber(ARGV[2])
		local max = tonumber(ARGV[3])

		local current = redis.call('GET', key)
		if current == false then
			current = 0
		end

		local new_val = tonumber(current) + increment

		-- Apply bounds
		if new_val < min then
			new_val = min
		end
		if new_val > max then
			new_val = max
		end

		redis.call('SET', key, new_val)
		return new_val
	`

	key = "counter:bounded"
	_ = rdb.Del(ctx, key).Err()

	fmt.Printf("Bounded counter (min=0, max=100)\n")

	for _, inc := range []int{-50, 25, 25, 25, -100, 50} {
		result, err := rdb.Eval(ctx, conditionalScript, []string{key}, inc, 0, 100).Result()
		if err != nil {
			log.Fatalf("Conditional script failed: %v", err)
		}
		fmt.Printf("Increment by %d: result=%d\n", inc, result)
	}

	fmt.Println()

	// ========== CLEANUP ==========

	fmt.Println("=== Cleanup ===")
	rdb.Del(ctx, "script:test", "script:test2", "ratelimit:user:123", "my:lock", "counter:bounded")
	fmt.Println("Cleanup completed")
}
