// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"encoding/json"
	binance_connector "github.com/binance/binance-connector-go"
	"strings"
)

// TradeHub maintains the set of active topicClients and broadcasts messages to the
// topicClients.
type TradeHub struct {
	// Registered topicClients.
	topicClients map[string]map[*Client]bool

	// Inbound messages from the topicClients.
	broadcast chan *binance_connector.WsAggTradeEvent

	// Register requests from the topicClients.
	register chan *Client

	// Unregister requests from topicClients.
	unregister chan *Client
	isRun      bool
	coins      map[string]int64
}

func newHub(symbols map[string]int64) *TradeHub {
	h := &TradeHub{
		broadcast:    make(chan *binance_connector.WsAggTradeEvent, 100),
		register:     make(chan *Client, 10),
		unregister:   make(chan *Client, 10),
		topicClients: make(map[string]map[*Client]bool, len(symbols)),
		coins:        symbols,
	}
	for symbol, _ := range symbols {
		h.topicClients[symbol] = make(map[*Client]bool, 5)
	}
	return h
}

func (h *TradeHub) PushMsg(message *binance_connector.WsAggTradeEvent) {
	if h != nil && h.isRun {
		h.broadcast <- message
	}
}
func (h *TradeHub) run() {
	h.isRun = true
	for {
		select {
		case client := <-h.register:
			h.topicClients[client.topic][client] = true
		case client := <-h.unregister:
			if _, ok := h.topicClients[client.topic][client]; ok {
				delete(h.topicClients[client.topic], client)
				close(client.send)
			}
		case message := <-h.broadcast:
			topic := strings.TrimSuffix(message.Symbol, "USDT")
			bs, _ := json.Marshal(message)
			for client := range h.topicClients[topic] {

				select {
				case client.send <- bs:
				default:
					close(client.send)
					delete(h.topicClients[client.topic], client)
				}
			}
		}
	}
}
