// Interactive client CLI

package main

import "os"
import "io"
import "fmt"
import "log"
import "time"
import "bytes"
import "strings"
import "net/url"
import "net/http"
import "encoding/json"
import "golang.org/x/term"
import "github.com/google/uuid"

type Post struct {
	Author, Body string
}

//type Relay struct {
//	Conn net.Conn
//	Enc  *json.Encoder
//	Dec  *json.Decoder
//}

//func writeln(line string) {
//	t.Write([]byte(line + "\n"))
//}

//func search(args string) {
//	what, name, _ := strings.Cut(args, " ")
//	switch what {
//	case "user":
//		searchUser(name)
//	case "group":
//		searchGroup(name)
//	default:
//		log.Printf("Unknown parameter: %s", what)
//	}
//}

func main() {
	// Set up console
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Println("")
	}()

	q := make(chan string)
	t := term.NewTerminal(os.Stdin, "> ")
	log.SetOutput(t)

	client := http.Client{Timeout: 5 * time.Second}
	topics := map[string]string{}
	relays := map[string]struct{}{}

	currentTopic := ""

	// For concurrent writing
	go func(ch chan string) {
		for msg := range ch {
			fmt.Fprint(t, msg)
		}
	}(q)

	for {
		input, err := t.ReadLine()
		if err == io.EOF {
			// ^C or ^D
			break
		} else if err != nil {
			panic(err)
		}
		trimmed := strings.TrimSpace(input)
		after, found := strings.CutPrefix(trimmed, "/")
		if found {
			cmd, args, _ := strings.Cut(after, " ")
			switch cmd {
			case "connect": // <address>
				relays[args] = struct{}{}
			case "add": // <UUID> <short name>
				// Add a topic
				tokens := strings.SplitN(args, " ", 2)
				topics[tokens[1]] = tokens[0]
			case "remove":
				// Remove a topic
			case "join": // <short name>
				// Leave current group and join another
				topic, found := topics[args]
				if !found {
					log.Printf("Unknown topic: %s", args)
					continue
				}
				currentTopic = topic
				t.SetPrompt(args + "> ")
				for relay := range relays {
					joined, err := url.JoinPath(relay, "messages", topic)
					if err != nil {
						log.Print(err)
						continue
					}
					go func() {
						resp, err := client.Get(joined)
						if err != nil {
							log.Print(err)
							return
						}
						var msg Post
						dec := json.NewDecoder(resp.Body)
						for dec.More() {
							err = dec.Decode(&msg)
							if err != nil {
								log.Print(err)
								return
							}
							q <- fmt.Sprintf("%s: %s", msg.Author, msg.Body)
						}
					}()
				}
			case "create": // <short name>
				topics[args] = uuid.New().String()
				log.Printf("Created topic %s -- %s\n", args, topics[args])
			case "leave":
				// Not really necessary
				//leaveGroup(args)
			case "search":
				//search(args)
			default:
				log.Printf("Unknown command: %s", trimmed)
			}
		} else {
			for relay, _ := range relays {
				joined, err := url.JoinPath(relay, "messages", currentTopic)
				data, err := json.Marshal(Post{"snoobles", trimmed})
				if err != nil {
					log.Print(err)
					continue
				}
				go client.Post(joined, "application/json", bytes.NewReader(data))
			}
		}
	}
}
