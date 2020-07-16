package gqlschema

import (
	"errors"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"k8s.io/apimachinery/pkg/types"
)

func MarshalUID(uid types.UID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(string(uid)))
	})
}

func UnmarshalUID(v interface{}) (types.UID, error) {
	in, ok := v.(string)
	if !ok {
		return "", errors.New("invalid UID type, expected string")
	}
	return types.UID(in), nil
}
