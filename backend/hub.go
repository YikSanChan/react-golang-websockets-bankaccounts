// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type TopicMessage struct {
	message []byte
	topic string
}

type TopicClient struct {
	client *Client
	topic string
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan TopicMessage

	// Register requests from the clients.
	register chan *TopicClient

	// Unregister requests from clients.
	unregister chan *TopicClient
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan TopicMessage),
		register:   make(chan *TopicClient),
		unregister: make(chan *TopicClient),
		clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case topicClient := <-h.register:
			clients := h.clients[topicClient.topic]
			if clients == nil {
				clients = make(map[*Client]bool)
				h.clients[topicClient.topic] = clients
			}
			h.clients[topicClient.topic][topicClient.client] = true
		case topicClient := <-h.unregister:
			clients := h.clients[topicClient.topic]
			if clients != nil {
				if _, ok := clients[topicClient.client]; ok {
					delete(clients, topicClient.client)
					close(topicClient.client.send)
				}
			}
		case topicMessage := <-h.broadcast:
			clients := h.clients[topicMessage.topic]
			for client := range clients {
				select {
				case client.send <- topicMessage.message:
				default:
					close(client.send)
					delete(clients, client)
					// TODO: In https://github.com/gorilla/websocket/issues/46#issuecomment-227906715 it deletes when len(clients) == 0
				}
			}
		}
	}
}
