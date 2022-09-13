package proxy_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/proxy"
)

func TestGatewayPath(t *testing.T) {
	type testCase struct {
		name     string
		url      string
		expected string
	}

	cases := []testCase{
		{
			name:     "OSS URL",
			url:      "http://central-application-gateway.kyma-system:8080/someapp/somesrv/api/endpoint",
			expected: "http://central-application-gateway.kyma-system:8080/someapp/somesrv",
		},
		{
			name:     "Compass URL",
			url:      "http://central-application-gateway.kyma-system:8082/someapp/somesrv/api/endpoint",
			expected: "http://central-application-gateway.kyma-system:8082/someapp/somesrv",
		},
		{
			name:     "No API Path",
			url:      "http://central-application-gateway.kyma-system:8082/someapp/somesrv",
			expected: "http://central-application-gateway.kyma-system:8082/someapp/somesrv",
		},
		{
			name:     "Trailing Slash",
			url:      "http://central-application-gateway.kyma-system:8080/someapp/somesrv/",
			expected: "http://central-application-gateway.kyma-system:8080/someapp/somesrv",
		},
		{
			name:     "Random words",
			url:      "https://non.quam.lacus.suspendisse.faucibus.interdum.posuere/someapp/somesrv",
			expected: "https://non.quam.lacus.suspendisse.faucibus.interdum.posuere/someapp/somesrv",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			input, err := url.Parse(c.url)
			require.Nil(t, err)
			output, err := proxy.ExtractGatewatURL(input)
			assert.Nil(t, err)
			assert.Equal(t, c.expected, output.String())

			fmt.Printf("%+v\n", *output)
		})
	}
}

// {
// 	Scheme:		http
// 	Opaque:
// 	User:
// 	Host:		central-application-gateway.kyma-system:8082
// 	Path:		/someapp/somesrv
// 	RawPath:
// 	OmitHost:	false
// 	ForceQuery:	false
// 	RawQuery:
// 	Fragment:
// 	RawFragment:
// }
