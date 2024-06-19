package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	h.Write([]byte(parent))
	h.Write([]byte(user.Name))
	h.Write([]byte(USER_SEPERATOR))
	h.Write([]byte(user.UUID))
	h.Write([]byte(content))
	ts, _ := time.MarshalBinary()
	h.Write(ts)
	id := h.Sum(nil)
	return &Message{
		hex.EncodeToString(id),
		topic,
		parent,
		user,
		time,
		content,
	}
}

func (m *Message) String() string {
	data, _ := json.Marshal(m)
	return string(data)
}
