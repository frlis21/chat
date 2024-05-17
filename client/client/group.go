package client

import "github.com/google/uuid"

type Group struct {
	Name    string
	UUID    string
	Members *[]*User
}

var defaultGroups map[string]*Group = nil

func setupGroups() map[string]*Group {
	if defaultGroups != nil {
		return defaultGroups
	}
	g1 := NewGroup("Group Name 1")
	g2 := NewGroup("Group Name 2")
	defaultGroups = map[string]*Group{}
}

func NewGroup(name string) *Group {
	return &Group{name, uuid.NewString(), &[]*User{}}
}

func ExistingGroup(name, UUID string, members *[]*User) *Group {
	return &Group{name, UUID, members}
}

func GetGroups() map[string]*Group {
	return &defaultGroup
}
