package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

func addChannel(c *Client, data interface{}) {
	var channel Channel

	mapstructure.Decode(data, &channel)

	fmt.Printf("%#v\n", channel)

	c.send <- Message{
		"channel.add",
		Channel{
			"123",
			channel.Name,
		},
	}
}
