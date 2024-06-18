package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Group struct {
	Name         string
	UUID         string
	Antecedent   string
	ErrorMessage string
	Relays       *[]*Relay
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
	return &Group{name, uuid.NewString(), "", "", &[]*Relay{}}
}

func ExistingGroup(name, UUID, antecedent string, relays *[]*Relay) *Group {
	return &Group{name, UUID, antecedent, "", relays}
}

func GetGroups() map[string]*Group {
	return setupGroups()
}

func (g *Group) GetMessages() []*Message {
	if defaultMessages != nil {
		return defaultMessages
	}
	m1 := NewMessage("Message Content 1", NewUser("DEFAULT 1", ID_EMPTY))
	m2 := NewMessage("Message Content 2", NewUser("DEFAULT 2", ID_EMPTY))
	defaultMessages = []*Message{m1, m2}
	return defaultMessages
}

func (g *Group) SendMessage(m *Message) error {
	requestBody := fmt.Sprintf(
		`{"antecedent": "%v", "author": "%v%v%v", "body": "%v", "relays": %v}`,
		g.Antecedent,
		m.SentBy.Name, USER_SEPERATOR, m.SentBy.UUID,
		m.Content,
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					fmt.Sprintf("%v", *g.Relays),
					" ",
					`", "`,
				),
				"[",
				`["`,
			),
			"]",
			`"]`,
		),
	)
	fmt.Printf("%v\n", requestBody)
	for _, relay := range *g.Relays {
		resp, err := http.Post(
			fmt.Sprintf("%v/posts/%v%v%v", relay, g.Name, RELAY_SEPERATOR, g.UUID),
			JSON_CONTENT_TYPE,
			strings.NewReader(requestBody),
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 {
			return fmt.Errorf("failed sending message to %v", relay)
		}
	}
	return nil
}

func (g *Group) String() string {
	return fmt.Sprintf("{\"Name\": \"%v\", \"UUID\": \"%v\"}", g.Name, g.UUID)
}
