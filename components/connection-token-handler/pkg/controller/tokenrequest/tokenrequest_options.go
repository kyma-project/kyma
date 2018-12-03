package tokenrequest

// Options describes the parameters struct handed to tokenrequest reconciller
type Options struct {
	TokenTTL            int
	ConnectorServiceURL string
}
