package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Pub/Sub pattern
func main() {
	// Create separate clients for publisher and subscriber
	pubClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	subClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	// Clean up
	defer pubClient.Close()
	defer subClient.Close()

	// ========== BASIC PUB/SUB ==========

	fmt.Println("=== Basic Pub/Sub Example ===")

	// Create a subscriber
	sub := subClient.Subscribe(ctx, "chat:room1")
	defer sub.Close()

	// Wait a bit for subscription
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Subscribed to 'chat:room1'")

	// Publisher sends messages
	go func() {
		time.Sleep(100 * time.Millisecond) // Give subscriber time to subscribe

		messages := []string{
			"Hello everyone!",
			"How are you?",
			"Redis Pub/Sub is great!",
		}

		for _, msg := range messages {
			err := pubClient.Publish(ctx, "chat:room1", msg).Err()
			if err != nil {
				log.Printf("Failed to publish: %v", err)
			} else {
				fmt.Printf("Published: '%s'\n", msg)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Consume messages
	for i := 0; i < 3; i++ {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Failed to receive: %v", err)
			continue
		}
		fmt.Printf("Received from %s: %s\n", msg.Channel, msg.Payload)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println()

	// ========== MULTI CHANNEL SUBSCRIPTION ==========

	fmt.Println("=== Multiple Channel Subscription ===")

	// Subscribe to multiple channels
	subMulti := subClient.Subscribe(ctx, "news:sports", "news:tech")
	defer subMulti.Close()

	// Wait a bit for subscription
	time.Sleep(100 * time.Millisecond)

	// Publish to both channels
	go func() {
		time.Sleep(100 * time.Millisecond)

		// Use pipeline for batching publishes
		pipe := pubClient.Pipeline()
		pipe.Publish(ctx, "news:sports", "Team A won the match!")
		pipe.Publish(ctx, "news:tech", "New AI model released!")
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Pipeline error: %v", err)
		}
		fmt.Println("Published to both channels")
	}()

	// Receive messages
	for i := 0; i < 2; i++ {
		msg, err := subMulti.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Receive error: %v", err)
			continue
		}
		fmt.Printf("Received: Channel=%s, Message=%s\n", msg.Channel, msg.Payload)
	}

	fmt.Println()

	// ========== CONCURRENT CONSUMERS ==========

	fmt.Println("=== Concurrent Consumers (Load Balancing) ===")

	// In Redis Pub/Sub, all subscribers get all messages (broadcast)
	// For load balancing, use Redis Streams (see streams.go example)

	var wg sync.WaitGroup
	numConsumers := 3
	received := make([]int, numConsumers)

	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()

			consumer := subClient.Subscribe(ctx, "broadcast:channel")
			defer consumer.Close()

			// Give time for all to subscribe
			time.Sleep(50 * time.Millisecond)

			// Each consumer receives ALL messages
			msg, err := consumer.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("Consumer %d error: %v", consumerID, err)
				return
			}
			received[consumerID] = 1 // Just count that we received something
			fmt.Printf("Consumer %d received: %s\n", consumerID, msg.Payload)
		}(i)
	}

	// Publish one message
	time.Sleep(100 * time.Millisecond)
	err := pubClient.Publish(ctx, "broadcast:channel", "Broadcast message").Err()
	if err != nil {
		log.Printf("Publish error: %v", err)
	}

	wg.Wait()
	fmt.Printf("All consumers received (broadcast behavior): %v\n\n", received)

	// ========== PUB/SUB WITH PATTERN MATCHING ==========

	fmt.Println("=== Pattern-based Subscription ===")

	// Subscribe to all channels matching pattern
	subPattern := subClient.Subscribe(ctx, "logs:*")
	defer subPattern.Close()

	// Wait a bit for subscription
	time.Sleep(100 * time.Millisecond)

	// Publish to different log channels
	logChannels := []string{"logs:info", "logs:warn", "logs:error"}
	messages := []string{"Info message", "Warning message", "Error message"}

	go func() {
		time.Sleep(100 * time.Millisecond)
		for i, channel := range logChannels {
			err := pubClient.Publish(ctx, channel, messages[i]).Err()
			if err != nil {
				log.Printf("Publish error: %v", err)
			} else {
				fmt.Printf("Published to %s\n", channel)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Receive all messages
	for i := 0; i < 3; i++ {
		msg, err := subPattern.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Receive error: %v", err)
			continue
		}
		fmt.Printf("Received via pattern: Channel=%s, Message=%s\n", msg.Channel, msg.Payload)
	}

	fmt.Println()

	// ========== PING/PONG HEARTBEAT ==========

	fmt.Println("=== Heartbeat with Ping/Pong ===")

	subHeartbeat := subClient.Subscribe(ctx, "heartbeat")
	defer subHeartbeat.Close()

	// Consumer side - receive messages
	go func() {
		for {
			msg, err := subHeartbeat.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			fmt.Printf("Received heartbeat: %s\n", msg.Payload)
		}
	}()

	// Publisher side - send periodic heartbeats
	go func() {
		for i := 0; i < 3; i++ {
			err := pubClient.Publish(ctx, "heartbeat", fmt.Sprintf("beat-%d", i)).Err()
			if err != nil {
				log.Printf("Publish error: %v", err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	fmt.Println()

	// ========== UNSUBSCRIBE ==========

	fmt.Println("=== Unsubscribe ===")

	subUnsub := subClient.Subscribe(ctx, "unsub:channel")
	defer subUnsub.Close()

	// Unsubscribe from the channel
	subUnsub.Unsubscribe(ctx, "unsub:channel")
	fmt.Println("Unsubscribed from unsub:channel")

	// Try to receive (should timeout or error)
	subUnsub.Ping(ctx) // Test connection still works
	fmt.Println("Connection still valid after unsubscribe\n")

	// ========== CLEANUP ==========

	fmt.Println("All Pub/Sub examples completed!")
}
