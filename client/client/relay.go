package client

import (
	"chat/client/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const RELAY_FILE_PATH string = BASE_DATA_PATH + "/relaydata"
const RELAY_SEPERATOR string = ":"
const MISSING_RELAY string = "user_missing"

type Relay struct {
	Address string
	Port    int
}

var relays []*Relay = make([]*Relay, 0, 10)

func NewRelay(address string, port int) *Relay {
	return &Relay{address, port}
}

func AddRelay(address string, port int) error {
	file, err := os.OpenFile(RELAY_FILE_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %v\n", RELAY_FILE_PATH)
		return err
	}
	_, err = file.WriteString(fmt.Sprintf("%v%v%v\n", address, RELAY_SEPERATOR, port))
	if err != nil {
		fmt.Printf("Failed Writing to %v\n", RELAY_FILE_PATH)
		return err
	}
	err = file.Close()
	if err != nil {
		fmt.Printf("Closing file %v gave an error\n", RELAY_FILE_PATH)
		return err
	}
	relays = append(relays, NewRelay(address, port))
	return nil
}

func GetRelays() []*Relay {
	data, err := os.ReadFile(RELAY_FILE_PATH)
	if err != nil {
		return relays
	}
	rs := strings.Split(string(data), "\n")
	savedRelays := make([]*Relay, len(rs)-1)
	for i, r := range rs[:len(rs)-1] {
		relay := strings.Split(r, RELAY_SEPERATOR)
		address := relay[0]
		port, _ := strconv.Atoi(relay[1])
		savedRelays[i] = NewRelay(address, port)
	}
	relays = savedRelays
	return relays
}

func (r *Relay) GroupSearch(name string) []*Group {
	resp, err := http.Get(fmt.Sprintf("http://%v/topics", r))
	if err != nil {
		fmt.Printf("%v: err: %v\n", r, err)
		return []*Group{}
	}
	foundGroups := make([]*Group, 0, 10)
	err = json.Unmarshal(utils.ReadFullResponse(resp), &foundGroups)
	if err != nil {
		return []*Group{}
	}
	// g := foundGroups[0]
	// fmt.Printf("%v, %v, %v, %v\n", g.Name, g.UUID, g.Antecedent, g.ErrorMessage)
	// fmt.Printf("status code: %v\ndata: %v\n", resp.StatusCode, foundGroups)
	return foundGroups
}

func (r *Relay) String() string {
	return fmt.Sprintf("%v%v%v", r.Address, RELAY_SEPERATOR, r.Port)
}
