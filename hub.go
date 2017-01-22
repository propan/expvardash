package main

import "github.com/gorilla/websocket"

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	dataCh chan []byte
}

type Hub struct {
	clients map[*Client]struct{}
	dataCh  chan []byte
	enterCh chan *Client
	leaveCh chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		dataCh:  make(chan []byte),
		enterCh: make(chan *Client),
		leaveCh: make(chan *Client),
	}
}

func (h *Hub) Start() {
	for {
		select {
		case client := <-h.enterCh:
			h.clients[client] = struct{}{}
		case client := <-h.leaveCh:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.dataCh)
			}
		case message := <-h.dataCh:
			for client := range h.clients {
				select {
				case client.dataCh <- message:
				default:
					close(client.dataCh)
					delete(h.clients, client)
				}
			}
		}
	}
}
