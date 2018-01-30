package main

import (
	"github.com/mitchellh/mapstructure"
	r "gopkg.in/gorethink/gorethink.v4"
	"time"
)

const (
	ChannelStop = iota
	UserStop
	MessageStop
)

func addChannel(c *Client, data interface{}) {
	var channel Channel

	err := mapstructure.Decode(data, &channel)

	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}

	go add(c, "channel", &channel)
}

func addMessage(c *Client, data interface{}) {
	var message ChannelMessage

	err := mapstructure.Decode(data, &message)

	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}

	message.Author = c.userName
	message.CreatedAt = time.Now()

	go add(c, "message", message)
}

func add(c *Client, table string, resource interface{}) {
	err := r.Table(table).
		Insert(resource).
		Exec(c.session)

	if err != nil {
		c.send <- Message{"error", err.Error()}
	}
}

func editUser(c *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)

	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		err = r.Table("user").
			Get(c.id).
			Update(user).
			Exec(c.session)

		if err != nil {
			c.send <- Message{"error", err.Error()}
		}

		c.userName = user.Name
	}()
}

func subscribeChannel(c *Client, data interface{}) {
	stop := c.NewStopChannel(ChannelStop)

	cursor, err := r.Table("channel").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(c.session)

	if err != nil {
		c.send <- Message{"error", err.Error()}
	}

	subscribe(c, cursor, "channel", stop)
}

func subscribe(c *Client, cursor *r.Cursor, msgPrefix string, stop <-chan bool) {
	result := make(chan r.ChangeResponse)

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				cursor.Close()
				return
			case change := <-result:
				if change.NewValue != nil && change.OldValue == nil {
					c.send <- Message{msgPrefix + ".add", change.NewValue}
				} else if change.NewValue != nil && change.OldValue != nil {
					c.send <- Message{msgPrefix + ".edit", change.NewValue}
				} else if change.NewValue == nil && change.OldValue != nil {
					c.send <- Message{msgPrefix + ".remove", change.OldValue}
				}
			}
		}
	}()
}

func unsubscribeChannel(c *Client, data interface{}) {
	c.StopForKey(ChannelStop)
}

func subscribeUser(c *Client, data interface{}) {
	stop := c.NewStopChannel(UserStop)

	cursor, err := r.Table("user").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(c.session)

	if err != nil {
		c.send <- Message{"error", err.Error()}
	}

	subscribe(c, cursor, "user", stop)
}

func unsubscribeUser(c *Client, data interface{}) {
	c.StopForKey(UserStop)
}

func subscribeMessage(c *Client, data interface{}) {
	var message ChannelMessage
	err := mapstructure.Decode(data, &message)

	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}

	stop := c.NewStopChannel(MessageStop)
	cursor, err := r.Table("message").
		OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
		Filter(r.Row.Field("channelId").Eq(message.ChannelId)).
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(c.session)

	if err != nil {
		c.send <- Message{"error", err.Error()}
	}

	subscribe(c, cursor, "message", stop)
}

func unsubscribeMessage(c *Client, data interface{}) {
	c.StopForKey(MessageStop)
}
