package http

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"github.com/yusufekoanggoro/flight-search-system/main-service/dto"
	"github.com/yusufekoanggoro/flight-search-system/main-service/usecase"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
)

type HTTPHandler struct {
	rootCtx  context.Context
	uc       usecase.Usecase
	validate *validator.Validate
	consumer stream.Consumer
}

func NewHTTPHandler(rootCtx context.Context, uc usecase.Usecase, consumer stream.Consumer) *HTTPHandler {
	return &HTTPHandler{
		rootCtx:  rootCtx,
		uc:       uc,
		validate: validator.New(),
		consumer: consumer,
	}
}

func (h *HTTPHandler) RegisterRoutes(group fiber.Router) {
	group.Post("/search", h.submitSearch)
	group.Get("/search/:search_id/stream", h.streamResults)
}

func (h *HTTPHandler) submitSearch(c *fiber.Ctx) error {
	var req dto.FlightSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	resp, err := h.uc.SubmitSearch(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to submit flight search")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Search request submitted",
		"data":    resp,
	})
}

func (h *HTTPHandler) streamResults(c *fiber.Ctx) error {
	prmSearchID := c.Params("search_id")
	if prmSearchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing search_id")
	}

	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	// c.Set("Transfer-Encoding", "chunked") // optional

	ctx, cancel := context.WithCancel(c.Context())

	stream := "flight.search.results"
	group := fmt.Sprintf("main.service.%s", prmSearchID) // 1 Consumer Group per SearchID
	consumerID := "main.service"

	// Subscribe ke Redis Stream duluan, sebelum buat stream writer
	messageChan, err := h.uc.StreamSearchResults(ctx, stream, group, consumerID, prmSearchID)
	if err != nil {
		log.Printf("[SSE] Failed to subscribe for search_id=%s: %v", prmSearchID, err)
		cancel()
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to subscribe")
	}

	// defer cancel()

	// Graceful shutdown dari server
	go func() {
		<-h.rootCtx.Done()
		log.Println("[SSE] Server shutdown signal received")
		cancel()
	}()

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[SSE] Panic recovered: %v", r)
				cancel()
			}
		}()

		fmt.Println("WRITER")
		log.Printf("[SSE] Client connected for search_id=%s", prmSearchID)

		ticker := time.NewTicker(10 * time.Second) // untuk heartbeat tiap 10 detik
		defer ticker.Stop()

		for { // listen Redis message dan mengirim setiap pesan baru ke frontend selama koneksi belum ditutup.
			log.Printf("[SSE] Loop tick for search_id=%s", prmSearchID)
			select {
			case <-ctx.Done():
				log.Printf("[SSE] Client disconnected for search_id=%s", prmSearchID)
				return
			case msg, ok := <-messageChan:
				if !ok {
					log.Printf("[SSE] messageChan closed for search_id=%s", prmSearchID)
					return
				}

				rawData, ok := msg.Values["data"]
				if !ok {
					log.Printf("[SSE] Missing 'data' in Redis message: %+v", msg.Values)
					continue
				}

				data := fmt.Sprintf("%v", rawData)
				fmt.Fprintf(w, "data: %s\n\n", data) // mengirim data ke frontend melalui SSE

				if err := w.Flush(); err != nil {
					// Refreshing page in web browser will establish a new
					// SSE connection, but only (the last) one is alive, so
					// dead connections must be closed here.
					log.Printf("[SSE] Flush error: %v (search_id=%s)", err, prmSearchID)
					cancel() // batalkan context supaya consumer berhenti
					return
				}
			case <-ticker.C:
				// Heartbeat dengan ini, meskipun Redis tidak mengirim data baru, SSE client akan tetap hidup dan tahu bahwa koneksi masih aktif (via : ping).
				fmt.Fprintf(w, ": ping\n\n") // ":" artinya comment di SSE (ignored by browser)
				_ = w.Flush()
			}
		}
	}))

	// cancel()

	return nil
}
