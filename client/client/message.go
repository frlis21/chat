package client

import "time"

type Message struct {
	Timestamp time.Time
	SentBy    *User
	Content   string
}

func NewMessage(content string, user *User) *Message {
	return &Message{time.Now(), user, content}
}
