package main

import (
	"github.com/gorilla/websocket"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	send        chan Message
	socket      *websocket.Conn
	findHandler FindHandler
}

func NewClient(sock *websocket.Conn, findHandler FindHandler) *Client {
	return &Client{
		send:        make(chan Message),
		socket:      sock,
		findHandler: findHandler,
	}
}

func (c *Client) Write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}

func (c *Client) Read() {
	var msg Message
	for {
		if err := c.socket.ReadJSON(&msg); err != nil {
			break
		}
		if handler, found := c.findHandler(msg.Name); found {
			handler(c, msg.Data)
		}
	}
	c.socket.Close()
}
