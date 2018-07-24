package uuid

// Generator is an interface of an UUID generator.
type Generator interface {
	// NewUUID generates an UUID.
	NewUUID() string
}

// GeneratorFunc is an adapter that simplifies creating Generator.
type GeneratorFunc func() string

// NewUUID returns new UUID. It calls the inner method of GeneratorFunc.
func (f GeneratorFunc) NewUUID() string {
	return f()
}
