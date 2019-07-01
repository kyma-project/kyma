package overrides

import (
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

//Map is a map of overrides. Values in the map can be nested maps (of the same type) or strings
type Map map[string]interface{}

//ToMap converts yaml to Map. Supports only map-like yamls (no lists!)
func ToMap(value string) (Map, error) {
	target := map[string]interface{}{}

	if value == "" {
		//Otherwise, nil Map is returned by yaml.Unmarshal
		return target, nil
	}

	err := yaml.Unmarshal([]byte(value), &target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

//ToYaml converts Map to YAML
func ToYaml(oMap Map) (string, error) {
	if len(oMap) == 0 {
		return "", nil
	}

	res, err := yaml.Marshal(oMap)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

//FlattenMap flattens an Map into a map of aggregated keys and value (to entries like, for example: "istio.ingress.service.gateway: xyz")
func FlattenMap(oMap Map) map[string]string {
	res := map[string]string{}
	flattenMap(oMap, "", res)
	return res
}

//UnflattenToMap converts external "flat" overrides into Map. Opposite of FlattenMap function.
func UnflattenToMap(sourceMap map[string]string) Map {
	mergedMap := map[string]interface{}{}
	if len(sourceMap) == 0 {
		return mergedMap
	}

	for key, value := range sourceMap {
		keys := strings.Split(key, ".")
		mergeIntoMap(keys, value, mergedMap)
	}

	return mergedMap
}

//MergeMaps merges all values from overridesMap map into baseMap, adding and/or overwriting final keys (string values) if both maps contain such entries.
//baseMap WILL be modified during merge.
//overridesMap won't be modified by future merges, since a deep-copy of it's nested maps are used for merging such nested maps.
func MergeMaps(baseMap, overridesMap Map) {

	if (overridesMap) == nil {
		return
	}

	//Helper function to deep-copy nested maps
	putValueToMap := func(baseMap map[string]interface{}, key string, overrideVal interface{}) {
		overrideValMap, overrideIsMap := overrideVal.(map[string]interface{})
		if overrideIsMap {
			baseMap[key] = deepCopyMap(overrideValMap)
		} else {
			baseMap[key] = overrideVal
		}
	}

	for key, overrideVal := range overridesMap {
		//Can be nil
		baseVal := baseMap[key]
		baseMapVal, baseIsMap := baseVal.(map[string]interface{})
		ovrrMapVal, newIsMap := overrideVal.(map[string]interface{})

		if baseIsMap && newIsMap {
			//Two maps case! Reccursion happens here!
			MergeMaps(baseMapVal, ovrrMapVal)
		} else {
			//All other cases, even "pathological" one, when baseMap[key] is a map and overrideVal is a string.
			putValueToMap(baseMap, key, overrideVal)
		}
	}
}

//FindOverrideStringValue looks for a string value assigned to the provided flat key.
func FindOverrideStringValue(overrides Map, flatName string) (string, bool) {

	res, isString := FindOverrideValue(overrides, flatName).(string)
	if isString {
		return res, true
	}
	return "", false
}

//FindOverrideValue looks for a value assigned to the provided flat key.
func FindOverrideValue(overrides Map, flatName string) interface{} {
	var findOverride func(m map[string]interface{}, keys []string) interface{}

	findOverride = func(m map[string]interface{}, keys []string) interface{} {
		if len(keys) == 1 {
			return m[keys[0]]
		}

		nestedMap, isMap := m[keys[0]].(map[string]interface{})
		if isMap {
			return findOverride(nestedMap, keys[1:])
		}

		return nil
	}

	keys := strings.Split(flatName, ".")
	return findOverride(overrides, keys)

}

//Recursively copies the map. Used to ensure immutability of input maps when merging.
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	dst := map[string]interface{}{}

	//Helper recursive function
	var cp func(src map[string]interface{}, dst map[string]interface{})

	cp = func(src map[string]interface{}, dst map[string]interface{}) {
		for key, value := range src {

			nestedMap, isMap := value.(map[string]interface{})
			if isMap {
				//Nested map!
				nestedCopy := map[string]interface{}{}
				cp(nestedMap, nestedCopy)
				dst[key] = nestedCopy
			} else {
				dst[key] = value
			}
		}
	}
	cp(src, dst)

	return dst
}

// Flattens given Map. The keys in result map will contain all intermediate keys joined with dots, e.g.: "istio.ingress.service.gateway: xyz"
func flattenMap(oMap Map, keys string, result map[string]string) {

	var prefix string

	if len(keys) == 0 {
		prefix = ""
	} else {
		prefix = keys + "."
	}

	for key, value := range oMap {

		aString, isString := value.(string)
		if isString {
			result[prefix+key] = aString
		} else {
			//Nested map!
			nestedMap := value.(map[string]interface{})
			flattenMap(nestedMap, prefix+key, result)
		}
	}
}

//Merges value into given map, introducing intermediate "nested" maps for every intermediate key.
func mergeIntoMap(keys []string, value string, dstMap map[string]interface{}) {
	currentKey := keys[0]
	//Last key points directly to string value
	if len(keys) == 1 {

		//Conversion to boolean to satisfy Helm requirements.yaml: "enable:true/false syntax"
		var vv interface{} = value
		if value == "true" || value == "false" {
			vv, _ = strconv.ParseBool(value)
		}

		dstMap[currentKey] = vv
		return
	}

	//All keys but the last one should point to a nested map
	nestedMap, isMap := dstMap[currentKey].(map[string]interface{})

	if !isMap {
		nestedMap = map[string]interface{}{}
		dstMap[currentKey] = nestedMap
	}

	mergeIntoMap(keys[1:], value, nestedMap)
}
