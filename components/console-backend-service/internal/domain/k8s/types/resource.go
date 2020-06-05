package types

type Resource struct {
	APIVersion string
	Name       string
	Namespace  string
	Kind       string
	Body       []byte
}
