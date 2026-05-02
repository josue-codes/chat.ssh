package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Time time.Time
	From string // empty = system message
	Text string
}

type Client struct {
	Username string
	Out      chan Message
}

type Hub struct {
	mu      sync.Mutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Join(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	n := len(h.clients)
	h.mu.Unlock()
	h.broadcast(Message{
		Time: time.Now(),
		Text: fmt.Sprintf("* %s joined (%d online)", c.Username, n),
	})
}

func (h *Hub) Leave(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c)
	n := len(h.clients)
	h.mu.Unlock()
	h.broadcast(Message{
		Time: time.Now(),
		Text: fmt.Sprintf("* %s left (%d online)", c.Username, n),
	})
}

func (h *Hub) Send(from, text string) {
	h.broadcast(Message{Time: time.Now(), From: from, Text: text})
}

func (h *Hub) broadcast(m Message) {
	h.mu.Lock()
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.Unlock()
	for _, c := range targets {
		select {
		case c.Out <- m:
		default:
			// Drop the message if the client's buffer is full rather than
			// blocking the broadcast on a slow reader.
		}
	}
}
