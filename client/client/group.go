package client

import (
	"chat/client/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

func CreateGroup(name string) *Group {
	g := NewGroup(name)
	user, _ := GetCurrentUser()
	for _, relay := range GetRelays() {
		requestBody, err := json.Marshal(Message{
			g.UUID,
			g.UUID,
			g.UUID,
			user,
			time.Now(),
			g.Name,
		})
		fmt.Printf("body: %v\nerr: %v\n", string(requestBody), err)
		if err != nil {
			return nil
		}
		resp, err := http.Post(
			fmt.Sprintf("http://%v/posts", relay),
			JSON_CONTENT_TYPE,
			strings.NewReader(string(requestBody)),
		)
		if err != nil || resp.StatusCode != http.StatusCreated {
			return nil
		}
	}
	return g

}

func (g *Group) GetMessages() []*Message {
	// messages := make(map[string]*Message)
	// for _, relay := range *g.Relays {
	// 	resp, err := http.Get(fmt.Sprintf("http://%v/posts/%v%v%v", relay, g.Name, RELAY_SEPERATOR, g.UUID))
	// 	if err != nil {
	// 		return []*Message{}
	// 	}
	// 	if resp.StatusCode == http.StatusOK {
	// 		return []*Message{}
	// 	}
	// 	data := utils.ReadFullResponse(resp)
	// 	json.Unmarshal()
	// 	for id, msg := range ms {
	// 		_, ok := messages[id]
	// 		if !ok {
	// 			messages[id] = msg
	// 		}
	// 	}
	// }
	// values := utils.Values(messages)
	// slices.SortFunc[[]*Message, *Message](
	// 	values,
	// 	func(a, b *Message) int {
	// 		return a.Timestamp.Compare(b.Timestamp)
	// 	},
	// )
	// return values
	return []*Message{}
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

func SearchGroups(req *http.Request) []*Group {
	groupName := req.FormValue("group_name")
	foundGroups := make(map[string]*Group)
	for _, relay := range GetRelays() {
		knownGroups := relay.GroupSearch(groupName)
		for _, group := range knownGroups {
			_, ok := foundGroups[group.UUID]
			if !ok {
				foundGroups[group.UUID] = group
			}
		}
	}
	return utils.Values(foundGroups)
}

func (g *Group) AddRelay(r *Relay) {
	*g.Relays = append(*g.Relays, r)
}

func (g *Group) String() string {
	return fmt.Sprintf("%v%v%v", g.Name, ":", g.UUID)
}
