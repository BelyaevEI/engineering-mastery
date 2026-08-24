package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis transactions with WATCH/MULTI/EXEC (optimistic locking)
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	// Clean up from previous runs
	rdb.Del(ctx, "balance:user1", "balance:user2")

	// ========== WATCH/MULTI/EXEC TRANSACTION ==========

	fmt.Println("=== Transaction Example: Transfer Money ===")

	// Initialize balances
	rdb.Set(ctx, "balance:user1", "1000", 0)
	rdb.Set(ctx, "balance:user2", "1000", 0)

	// Transaction function
	transfer := func(amount int64) error {
		// Optimistic lock on both accounts
		err := rdb.Watch(ctx, func(tx *redis.Tx) error {
			// Read current balances
			user1Bal, err := tx.Get(ctx, "balance:user1").Result()
			if err != nil {
				return fmt.Errorf("failed to get user1 balance: %w", err)
			}

			user2Bal, err := tx.Get(ctx, "balance:user2").Result()
			if err != nil {
				return fmt.Errorf("failed to get user2 balance: %w", err)
			}

			// Parse balances
			bal1, _ := strconv.ParseInt(user1Bal, 10, 64)
			_, _ = strconv.ParseInt(user2Bal, 10, 64)

			// Check if user1 has enough money
			if bal1 < amount {
				return fmt.Errorf("insufficient funds: %d < %d", bal1, amount)
			}

			// Prepare transaction
			_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.DecrBy(ctx, "balance:user1", amount)
				pipe.IncrBy(ctx, "balance:user2", amount)
				return nil
			})

			return err
		}, "balance:user1", "balance:user2")

		return err
	}

	// Execute transfer
	err := transfer(100)
	if err != nil {
		log.Printf("Transfer failed: %v", err)
	} else {
		fmt.Println("Transfer successful")
	}

	// Check final balances
	b1, _ := rdb.Get(ctx, "balance:user1").Result()
	b2, _ := rdb.Get(ctx, "balance:user2").Result()
	fmt.Printf("Final balances: user1=%s, user2=%s\n\n", b1, b2)

	// ========== CAS (Check-And-Set) WITH OPTIMISTIC LOCKING ==========

	fmt.Println("=== CAS Example: Increment Counter ===")

	rdb.Del(ctx, "counter")
	err = rdb.Set(ctx, "counter", "0", 0).Err()
	if err != nil {
		log.Fatal(err)
	}

	// Try to increment counter multiple times concurrently
	for i := 0; i < 5; i++ {
		err := rdb.Watch(ctx, func(tx *redis.Tx) error {
			// Get current value
			val, err := tx.Get(ctx, "counter").Result()
			if err != nil {
				return err
			}

			current, _ := strconv.ParseInt(val, 10, 64)
			newVal := current + 1

			// Execute transaction
			_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, "counter", fmt.Sprintf("%d", newVal), 0)
				return nil
			})

			return err
		}, "counter")

		if err != nil {
			log.Printf("Iteration %d failed: %v", i+1, err)
		} else {
			fmt.Printf("Iteration %d: Successfully incremented counter\n", i+1)
		}
	}

	finalCounter, _ := rdb.Get(ctx, "counter").Result()
	fmt.Printf("Final counter value: %s (expected: 5)\n\n", finalCounter)

	// ========== TRANSACTION WITH ERRORS ==========

	fmt.Println("=== Transaction with Discard ===")

	rdb.Del(ctx, "mylist")
	rdb.LPush(ctx, "mylist", "a", "b", "c")

	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		// This will cause an error - trying to pop from empty list
		_, err := tx.LPop(ctx, "mylist").Result()
		if err != nil {
			fmt.Println("Error in transaction, discarding...")
			return err
		}

		// This won't execute due to error above
		_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "discard:test", "should-not-exist", 0)
			return nil
		})
		return err
	}, "mylist")

	if err != nil {
		fmt.Println("Transaction was discarded due to error")
	}

	// Verify no partial execution
	val, _ := rdb.Get(ctx, "discard:test").Result()
	fmt.Printf("discard:test = '%s' (empty = transaction was discarded)\n\n", val)

	// ========== PIPELINE VS TRANSACTION ==========

	fmt.Println("=== Pipeline (batch) vs Transaction ===")

	rdb.Del(ctx, "p:1", "p:2", "p:3")

	// Pipeline - no transaction, just batching
	_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i := 1; i <= 3; i++ {
			pipe.Set(ctx, fmt.Sprintf("p:%d", i), fmt.Sprintf("value%d", i), 0)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pipeline executed successfully")

	// Verify pipeline results
	for i := 1; i <= 3; i++ {
		val, _ := rdb.Get(ctx, fmt.Sprintf("p:%d", i)).Result()
		fmt.Printf("p:%d = %s\n", i, val)
	}

	// Close connection
	rdb.Close()
}
