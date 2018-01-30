package main

import (
	r "gopkg.in/gorethink/gorethink.v4"
	"log"
	"net/http"
	"time"
)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type ChannelMessage struct {
	Id        string    `json:"id" gorethink:"id,omitempty"`
	ChannelId string    `json:"channelId" gorethink:"channelId"`
	Author    string    `json:"author" gorethink:"author"`
	Body      string    `json:"body" gorethink:"body"`
	CreatedAt time.Time `json:"createdAt" gorethink:"createdAt"`
}

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "rethinkdb:28015",
		Database: "chat",
	})

	if err != nil {
		log.Panic(err.Error())
		return
	}

	r.Table("user").Delete().Exec(session)

	router := NewRouter(session)

	router.Handle("channel.add", addChannel)
	router.Handle("channel.subscribe", subscribeChannel)
	router.Handle("channel.unsubscribe", unsubscribeChannel)
	router.Handle("user.edit", editUser)
	router.Handle("user.subscribe", subscribeUser)
	router.Handle("user.unsubscribe", unsubscribeUser)
	router.Handle("message.add", addMessage)
	router.Handle("message.subscribe", subscribeMessage)
	router.Handle("message.unsubscribe", unsubscribeMessage)

	http.Handle("/", router)
	http.ListenAndServe(":3001", nil)
}
