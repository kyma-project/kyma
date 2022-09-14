//go:build unit
// +build unit

package httpclient_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
)

func TestErrorDescription(t *testing.T) {
	tableTests := []struct {
		name            string
		giveErr         *httpclient.Error
		wantDescription string
	}{
		{
			name: "all arguments given",
			giveErr: func() *httpclient.Error {
				wrapped := httpclient.NewError(fmt.Errorf("my error"), httpclient.WithStatusCode(500), httpclient.WithMessage("this is the http response"))
				return wrapped
			}(),
			wantDescription: `message: this is the http response; status code: 500; cause: my error`,
		},
		{
			name: "cause only",
			giveErr: func() *httpclient.Error {
				existing := url.Error{
					Op:  "Delete",
					URL: "/foo/bar",
					Err: errors.New("unsupported protocol scheme"),
				}
				wrapped := httpclient.NewError(&existing)
				return wrapped
			}(),
			wantDescription: `cause: Delete "/foo/bar": unsupported protocol scheme`,
		},
		{
			name: "message only",
			giveErr: func() *httpclient.Error {
				wrapped := httpclient.NewError(nil, httpclient.WithMessage("message"))
				return wrapped
			}(),
			wantDescription: `message: message`,
		},
		{
			name: "status code only",
			giveErr: func() *httpclient.Error {
				wrapped := httpclient.NewError(nil, httpclient.WithStatusCode(200))
				return wrapped
			}(),
			wantDescription: `status code: 200`,
		},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.giveErr.Error() != tt.wantDescription {
				t.Errorf("error message should not be %q, but: %q", tt.giveErr.Error(), tt.wantDescription)
			}
		})
	}
}
