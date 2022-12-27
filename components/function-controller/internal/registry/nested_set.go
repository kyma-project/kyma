package registry

type NestedSet map[string]map[string]struct{}

func NewNestedSet() NestedSet {
	return NestedSet{}
}

func (f NestedSet) AddKeyWithValue(key, value string) {
	if _, ok := f[key]; !ok {
		f[key] = map[string]struct{}{}
	}
	f[key][value] = struct{}{}
}

func (f NestedSet) HasKeyWithValue(key, value string) bool {
	values, hasKey := f[key]
	if !hasKey {
		return false
	}
	_, hasValue := values[value]
	return hasValue
}

func (f NestedSet) HasKey(key string) bool {
	_, hasKey := f[key]
	return hasKey
}

func (f NestedSet) ListKeys() []string {
	keys := []string{}
	for k := range f {
		keys = append(keys, k)
	}
	return keys
}

func (f NestedSet) GetKey(key string) map[string]struct{} {
	values, hasKey := f[key]
	if !hasKey {
		return nil
	}
	return values
}
