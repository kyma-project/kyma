package helpers

import "encoding/json"

func PrettyMarshall(object interface{}) (string, error) {
	out, err := json.MarshalIndent(object, "", "  ")
	return string(out), err
}
