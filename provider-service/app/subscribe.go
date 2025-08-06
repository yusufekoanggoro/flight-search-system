package app

import (
	"log"

	"github.com/yusufekoanggoro/flight-search-system/provider-service/delivery/redisstream"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/repository"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/usecase"
)

func (a *App) Subscribe() {
	repo := repository.NewRedisRepository(a.producer, a.consumer)
	usecase := usecase.NewUsecase(repo)
	streamHandler := redisstream.NewConsumerHandler(a.rootCtx, usecase, a.consumer)

	// group := a.httpServer.Group("/api/flights")
	streamHandler.RegisterSubscribers(a.rootCtx)

	log.Println("Server is running on :3000")

	if err := a.httpServer.Listen(":3000"); err != nil {
		log.Printf("Fiber app terminated: %v", err)
	}
}
