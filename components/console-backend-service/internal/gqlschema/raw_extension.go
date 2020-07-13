package gqlschema

import (
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func MarshalRawExtension(e runtime.RawExtension) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write(e.Raw)
	})
}

func UnmarshalRawExtension(v interface{}) (runtime.RawExtension, error) {
	in, ok := v.(string)
	if !ok {
		return runtime.RawExtension{}, errors.New("Invalid RawExtension type, expected string")
	}

	return runtime.RawExtension{Raw: []byte(in)}, nil
}
