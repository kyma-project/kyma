package process

// Step defines interface for process steps
type Step interface {
	Do() error
	ToString() string
}
