package client

import (
	"io"
	"net/http"
	"strings"
)

func ReadFullResponse(r *http.Response) string {
	var data strings.Builder
	buffer := make([]byte, 1024)
	for n, err := r.Body.Read(buffer); err != io.EOF; {
		for i := 0; i < n; i++ {
			data.WriteByte(buffer[i])
			buffer[i] = 0
		}
	}

	return data.String()
}

func Decode(data string) map[string]*Message {
	return make(map[string]*Message)
}

func Values[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}
