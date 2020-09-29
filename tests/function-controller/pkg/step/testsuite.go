package step

type TestSuite interface {
	OnError(err error) error
	Steps() []Step
}
