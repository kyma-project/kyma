// Copied from https://github.com/kubernetes/apimachinery/blob/master/pkg/runtime/converter.go
// TODO: Switch to "k8s.io/apimachinery/pkg/runtime" dependency when we're using kubernetes 1.9 dependency or above (Service Catalog upgrade)

package jsoncopy

import (
	encodingjson "encoding/json"
	"fmt"
)

// DeepCopyJSON deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal().
func DeepCopyJSON(x map[string]interface{}) map[string]interface{} {
	return DeepCopyJSONValue(x).(map[string]interface{})
}

// DeepCopyJSONValue deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal().
func DeepCopyJSONValue(x interface{}) interface{} {
	switch x := x.(type) {
	case map[string]interface{}:
		clone := make(map[string]interface{}, len(x))
		for k, v := range x {
			clone[k] = DeepCopyJSONValue(v)
		}
		return clone
	case []interface{}:
		clone := make([]interface{}, len(x))
		for i, v := range x {
			clone[i] = DeepCopyJSONValue(v)
		}
		return clone
	case string, int64, bool, float64, nil, encodingjson.Number:
		return x
	default:
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}
