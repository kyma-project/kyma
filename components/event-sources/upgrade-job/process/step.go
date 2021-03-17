package process

type Step interface {
	Do() error
	ToString() string
}
