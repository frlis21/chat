package client

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

const USER_FILE_PATH string = BASE_DATA_PATH + "/userdata"
const ID_EMPTY string = ""
const USER_SEPERATOR string = ":"
const MISSING_USER string = "user_missing"

var currentUser *User = nil

type User struct {
	Name string
	UUID string
}

func NewUser(name, id string) *User {
	if id == ID_EMPTY {
		id = uuid.NewString()
	}
	return &User{name, id}
}

func GetCurrentUser() (*User, error) {
	if currentUser == nil {
		fileData, err := os.ReadFile(USER_FILE_PATH)
		if err != nil {
			return nil, err
		}
		userData := strings.Split(string(fileData), USER_SEPERATOR)
		currentUser = NewUser(userData[0], userData[1])
	}
	return currentUser, nil
}

func SetUser(username string) error {
	currentUser = NewUser(username, ID_EMPTY)
	err := os.WriteFile(USER_FILE_PATH, []byte(currentUser.Name+USER_SEPERATOR+currentUser.UUID), 0777)
	if err != nil {
		return err
	}
	return nil
}
