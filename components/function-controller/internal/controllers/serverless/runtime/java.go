package runtime

type java struct {
	Config
}

func (p java) SanitizeDependencies(dependencies string) string {
	return dependencies
}

var _ Runtime = java{}
