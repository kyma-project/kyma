package runtime

type python struct {
	Config
}

func (p python) SanitizeDependencies(dependencies string) string {
	return dependencies
}

var _ Runtime = python{}
