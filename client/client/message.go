package client

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Message struct {
	Id        string    `json:"id"`
	Group     string    `json:"topic"`
	Parent    string    `json:"parent"`
	Author    *User     `json:"author"`
	Timestamp time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

func NewMessage(topic, parent, content string, time time.Time, user *User) *Message {
	h := sha256.New()
	h.Write([]byte(topic))
	h.Write([]byte(parent))
	h.Write([]byte(user.Name))
	h.Write([]byte(USER_SEPERATOR))
	h.Write([]byte(user.UUID))
	h.Write([]byte(content))
	id := make([]byte, 0, 32)
	h.Sum(id)
	sha256.Sum256([]byte(fmt.Sprintf("%v%v%v%v%v%v", topic, parent, user.Name, USER_SEPERATOR, user.UUID, content)))
	return &Message{
		hex.Dump(id),
		topic,
		parent,
		user,
		time,
		content,
	}
}
