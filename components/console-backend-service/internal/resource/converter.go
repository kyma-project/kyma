package resource

type Converter interface {
	NewK8s() interface{}
}
