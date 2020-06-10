package gqlschema

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"io"
)

func MarshalPort(port uint32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(string(port)))
	})

}

func UnmarshalPort(v interface{}) (uint32, error) {
	in, ok := v.(int)
	if !ok {
		return 0, errors.New("Invalid RawExtension type, expected int")
	}

	if in < 0 || in > 65535 {
		return 0, errors.New("invalid port value, should be in range <0, 65535>")
	}

	return uint32(in), nil
}
