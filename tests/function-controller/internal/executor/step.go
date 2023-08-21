package executor

// Step represents a single action in test scenario
type Step interface {
	// Name returns Name of the step
	Name() string
	// Run executes the step
	Run() error
	// Cleanup removes all resources that may possibly created by the step
	Cleanup() error
	// OnError is callback in case of error
	OnError() error
}
