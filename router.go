package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Router struct {
	rules map[string]Handler
}

type Handler func(*Client, interface{})

func NewRouter() *Router {
	return &Router{
		rules: make(map[string]Handler),
	}
}

func (r *Router) FindHandler(msgName string) (Handler, bool) {
	h, found := r.rules[msgName]

	return h, found
}

func (r *Router) Handle(msgName string, h Handler) {
	r.rules[msgName] = h
}

func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sock, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	c := NewClient(sock, e.FindHandler)
	go c.Write()
	go c.Read()
}
