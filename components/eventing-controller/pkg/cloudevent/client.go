// Package cloudevent provides an abstraction over the cloudvent sdk.
// Use this in favor of using the cloudevents sdk directly.
// This helps in testing, as we have the code under our control.
// Therefore, it is easier to use seams.
package cloudevent

import (
	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// ClientFactoryInterface is an interface for the creation of a cloudevent client.
type ClientFactoryInterface interface {
	NewHTTP(opts ...http.Option) (cev2.Client, error)
}

type ClientFactory struct{}

// NewHTTP creates a cloudevent client at the HTTP protocol level.
func (c ClientFactory) NewHTTP(opts ...http.Option) (cev2.Client, error) {
	ceClient, err := cev2.NewClientHTTP(opts...)
	return ceClient, err
}
