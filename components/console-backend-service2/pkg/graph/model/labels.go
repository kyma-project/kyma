package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

func MarshalLabels(val map[string]string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(val)
		if err != nil {
			panic(err)
		}
	})
}

func UnmarshalLabels(v interface{}) (map[string]string, error) {
	if m, ok := v.(map[string]interface{}); ok {
		result := make(map[string]string, len(m))
		for k, v := range m {
			result[k], ok = v.(string)
			if !ok {
				return nil, fmt.Errorf("%T is not a map", v)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("%T is not a map", v)
}
