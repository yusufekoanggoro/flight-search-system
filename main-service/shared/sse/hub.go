package sse

import (
	"encoding/json"
	"fmt"
	"log"
)

type Hub interface {
	Run()
	Register(client *Client)
	Unregister(client *Client)
	SendToOne(*SSEMessage)
	Close()
}

type SSEMessage struct {
	Event    string      `json:"event"`
	SearchID string      `json:"searchId"`
	Data     interface{} `json:"data"`
}

type hub struct {
	clients    map[string]*Client // key = search_id
	register   chan *Client
	unregister chan *Client
	sendToOne  chan *SSEMessage
	stop       chan struct{}
}

func NewHub() Hub {
	return &hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sendToOne:  make(chan *SSEMessage, 100),
		stop:       make(chan struct{}),
	}
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.register:
			if old, exists := h.clients[client.ID]; exists {
				close(old.Send)
			}
			h.clients[client.ID] = client

		case client := <-h.unregister:
			if c, ok := h.clients[client.ID]; ok {
				close(c.Send)
				delete(h.clients, c.ID)
			}

		case req := <-h.sendToOne:
			if c, ok := h.clients[req.SearchID]; ok {
				select {
				case c.Send <- marshalToBytes(req):
				default:
				}
			}

		case <-h.stop:
			for _, client := range h.clients {
				close(client.Send)
			}
			h.clients = map[string]*Client{}
			return
		}

	}
}

func (h *hub) Register(client *Client) {
	log.Println("register client:", client.ID)
	h.register <- client
}

func (h *hub) Unregister(client *Client) {
	log.Println("unregister client:", client.ID)
	h.unregister <- client
}

func (h *hub) SendToOne(req *SSEMessage) {
	log.Println("send to one:", req.SearchID)
	h.sendToOne <- req
}

func (h *hub) Close() {
	close(h.stop)
}

func marshalToBytes(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshal:", err)
		return []byte("{}")
	}
	return b
}
