package main

import (
	"github.com/gorilla/websocket"
	r "gopkg.in/gorethink/gorethink.v4"
	"log"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	id           string
	userName     string
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
}

func NewClient(sock *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
	user := User{
		Name: "Anonymous",
	}
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		id:           id,
		userName:     user.Name,
		send:         make(chan Message),
		socket:       sock,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
	}
}

func (c *Client) NewStopChannel(key int) chan bool {
	c.StopForKey(key)

	stop := make(chan bool)
	c.stopChannels[key] = stop

	return stop
}

func (c *Client) StopForKey(key int) {
	if stop, found := c.stopChannels[key]; found {
		stop <- true
		delete(c.stopChannels, key)
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
	c.Close()
}

func (c *Client) Close() {
	for _, stop := range c.stopChannels {
		stop <- true
	}

	r.Table("user").Get(c.id).Delete().Exec(c.session)

	close(c.send)
	c.socket.Close()

}
