package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

type Labels map[string]string

func MarshalLabels(val Labels) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(val)
		if err != nil {
			panic(err)
		}
	})
}

func UnmarshalLabels(v interface{}) (Labels, error) {
	if m, ok := v.(Labels); ok {
		return m, nil
	}

	return nil, fmt.Errorf("%T is not a map", v)
}

