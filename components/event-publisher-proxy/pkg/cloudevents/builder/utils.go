package builder

func CheckForEmptySegments(segments []string) bool {
	for _, segment := range segments {
		if segment == "" {
			return true
		}
	}
	return false
}
