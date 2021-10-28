// +build integration

package dex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/common/ingressgateway"
	"github.com/kyma-project/kyma/common/resilient"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vrischmann/envconfig"
)

const (
	clientId = "kyma-client"
)

type Config struct {
	IsLocalEnv      bool   `envconfig:"default=true"`
	DomainName      string `envconfig:"default=kyma.local"`
	DexUserEmail    string `envconfig:"default=admin@kyma.cx"`
	DexUserPassword string
}

func TestSpec(t *testing.T) {
	cfg := Config{}
	err := envconfig.Init(&cfg)
	if err != nil {
		t.Errorf("Error while loading env config: %s", err)
	}

	if !cfg.IsLocalEnv {
		t.Skip("Test is enabled only in local env")
	}

	ingressClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Error while creating ingress gateway client: %s", err)
	}

	ingressClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := resilient.WrapHttpClient(ingressClient)

	idProviderConfig := idProviderConfig{
		dexConfig: dexConfig{
			baseUrl:           fmt.Sprintf("https://dex.%s", cfg.DomainName),
			authorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", cfg.DomainName),
			tokenEndpoint:     fmt.Sprintf("https://dex.%s/token", cfg.DomainName),
		},

		clientConfig: clientConfig{
			id:          clientId,
			redirectUri: "http://127.0.0.1:5555/callback",
		},

		userCredentials: userCredentials{
			username: cfg.DexUserEmail,
			password: cfg.DexUserPassword,
		},
	}

	idTokenProvider := newDexIdTokenProvider(client, idProviderConfig)

	Convey("Should issue an ID token", t, func() {

		idToken, err := idTokenProvider.fetchIdToken()

		So(err, ShouldBeNil)
		So(idToken, ShouldNotBeEmpty)

		tokenParts := strings.Split(idToken, ".")
		So(len(tokenParts), ShouldEqual, 3)

		tokenPayloadEncoded := tokenParts[1]

		missingTokenBytes := (3 - len(tokenPayloadEncoded)%3) % 3
		tokenPayloadEncoded += strings.Repeat("=", missingTokenBytes)

		tokenPayloadDecoded, err := base64.StdEncoding.DecodeString(tokenPayloadEncoded)
		So(err, ShouldBeNil)

		tokenPayload := make(map[string]interface{})
		err = json.Unmarshal(tokenPayloadDecoded, &tokenPayload)

		So(err, ShouldBeNil)
		So(tokenPayload["iss"].(string), ShouldEqual, idProviderConfig.dexConfig.baseUrl)
		So(tokenPayload["aud"].(string), ShouldEqual, idProviderConfig.clientConfig.id)
		So(tokenPayload["email"].(string), ShouldEqual, idProviderConfig.userCredentials.username)
		So(tokenPayload["email_verified"].(bool), ShouldBeTrue)
	})
}
