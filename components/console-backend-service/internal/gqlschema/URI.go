package gqlschema

import (
	"errors"
	"fmt"
	"io"

	"knative.dev/pkg/apis"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalURI(url apis.URL) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(fmt.Sprintf(`"%s"`, url.String())))
	})
}

func UnmarshalURI(v interface{}) (apis.URL, error) {
	in, ok := v.(string)
	if !ok {
		return apis.URL{}, errors.New("invalid URI type, expected string")
	}
	url, err := apis.ParseURL(in)
	if err != nil {
		return apis.URL{}, err
	}
	if url == nil {
		return apis.URL{}, nil
	}
	return *url, nil
}
