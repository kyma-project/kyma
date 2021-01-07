package utils

//MergeMaps merges all values from overridesMap map into baseMap, adding and/or overwriting final keys (string values) if both maps contain such entries.
//baseMap WILL be modified during merge.
//overridesMap won't be modified by future merges, since a deep-copy of it's nested maps are used for merging such nested maps.
func MergeMaps(baseMap, overridesMap map[string]interface{}) {

	if (overridesMap) == nil {
		return
	}

	//Helper function to deep-copy nested maps
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

func putValueToMap(baseMap map[string]interface{}, key string, overrideVal interface{}) {
	overrideValMap, overrideIsMap := overrideVal.(map[string]interface{})
	if overrideIsMap {
		baseMap[key] = deepCopyMap(overrideValMap)
	} else {
		baseMap[key] = overrideVal
	}
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
