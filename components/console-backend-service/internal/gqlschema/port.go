package gqlschema

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalPort(port uint32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(fmt.Sprintf("%d", port)))
	})

}

func UnmarshalPort(v interface{}) (uint32, error) {
	var in int64
	switch v.(type) {
	case int, int64:
		in, _ = v.(int64)
	case json.Number:
		var err error
		in, err = v.(json.Number).Int64()
		if err != nil {
			return 0, errors.New("Invalid Port type, expected int")
		}
	default:
		return 0, errors.New("Invalid Port type, expected int")
	}

	if in < 0 || in > 65535 {
		return 0, errors.New("invalid port value, should be in range <0, 65535>")
	}

	return uint32(in), nil
}
