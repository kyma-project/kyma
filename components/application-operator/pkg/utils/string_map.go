package utils

const (
	Separator     string = "."
	RuneSeparator        = '.'
)

type InterfaceMap map[string]interface{}
type StringMap map[string]string

func NewStringMap(source InterfaceMap) StringMap {
	releaseMap := make(StringMap)
	releaseMap.flatten("", source)
	return releaseMap
}

func (m StringMap) ContainsAll(labels StringMap) bool {
	if len(m) != len(labels) {
		return false
	}

	for key, applicationValue := range labels {
		if releaseVal, ok := m[key]; !ok || applicationValue != releaseVal {
			return false
		}
	}

	return true
}

func (m StringMap) flatten(prefix string, source map[string]interface{}) {
	for key, value := range source {
		newKey := prefix + key

		if casted, ok := value.(map[string]interface{}); ok {
			m.flatten(newKey+Separator, casted)
		} else {
			m[newKey] = value.(string)
		}
	}
}
