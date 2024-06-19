package sse

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"context"
)

import "chat/stream"

type Event struct {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	Name  string
	Data  []byte
	ID    string
	Retry uint64
}

func (e *Event) MarshalText() (text []byte, err error) {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "event:%s\n", e.Name)
	fmt.Fprint(buf, "data:")
	data := bytes.ReplaceAll(bytes.TrimSpace(e.Data), []byte("\n"), []byte("\ndata:"))
	fmt.Fprintf(buf, "%s\n", data)
	if e.Retry != 0 {
		fmt.Fprintf(buf, "retry:%d\n", e.Retry)
	}
	fmt.Fprintf(buf, "id:%s\n\n", e.ID)
	return buf.Bytes(), nil
}

func (e *Event) UnmarshalText(text []byte) error {
	for _, line := range bytes.Split(text, []byte("\n")) {
		parsed := bytes.SplitN(line, []byte(":"), 2)
		field := bytes.TrimSpace(parsed[0])
		value := bytes.TrimSpace(parsed[1])
		switch {
		case bytes.Equal(field, []byte("event")):
			e.Name = string(value)
		case bytes.Equal(field, []byte("data")):
			e.Data = append(e.Data, value...)
			e.Data = append(e.Data, []byte("\n")...)
		case bytes.Equal(field, []byte("id")):
			e.ID = string(value)
		case bytes.Equal(field, []byte("retry")):
			retry, err := strconv.ParseUint(string(value), 10, 64)
			if err != nil {
				return err
			}
			e.Retry = retry
		default:
			return fmt.Errorf("Unknown field: %s", field)
		}
	}
	return nil
}

type Reader struct {
	io.Reader
}

// Read events e.g. from a Response.Body
func (r *Reader) ReadEvent() (event Event, err error) {
	var data []byte
	_, err = fmt.Fscanf(r, "%s\n\n", &data)
	if err != nil {
		return event, err
	}
	err = event.UnmarshalText(data)
	return event, err
}

// Why a stream and not just a chan? Answer: inversion of control.
// We need to be able to close channel from here
// or else the entire web server would jam up.
type Server struct {
	events *stream.Stream[Event]
}

func New(events *stream.Stream[Event]) *Server {
	return &Server{events}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// For flushing
	rc := http.NewResponseController(w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := req.Context()
	ch := s.events.Chan(32)
	context.AfterFunc(ctx, func() {
		s.events.Close(ch)
	})

	for e := range ch {
		text, _ := e.MarshalText()
		w.Write(text)
		rc.Flush() // TODO throttle flushing?
	}
}
