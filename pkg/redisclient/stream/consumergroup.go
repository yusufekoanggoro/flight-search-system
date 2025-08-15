package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type consumerGroup struct {
	config   *Config
	client   redis.UniversalClient
	handlers map[string][]MessageHandlerFunc // one stream to many modules
}

func NewConsumerGroup(config *Config, client redis.UniversalClient) ConsumerGroup {
	return &consumerGroup{
		config:   config,
		client:   client,
		handlers: make(map[string][]MessageHandlerFunc),
	}
}

func (c *consumerGroup) AddMessageHandler(stream string, handler MessageHandlerFunc) {
	c.handlers[stream] = append(c.handlers[stream], handler)
}

func (c *consumerGroup) StartConsuming(ctx context.Context) error {
	for _, stream := range c.config.Streams {
		// startID bisa diisi dengan:
		// "$" → hanya entri baru
		// "0" → mulai dari awal
		// atau ID spesifik seperti "1691234567890-0"
		err := c.client.XGroupCreateMkStream(ctx, stream, c.config.Group, c.config.GroupStartID).Err()
		if err != nil {
			if strings.Contains(err.Error(), "BUSYGROUP") {
				log.Printf("Group %s for stream %s already exists, skipping creation", c.config.Group, stream)
				continue
			}
			return fmt.Errorf("failed to create group %s for stream %s: %w", c.config.Group, stream, err)
		}
		log.Printf("Group %s created for stream %s", c.config.Group, stream)
	}

	streamsAndIDs := []string{}
	for _, stream := range c.config.Streams {
		streamsAndIDs = append(streamsAndIDs, stream, c.config.ReadStartID)
	}

	log.Println("redis stream consumer up and running!...")

	for {
		select {
		case <-ctx.Done():
			log.Println("consumer context canceled, stopping consumer loop")
			return nil
		default:
			// lanjut ke XReadGroup
		}

		streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.config.Group,
			Consumer: c.config.Consumer,
			Streams:  streamsAndIDs,
			// []string{c.config.Stream, c.config.ReadStartID},
			Count: c.config.Count,
			Block: c.config.Block,
		}).Result()
		if err != nil {
			// Bisa karena context canceled atau redis down
			if errors.Is(err, context.Canceled) {
				log.Println("consumer context canceled, stopping consumer loop")
				return nil
			}
			if errors.Is(err, redis.Nil) {
				// Tidak ada pesan baru, bisa dilewati tanpa log error
				continue
			}
			log.Printf("failed to read from stream group: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				raw, err := json.Marshal(message.Values)
				if err != nil {
					log.Printf("failed to encode message values: %v", err)
					continue
				}

				if handlers, ok := c.handlers[stream.Stream]; ok {
					for _, handler := range handlers {
						err := handler([]byte(message.ID), raw)
						if err != nil {
							log.Printf("message %s not acknowledged. Handler error: %v", message.ID, err)
							continue
						}

						if err := c.client.XAck(ctx, stream.Stream, c.config.Group, message.ID).Err(); err != nil {
							log.Printf("failed to acknowledge message %s: %v", message.ID, err)
						} else {
							log.Printf("message %s acknowledged successfully", message.ID)
						}
					}
				}
			}
		}
	}
}

func (c *consumerGroup) HealthCheck() {}

func (c *consumerGroup) RetryFailedMessages() {}
