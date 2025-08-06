package app

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
)

type App struct {
	rootCtx     context.Context
	httpServer  *fiber.App
	redisClient redisstream.RedisClient
	producer    stream.Producer
	consumer    stream.Consumer
}

func New(rootCtx context.Context, httpServer *fiber.App, redisClient redisstream.RedisClient, producer stream.Producer, consumer stream.Consumer) *App {
	return &App{
		rootCtx:     rootCtx,
		httpServer:  httpServer,
		redisClient: redisClient,
		producer:    producer,
		consumer:    consumer,
	}
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.httpServer.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Fiber server stopped successfully")
	}

	a.redisClient.Shutdown()
}
