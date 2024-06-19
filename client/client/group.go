package client

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const GROUPS_FILE_PATH string = BASE_DATA_PATH + "/groupdata"

type Group struct {
	Name         string
	UUID         string
	Antecedent   string
	ErrorMessage string
	Relays       *[]*Relay
}

var defaultGroups map[string]*Group = nil

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
	messages := make(map[string]*Message)
	for _, relay := range *g.Relays {
		resp, err := http.Get(fmt.Sprintf("http://%v/posts/%v%v%v", relay, g.Name, RELAY_SEPERATOR, g.UUID))
		if err != nil {
			return []*Message{}
		}
		if resp.StatusCode == http.StatusOK {
			return []*Message{}
		}
		data := ReadFullResponse(resp)
		ms := Decode(data)
		for id, msg := range ms {
			_, ok := messages[id]
			if !ok {
				messages[id] = msg
			}
		}
	}
	values := Values(messages)
	slices.SortFunc[[]*Message, *Message](
		values,
		func(a, b *Message) int {
			return a.Timestamp.Compare(b.Timestamp)
		},
	)
	return values
}

func (g *Group) SendMessage(m *Message) error {
	requestBody := fmt.Sprintf(
		`{"antecedent": "%v", "author": "%v%v%v", "body": "%v", "relays": %v}`,
		g.Antecedent,
		m.Author.Name, USER_SEPERATOR, m.Author.UUID,
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
			fmt.Sprintf("http://%v/posts/%v%v%v", relay, g.Name, RELAY_SEPERATOR, g.UUID),
			JSON_CONTENT_TYPE,
			strings.NewReader(requestBody),
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("failed sending message to %v", relay)
		}
	}
	return nil
}

func (g *Group) AddRelay(r *Relay) {
	*g.Relays = append(*g.Relays, r)
}

func (g *Group) String() string {
	return fmt.Sprintf("%v%v%v", g.Name, ":", g.UUID)
}
