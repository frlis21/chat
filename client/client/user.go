package client

import "github.com/google/uuid"

type User struct {
	Name    string
	UUID    string
	Address string
}

var defaultUser = User{"DEFAULT", uuid.NewString(), "127.0.0.1"}

func NewUser(name, address string) *User {
	return &User{name, uuid.NewString(), address}
}
