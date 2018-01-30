package main

import (
	"net/http"
)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	router := NewRouter()

	router.Handle("channel.add", addChannel)

	http.Handle("/", router)
	http.ListenAndServe(":3001", nil)
}