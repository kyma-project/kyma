package gqlschema

import (
	"fmt"
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalAPIRuleAccessStrategy(v v1alpha1.Authenticator) graphql.Marshaler {
	fmt.Printf("marshal %v\n", v)
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte("TODO"))
	})
}

func UnmarshalAPIRuleAccessStrategy(v interface{}) (v1alpha1.Authenticator, error) {
	fmt.Printf("unmarshal %v\n", v)
	return v1alpha1.Authenticator{}, nil
}
