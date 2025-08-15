package stream

import (
	"context"
	"time"
)

type Config struct {
	Group        string
	Consumer     string
	Streams      []string
	Count        int64
	Block        time.Duration
	NoAck        bool
	GroupStartID string // for XGroupCreate
	ReadStartID  string // for XReadGroup
}

type MessageHandlerFunc func(key, value []byte) error

type ConsumerGroup interface {
	StartConsuming(ctx context.Context) error
	HealthCheck()
	RetryFailedMessages()
	AddMessageHandler(stream string, handler MessageHandlerFunc)
}

type Client interface {
	Add(ctx context.Context, stream string, values interface{}) error
	Ack(ctx context.Context, stream, group, messageID string) error
}
