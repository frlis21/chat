package client

import (
	"fmt"

	"github.com/google/uuid"
)

type Group struct {
	Name    string
	UUID    string
	Members *[]*User
}

var defaultGroups map[string]*Group = nil
var defaultMessages []*Message = nil

func setupGroups() map[string]*Group {
	if defaultGroups != nil {
		return defaultGroups
	}
	g1 := NewGroup("Group Name 1")
	g2 := NewGroup("Group Name 2")
	defaultGroups = map[string]*Group{g1.UUID: g1, g2.UUID: g2}
	return defaultGroups
}

func NewGroup(name string) *Group {
	return &Group{name, uuid.NewString(), &[]*User{}}
}

func ExistingGroup(name, UUID string, members *[]*User) *Group {
	return &Group{name, UUID, members}
}

func GetGroups() map[string]*Group {
	return setupGroups()
}

func (g *Group) GetMessages() []*Message {
	if defaultMessages != nil {
		return defaultMessages
	}
	m1 := NewMessage("Message Content 1", &defaultUser)
	m2 := NewMessage("Message Content 2", &defaultUser)
	defaultMessages = []*Message{m1, m2}
	return defaultMessages
}

func (g *Group) String() string {
	return fmt.Sprintf("{\"Name\": \"%v\", \"UUID\": \"%v\"}", g.Name, g.UUID)
}
