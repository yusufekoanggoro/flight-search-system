package app

import (
	"log"

	"github.com/gofiber/fiber/v2/middleware/cors"
	rest "github.com/yusufekoanggoro/flight-search-system/main-service/delivery/http"
	"github.com/yusufekoanggoro/flight-search-system/main-service/repository"
	"github.com/yusufekoanggoro/flight-search-system/main-service/usecase"
)

func (a *App) ServeHTTP() {
	a.httpServer.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Cache-Control",
	}))

	repo := repository.NewRedisRepository(a.producer, a.consumer)
	usecase := usecase.NewFlightUsecase(repo)
	httpHandler := rest.NewHTTPHandler(a.rootCtx, usecase, a.consumer)

	group := a.httpServer.Group("/api/flights")
	httpHandler.RegisterRoutes(group)

	log.Println("Server is running on :3000")

	if err := a.httpServer.Listen(":3000"); err != nil {
		log.Printf("Fiber app terminated: %v", err)
	}
}
