package name

func EmptyIfNil(input *string) string {
	if input == nil {
		return ""
	}
	return *input
}
