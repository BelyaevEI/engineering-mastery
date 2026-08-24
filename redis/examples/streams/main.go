package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Streams - persistent, append-only log structure
// Used for message queues, event sourcing, real-time data processing
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	defer rdb.Close()

	// Clean up
	rdb.Del(ctx, "mystream", "trimtest")

	fmt.Println("=== Redis Streams Basics ===")

	// XADD - Add a message to a stream
	// Syntax: XADD stream * field1 value1 field2 value2 ...
	// * means let Redis generate the ID
	id, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{
			"event":   "user.login",
			"user_id": "123",
			"ip":      "192.168.1.1",
		},
	}).Result()
	if err != nil {
		log.Fatalf("XAdd failed: %v", err)
	}
	fmt.Printf("Added message with ID: %s\n", id)

	// Add another message with auto-generated ID
	id2, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{
			"event":   "user.purchase",
			"user_id": "123",
			"amount":  99.99,
		},
	}).Result()
	fmt.Printf("Added message with ID: %s\n", id2)

	// Add a message with a specific ID (useful for replay)
	id3, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		ID:     "1000-0", // Custom ID (must be unique within stream)
		Values: map[string]interface{}{
			"event":  "custom.event",
			"source": "manual",
		},
	}).Result()
	fmt.Printf("Added message with custom ID: %s\n", id3)

	fmt.Println()

	// ========== XREAD - Read Messages ==========

	fmt.Println("=== XREAD: Reading from Stream ===")

	// Read from the beginning of the stream
	// $ means start from the last message (like tail -f)
	// 0-0 means start from the beginning
	results, err := rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"mystream", "0-0"},
		Count:   10, // Limit number of messages
	}).Result()
	if err != nil {
		log.Fatalf("XRead failed: %v", err)
	}

	for _, stream := range results {
		fmt.Printf("Stream: %s\n", stream.Stream)
		for _, msg := range stream.Messages {
			fmt.Printf("  ID: %s, Values: %v\n", msg.ID, msg.Values)
		}
	}

	fmt.Println()

	// ========== XREAD with BLOCK (Blocking Read) ==========

	fmt.Println("=== XREAD BLOCK: Polling for New Messages ===")

	// In a real app, this would be in a goroutine
	// Block for up to 5 seconds waiting for new messages
	// 0 means block indefinitely
	go func() {
		results, err := rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"mystream", "$"}, // $ means read from newest
			Count:   1,
			Block:   5 * time.Second,
		}).Result()
		if err != nil {
			fmt.Printf("XRead blocked (timeout expected): %v\n", err)
			return
		}
		fmt.Println("New message received (blocking read):")
		for _, stream := range results {
			for _, msg := range stream.Messages {
				fmt.Printf("  ID: %s, Values: %v\n", msg.ID, msg.Values)
			}
		}
	}()

	// Add a new message after starting the read
	time.Sleep(100 * time.Millisecond)
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{
			"event": "new.message",
		},
	})

	fmt.Println()

	// ========== XREADGROUP - Consumer Groups ==========

	fmt.Println("=== XREADGROUP: Consumer Groups ===")

	// Create a consumer group
	// XGROUP CREATE mystream mygroup $ MKSTREAM
	// $ means start reading from the end (new messages only)
	// MKSTREAM creates the stream if it doesn't exist
	err = rdb.XGroupCreate(ctx, "mystream", "mygroup", "$").Err()
	if err != nil {
		// Group might already exist, ignore error
		fmt.Printf("XGroupCreate (may already exist): %v\n", err)
	}

	// Consumer 1 reads from the group
	fmt.Println("Consumer 1 reading messages...")
	entries1, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "mygroup",    // Consumer group name
		Consumer: "consumer-1", // Consumer name
		Streams:  []string{"mystream", ">"}, // > means unread messages only
		Count:    10,
		NoAck:    false, // Don't skip acknowledge (must acknowledge later)
	}).Result()
	if err != nil {
		log.Fatalf("XReadGroup failed: %v", err)
	}

	for _, stream := range entries1 {
		for _, msg := range stream.Messages {
			fmt.Printf("  Consumer 1 received: ID=%s, Values=%v\n", msg.ID, msg.Values)
		}
	}

	// Consumer 2 reads from the same group
	fmt.Println("Consumer 2 reading messages...")
	entries2, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "mygroup",
		Consumer: "consumer-2",
		Streams:  []string{"mystream", ">"},
		Count:    10,
	}).Result()
	if err != nil {
		log.Fatalf("XReadGroup failed: %v", err)
	}

	for _, stream := range entries2 {
		for _, msg := range stream.Messages {
			fmt.Printf("  Consumer 2 received: ID=%s, Values=%v\n", msg.ID, msg.Values)
		}
	}

	// Acknowledge processed messages
	// Messages in Pending Entries List (PEL) are removed after ACK
	if len(entries1) > 0 && len(entries1[0].Messages) > 0 {
		ackID := entries1[0].Messages[0].ID
		ackCount, err := rdb.XAck(ctx, "mystream", "mygroup", ackID).Result()
		if err != nil {
			log.Printf("XAck failed: %v", err)
		} else {
			fmt.Printf("Acknowledged message %s, count: %d\n", ackID, ackCount)
		}
	}

	fmt.Println()

	// ========== XPENDING - Check Pending Messages ==========

	fmt.Println("=== XPENDING: Pending Entries List ===")

	// Show pending messages for consumer group
	pending, err := rdb.XPending(ctx, "mystream", "mygroup").Result()
	if err != nil {
		log.Fatalf("XPending failed: %v", err)
	}
	fmt.Printf("Pending messages: %d\n", pending.Count)

	// ========== XCLAIM - Reclaim Stuck Messages ==========

	fmt.Println("=== XCLAIM: Reclaim Pending Messages ===")

	// Add a new message for claiming demo
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{"claim_test": "value"},
	})

	// Consumer 2 reads the message (gets pending)
	_, err = rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "mygroup",
		Consumer: "consumer-2",
		Streams:  []string{"mystream", ">"},
	}).Result()

	// Consumer 1 claims the pending message from consumer 2
	// Only if idle time > 0 (message has been pending)
	claims, err := rdb.XClaim(ctx, &redis.XClaimArgs{
		Stream:   "mystream",
		Group:    "mygroup",
		Consumer: "consumer-1",
		MinIdle:  0, // Any pending message
		Messages: []string{}, // Empty = claim all pending
	}).Result()
	if err != nil {
		log.Printf("XClaim failed: %v", err)
	} else {
		fmt.Printf("Claimed %d messages\n", len(claims))
		for _, msg := range claims {
			fmt.Printf("  Claimed message ID: %s\n", msg.ID)
		}
	}

	fmt.Println()

	// ========== XDEL - Delete Messages ==========

	fmt.Println("=== XDEL: Deleting Messages ===")

	// Delete a specific message
	idToDelete := id // First message we added
	deleted, err := rdb.XDel(ctx, "mystream", idToDelete).Result()
	if err != nil {
		log.Printf("XDel failed: %v", err)
	} else {
		fmt.Printf("Deleted %d message(s)\n", deleted)
	}

	// ========== XTRIM - Trim Stream Size ==========

	fmt.Println("=== XTRIM: Trimming Stream ===")

	// Add more messages to demonstrate trimming
	for i := 0; i < 5; i++ {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "trimtest",
			Values: map[string]interface{}{"seq": i},
		})
	}

	// Trim to keep only last 3 messages
	// MAXLEN ~ means approximate trimming (faster)
	// MAXLEN = means exact trimming
	trimCount, err := rdb.XTrimMaxLen(ctx, "trimtest", 3).Result()
	if err != nil {
		log.Printf("XTrim failed: %v", err)
	} else {
		fmt.Printf("Trimmed %d message(s), kept last 3\n", trimCount)
	}

	// Show trimmed stream
	trimmed, _ := rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"trimtest", "0-0"},
	}).Result()
	for _, stream := range trimmed {
		fmt.Printf("  Truncated stream contents: %d messages\n", len(stream.Messages))
		for _, msg := range stream.Messages {
			fmt.Printf("    %s: %v\n", msg.ID, msg.Values)
		}
	}

	fmt.Println()

	// ========== CLEANUP ==========

	fmt.Println("=== Cleanup ===")
	rdb.Del(ctx, "mystream", "trimtest")
	fmt.Println("Deleted test streams")
}
