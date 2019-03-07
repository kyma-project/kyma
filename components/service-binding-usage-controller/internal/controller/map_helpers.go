package controller

// EnsureMapIsInitiated ensures that given map is initiated.
// - returns given map if it's already allocated
// - otherwise returns empty map
func EnsureMapIsInitiated(m map[string]string) map[string]string {
	if m == nil {
		empty := make(map[string]string)
		return empty
	}

	return m
}

// Merge returns map which is union of input maps. In case of of the same key defined in both maps, ConflictError will be returned
func Merge(m1, m2 map[string]string) (map[string]string, error) {
	out := make(map[string]string)
	for k, v := range m1 {
		out[k] = v
	}

	for k, v := range m2 {
		if _, ex := out[k]; ex {
			return nil, &ConflictError{ConflictingResource: k}
		}
		out[k] = v
	}
	return out, nil
}
