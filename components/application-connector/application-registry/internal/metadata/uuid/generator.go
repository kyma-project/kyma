package uuid

// Generator is an interface of an UUID generator.
type Generator interface {
	// NewUUID generates an UUID.
	NewUUID() (string, error)
}

// GeneratorFunc is an adapter that simplifies creating Generator.
type GeneratorFunc func() (string, error)

// NewUUID returns new UUID. It calls the inner method of GeneratorFunc.
func (f GeneratorFunc) NewUUID() (string, error) {
	return f()
}
