package client

import (
	"chat/client/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const GROUPS_FILE_PATH string = BASE_DATA_PATH + "/groupdata"

type Group struct {
	Name         string    `json:"name"`
	UUID         string    `json:"id"`
	Antecedent   string    `json:"-"`
	ErrorMessage string    `json:"-"`
	Relays       *[]*Relay `json:"-"`
}

func NewGroup(name string) *Group {
	return &Group{name, uuid.NewString(), "", "", &[]*Relay{}}
}

func ExistingGroup(name, UUID, antecedent string, relays *[]*Relay) *Group {
	return &Group{name, UUID, antecedent, "", relays}
}

func GetGroups() map[string]*Group {
	return make(map[string]*Group)
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
	g.Antecedent = g.UUID
	return g

}

func (g *Group) GetMessages() []*Message {
	messages := make(map[string]*Message)
	for _, relay := range GetRelays() {
		resp, err := http.Get(fmt.Sprintf("http://%v/posts?topic=%v", relay, g.UUID))
		if err != nil {
			fmt.Printf("%v\n", err)
			return []*Message{}
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("%v\n", resp.StatusCode)
			return []*Message{}
		}
		data := utils.ReadFullResponse(resp)
		ms := make([]*Message, 32)
		json.Unmarshal(data, &ms)
		for _, m := range ms {
			_, ok := messages[m.Id]
			if !ok {
				messages[m.Id] = m
			}
		}
	}
	values := utils.Values(messages)
	slices.SortFunc(
		values,
		func(a, b *Message) int {
			return a.Timestamp.Compare(b.Timestamp)
		},
	)
	if len(values) > 0 {
		g.Antecedent = values[len(values)-1].Id
	}
	return values[1:]
}

func (g *Group) SendMessage(m *Message) error {
	requestBody, _ := json.Marshal(m)
	fmt.Printf("%v\n", string(requestBody))
	fmt.Printf("%v\n", g.Antecedent)
	for _, relay := range GetRelays() {
		resp, err := http.Post(
			fmt.Sprintf("http://%v/posts", relay),
			JSON_CONTENT_TYPE,
			strings.NewReader(string(requestBody)),
		)
		if err != nil {
			return err
		}
		fmt.Printf("status_code: %v\n", resp.StatusCode)
		data := utils.ReadFullResponse(resp)
		fmt.Printf("data: %v\n", string(data))
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("failed sending message to %v", relay)
		}
	}
	g.Antecedent = m.Id
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

func (g *Group) JoinGroup() error {
	user, _ := GetCurrentUser()
	requestBody := strings.NewReader(fmt.Sprintf(`{"user_id": "%v", "topics": [%v]}`, user, g))
	for _, relay := range GetRelays() {
		_, err := http.Post(
			fmt.Sprintf("http://%v/posts", relay),
			JSON_CONTENT_TYPE,
			requestBody,
		)
		if err != nil {
			return err
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
