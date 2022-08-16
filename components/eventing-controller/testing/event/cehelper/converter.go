package cehelper

import (
	"context"
	"fmt"
	"net/http"

	cebinding "github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

// RequestToEventString takes a http request that is based on a CloudEvent and tries to convert it to a string that
// contains all the CloudEvent related data. Here is an example of a possible output:
//
// Context Attributes,
//
//	specversion: 1.0
//	type: prefix.testapp1023.order.created.v1
//	source: /default/sap.kyma/id
//	subject: prefix.testapp1023.order.created.v1
//	id: 1.0
//	time: 2022-07-15T18:40:15.808918Z
//	datacontenttype: application/json
//
// Data,
//
//	"{\"foo\":\"bar\"}"
func RequestToEventString(r *http.Request) (string, error) {
	msg := cehttp.NewMessageFromHttpRequest(r)
	event, err := cebinding.ToEvent(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("failed to build a CloudEvent: %s", err.Error())
	}
	return event.String(), nil
}
