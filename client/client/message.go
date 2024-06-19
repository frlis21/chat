package client

import "time"

type Message struct {
	SentBy    *User
	Timestamp time.Time
	Content   string
}

func NewMessage(content string, time time.Time, user *User) *Message {
	return &Message{user, time, content}
}
