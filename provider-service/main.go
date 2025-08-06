package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yusufekoanggoro/flight-search-system/main-service/app"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
)

const serviceName = "provider-service"

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		stop()
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mFailed to start %s service: %v\x1b[0m\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	config := redisstream.Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	redisClient, err := redisstream.NewRedisClient(rootCtx, config)
	if err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	producer := stream.NewProducer(redisClient.Client())
	consumer := stream.NewConsumer(rootCtx, redisClient.Client())

	fiberApp := fiber.New()

	app := app.New(rootCtx, fiberApp, redisClient, producer, consumer)

	go func() {
		app.ServeHTTP()
	}()

	<-rootCtx.Done()
	log.Println("Termination signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.Shutdown(shutdownCtx)

	log.Println("Cleanup finished. Exiting.")
}
