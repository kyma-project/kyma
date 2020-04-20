package shared

type Logger interface {
	Logf(format string, args ...interface{})
}
