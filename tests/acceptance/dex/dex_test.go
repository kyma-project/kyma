package dex

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/glog"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	ingressGatewayControllerServiceURLEnv = "INGRESSGATEWAY_FQDN"
	domainEnvName                         = "KYMA_DOMAIN"
	isLocalEnvEnvName                     = "IS_LOCAL_ENV"
	clientId                              = "kyma-client"
	username                              = "admin@kyma.cx"
	password                              = "nimda123"
)

func TestSpec(t *testing.T) {

	ingressGatewayControllerServiceURL, envFound := os.LookupEnv(ingressGatewayControllerServiceURLEnv)
	if !envFound {
		t.Fatal(ingressGatewayControllerServiceURLEnv + " env variable not set")
	}

	isLocalEnv, envFound := os.LookupEnv(isLocalEnvEnvName)
	if !envFound {
		t.Fatal(isLocalEnvEnvName + " env variable not set")
	}

	testRunningInLocalEnv, err := strconv.ParseBool(isLocalEnv)
	if err != nil {
		t.Fatal(err)
	} else if !testRunningInLocalEnv {
		t.Skip()
	}

	ingressGatewayControllerAddr, err := net.LookupHost(ingressGatewayControllerServiceURL)
	if err != nil {

		glog.Warningf("Unable to resolve host '%s' (if you are running this test from outside of Kyma ignore this log). Root cause: %v", ingressGatewayControllerServiceURL, err)

		if minikubeIp := tryToGetMinikubeIp(); minikubeIp != "" {
			ingressGatewayControllerAddr = []string{minikubeIp}
		} else {
			t.Fatal(err)
		}
	}

	domain, envFound := os.LookupEnv(domainEnvName)
	if !envFound {
		t.Fatal(domainEnvName + " env variable not set")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if testRunningInLocalEnv {
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Changes request destination address to ingress gateway internal cluster address for requests to dex.
			if strings.HasPrefix(addr, "dex") {
				addr = ingressGatewayControllerAddr[0] + ":443"
			}
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	dexHttpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	idProviderConfig := idProviderConfig{
		dexConfig: dexConfig{
			baseUrl:           fmt.Sprintf("https://dex.%s", domain),
			authorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", domain),
			tokenEndpoint:     fmt.Sprintf("https://dex.%s/token", domain),
		},
		clientConfig: clientConfig{
			id:          clientId,
			redirectUri: "http://127.0.0.1:5555/callback",
		},

		userCredentials: userCredentials{
			username: username,
			password: password,
		},
	}

	idTokenProvider := newDexIdTokenProvider(dexHttpClient, idProviderConfig)

	Convey("Should issue an ID token", t, func() {

		idToken, err := idTokenProvider.fetchIdToken()

		So(err, ShouldBeNil)
		So(idToken, ShouldNotBeEmpty)

		tokenParts := strings.Split(idToken, ".")

		tokenPayloadEncoded := tokenParts[1] + "="

		tokenPayloadDecoded, err := base64.StdEncoding.DecodeString(tokenPayloadEncoded)
		if err != nil {
			t.Fatal(err)
		}

		tokenPayload := make(map[string]interface{})
		err = json.Unmarshal(tokenPayloadDecoded, &tokenPayload)

		So(tokenPayload["iss"].(string), ShouldEqual, idProviderConfig.dexConfig.baseUrl)
		So(tokenPayload["aud"].(string), ShouldEqual, idProviderConfig.clientConfig.id)
		So(tokenPayload["email"].(string), ShouldEqual, idProviderConfig.userCredentials.username)
		So(tokenPayload["email_verified"].(bool), ShouldBeTrue)
	})
}

func tryToGetMinikubeIp() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		glog.Warningf("Error while getting minikube IP (ignore this message if you are running this test inside Kyma). Root cause: %s", err)
		return ""
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}
