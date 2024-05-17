package client

type Message struct {
	SentBy  *User
	Content string
}

func NewMessage(content string, user *User) *Message {
	return &Message{user, content}
}
