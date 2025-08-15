package delivery

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"github.com/yusufekoanggoro/flight-search-system/main-service/domain"
	shared "github.com/yusufekoanggoro/flight-search-system/main-service/shared/sse"
	"github.com/yusufekoanggoro/flight-search-system/main-service/usecase"
)

type RestHandler struct {
	validate *validator.Validate
	usecase  usecase.Usecase
	hub      shared.Hub
}

func NewRestHandler(usecase usecase.Usecase, hub shared.Hub) *RestHandler {
	return &RestHandler{
		validate: validator.New(),
		usecase:  usecase,
		hub:      hub,
	}
}

func (h *RestHandler) RegisterRoutes(group fiber.Router) {
	group.Post("/search", h.submitSearch)
	group.Get("/search/:search_id/stream", h.streamResults)
}

func (h *RestHandler) submitSearch(c *fiber.Ctx) error {
	var req domain.FlightSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	resp, err := h.usecase.SubmitSearch(c.Context(), &req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to submit flight search")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Search request submitted",
		"data":    resp,
	})
}

func (h *RestHandler) streamResults(c *fiber.Ctx) error {
	searchID := c.Params("search_id")
	if searchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing search_id")
	}

	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	client := &shared.Client{
		ID:   searchID,
		Send: make(chan []byte, 100),
	}

	h.hub.Register(client)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		keepAliveTickler := time.NewTicker(2 * time.Second) // untuk heartbeat tiap 10 detik
		keepAliveMsg := ":keepalive\n"

		defer func() {
			keepAliveTickler.Stop()
			h.hub.Unregister(client)
			log.Println("Exiting stream")
		}()

		for {
			select {
			case msg, ok := <-client.Send:
				if !ok {
					// channel sudah close, keluar loop
					fmt.Println("Channel sudah ditutup, berhenti kirim pesan")
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", msg)
				if err := w.Flush(); err != nil {
					log.Println("Error flushing:", err)
					return
				}

				time.Sleep(2 * time.Second)

			case <-keepAliveTickler.C:
				fmt.Fprintf(w, keepAliveMsg)
				err := w.Flush()
				if err != nil {
					log.Printf("Error while flushing: %v.\n", err)
					return
				}
			}
		}
	}))

	// biasanya kalau nggak ada ticker atau mekanisme kirim data rutin, browser bisa menutup koneksi SSE kalau nggak ada traffic sama sekali dalam jangka waktu tertentu.
	// Penyebabnya:

	// Beberapa proxy, load balancer, atau browser menganggap koneksi “idle” kalau nggak ada data yang lewat, dan otomatis memutuskan.

	// SSE sendiri nggak punya mekanisme “ping” bawaan, jadi kalau tidak ada event, koneksi bisa timeout.

	// Makanya banyak implementasi SSE menambahkan heartbeat/ping, misalnya tiap 15–30 detik kirim komentar SSE (baris diawali :) supaya tidak mengganggu data utama:

	return nil
}
