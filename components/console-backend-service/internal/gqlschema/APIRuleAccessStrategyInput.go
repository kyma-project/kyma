package gqlschema

import (
	"encoding/json"

	"github.com/ory/oathkeeper-maester/api/v1alpha1"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalAPIRuleAccessStrategyInput(v v1alpha1.Authenticator) graphql.Marshaler {
	panic("Not used") // no point marshalling the input type
}

func UnmarshalAPIRuleAccessStrategyInput(v interface{}) (v1alpha1.Authenticator, error) {
	var it = v1alpha1.Authenticator{Handler: &v1alpha1.Handler{}}
	var asMap = v.(map[string]interface{})
	for k, v := range asMap {
		switch k {
		case "name":
			name, ok := v.(string)
			if !ok {
				return it, errors.New("Invalid 'name' type, expected string")
			}
			it.Name = name
		case "config":
			if v == nil {
				it.Config = nil
				break
			}
			raw, err := json.Marshal(v)
			if err != nil {
				return it, err
			}
			config, err := UnmarshalRawExtension(string(raw))
			if err != nil {
				return it, err
			}
			it.Config = &config
		}
	}
	return it, nil
}
