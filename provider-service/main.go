package main

import (
	"context"
	"log"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient/stream"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/delivery"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/service"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("\x1b[31;1mFailed to start service: %v\x1b[0m\n", r)
			log.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	redisConfig := &redisclient.Config{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	}

	rdb, err := redisclient.New(ctx, redisConfig)
	if err != nil {
		log.Fatal(err)
	}

	redisStreamConfig := stream.Config{
		Group:        "provider-service",
		Consumer:     "consumer-1",
		Streams:      []string{"flight.search.requested"},
		Count:        10,
		Block:        5 * time.Second,
		NoAck:        false,
		GroupStartID: "0", // ketika buat group, baca dari awal stream
		ReadStartID:  ">", // baca pesan yang belum di-ack (baru)
	}
	redisStreamConsumer := stream.NewConsumerGroup(&redisStreamConfig, rdb.Client())

	redisStream := stream.NewClient(rdb.Client())

	service := service.NewFlightService()
	usecase := usecase.NewFlightUsecase(redisStream, service)
	redisHandler := delivery.NewRedisHandler(usecase, redisStreamConsumer)

	redisHandler.RegisterSubscribers()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := redisStreamConsumer.StartConsuming(ctx); err != nil {
			log.Fatalf("Redis Stream consumer error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Context canceled, shutting down services...")

	rdb.Close()

	wg.Wait()
	log.Println("App exited")
}
