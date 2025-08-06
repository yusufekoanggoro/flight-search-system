package stream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Consumer interface {
	SubscribeGroup(ctx context.Context, stream, group, consumerID string, handler func(msg redis.XMessage, rawData string) (bool, error), expectedID string) error
}

type consumer struct {
	rootCtx context.Context
	client  *redis.Client
}

func NewConsumer(rootCtx context.Context, client *redis.Client) Consumer {
	return &consumer{
		rootCtx: rootCtx,
		client:  client,
	}
}

func (c *consumer) SubscribeGroup(
	ctx context.Context,
	stream,
	group,
	consumerID string,
	handler func(msg redis.XMessage, rawData string) (bool, error),
	expectedID string,
) error {
	// Create group if not exists (MKSTREAM to create stream if belum ada)
	err := c.client.XGroupCreateMkStream(context.Background(), stream, group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatalf("XGroupCreate failed: %v", err)
	}

	log.Printf("[Redis] Consumer started: group=%s consumerID=%s", group, consumerID)

	var retryCount int

	for {
		// log.Println("[DEBUG] Redis loop running...")
		select {
		case <-ctx.Done():
			log.Printf("[Redis] Stop consuming group %s for %s (context done)", group, expectedID)
			c.deleteConsumer(stream, group, consumerID)
			// c.cleanup(stream, group, consumerID)
			// Cleanup: hapus consumer dari group
			// if err := c.client.XGroupDelConsumer(ctx, stream, group, consumerID).Err(); err != nil {
			// 	log.Printf("[Redis] Failed to delete consumer %s: %v", consumerID, err)
			// }

			// // Hapus group kalau sudah tidak diperlukan lagi
			// if err := c.client.XGroupDestroy(ctx, stream, group).Err(); err != nil {
			// 	log.Printf("[Redis] Failed to destroy group %s: %v", group, err)
			// } else {
			// 	log.Printf("[Redis] Group %s destroyed successfully", group)
			// }
			return ctx.Err()
			// return nil
		default:
		}

		// Blocking read from Redis
		streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumerID,
			Streams:  []string{stream, ">"}, // hanya ambil pesan baru (belum pernah dibaca dalam group).
			Count:    10,                    // redis akan membaca maksimal 10 pesan dari stream
			Block:    1 * time.Second,       // durasi maksimal client akan menunggu pesan baru jika belum ada pesan yang tersedia.
		}).Result()

		if err != nil && err != redis.Nil {
			log.Printf("[Redis] XReadGroup error: %v", err)

			// Jika context sudah dicancel, keluar
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				log.Printf("[Redis] Stop consuming group %s for %s (context canceled in error)", group, expectedID)
				return nil
			}

			// Exponential backoff delay
			retryCount++
			backoffDuration := time.Duration(500*(1<<retryCount)) * time.Millisecond
			if backoffDuration > 10*time.Second {
				backoffDuration = 10 * time.Second
			}
			log.Printf("[Redis] Backing off for %v before retrying... (retry #%d)", backoffDuration, retryCount)
			time.Sleep(backoffDuration)
			continue
		}

		// reset backoff setelah berhasil
		retryCount = 0

		for _, s := range streams {
			for _, message := range s.Messages {
				rawData, ok := message.Values["data"]
				if !ok {
					log.Println("data not found")
					continue
				}

				dataStr, ok := rawData.(string)
				if !ok {
					log.Println("data is not a string")
					continue
				}

				// var result dto.FlightSearchResult
				// if err := json.Unmarshal([]byte(dataStr), &result); err != nil {
				// 	log.Println("unmarshal error:", err)
				// 	continue
				// }

				// if result.SearchID != expectedID {
				// 	log.Printf("Message %s skipped: ID mismatch (expected: %s, got: %s)", message.ID, expectedID, result.SearchID)
				// 	continue // tidak di-ack lompat iterasi
				// }

				shouldAck, err := handler(message, dataStr)
				if err != nil {
					log.Printf("Handler failed for message %s: %v", message.ID, err)
					continue
				}
				// err := handler(message)
				// if err != nil {
				// 	log.Printf("Failed to process message %s: %v", message.ID, err)
				// 	continue
				// }

				// XAck
				if shouldAck {
					if err := c.client.XAck(ctx, stream, group, message.ID).Err(); err != nil {
						log.Printf("Failed to acknowledge message %s: %v", message.ID, err)
					} else {
						fmt.Printf("Message %s acknowledged successfully\n", message.ID)
					}
				}
				// if err := c.client.XAck(ctx, stream, group, message.ID).Err(); err != nil {
				// 	log.Printf("Failed to acknowledge message %s: %v", message.ID, err)
				// } else {
				// 	fmt.Printf("Message %s acknowledged successfully\n", message.ID)
				// }
			}
		}
	}
}

// func (c *consumer) Shutdown() {
// 	close(c.stopCh)
// 	log.Println("Consumer shutdown triggered.")
// }

func (c *consumer) deleteConsumer(stream, group, consumerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.client.XGroupDelConsumer(ctx, stream, group, consumerID).Err(); err != nil {
		log.Printf("[Redis] Failed to delete consumer %s: %v", consumerID, err)
	} else {
		log.Printf("[Redis] Consumer %s removed from group %s", consumerID, group)
	}
}

func (c *consumer) destroyGroup(stream, group string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.client.XGroupDestroy(ctx, stream, group).Err(); err != nil {
		log.Printf("[Redis] Failed to destroy group %s: %v", group, err)
	} else {
		log.Printf("[Redis] Group %s destroyed successfully", group)
	}
}
