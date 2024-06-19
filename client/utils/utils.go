package utils

import (
	"io"
	"net/http"
)

func ReadFullResponse(r *http.Response) []byte {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return []byte{}
	}
	return data
}

func Values[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}
