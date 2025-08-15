package sse

type Client struct {
	ID   string
	Hub  Hub
	Send chan []byte
}
