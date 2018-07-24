package dex

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
)

const (
	ingressControllerServiceURL = "istio-ingress.istio-system.svc.cluster.local"
	domainEnvName               = "KYMA_DOMAIN"
	isLocalEnvEnvName           = "IS_LOCAL_ENV"
	crossClient                 = "kyma-client"
	username                    = "admin@kyma.cx"
	password                    = "nimda123"
)

func TestSpec(t *testing.T) {

	isLocalEnv, envFound := os.LookupEnv(isLocalEnvEnvName)
	if !envFound {
		t.Fatal(isLocalEnvEnvName + " env variable not set")
	}

	if testRunningInLocalEnv, err := strconv.ParseBool(isLocalEnv); err != nil {
		t.Fatal(err)
	} else if !testRunningInLocalEnv {
		t.Skip()
	}

	ingressControllerAddr, err := net.LookupHost(ingressControllerServiceURL)
	if err != nil {
		t.Fatal(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Changes request destination address to ingress internal cluster address for requests to dex-web.
			// Because hostNetwork can't be set to true due to use of internal cluster addresses in servicecatalog test in this repo, dex and dex-web isn't by default accessible by its external URL for pod running this tests. There are two ways out - one to use internal address of dex-web service, and second (implemented here) to use ingress controller like it was a request from outside of the cluster.
			if strings.HasPrefix(addr, "dex") {
				addr = ingressControllerAddr[0] + ":443"
			}
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	Convey("Should issue an ID token", t, func() {

		domain, envFound := os.LookupEnv(domainEnvName)
		if !envFound {
			t.Fatal(domainEnvName + " env variable not set")
		}

		dexURL := "https://dex." + domain
		dexWebApplicationURL := "https://dex-web." + domain

		resp, err := client.PostForm(dexWebApplicationURL+"/login", url.Values{"cross_client": {crossClient}})

		So(err, ShouldBeNil)

		ap := findAuthenticationPath(resp.Body)

		_, err = client.Get(dexURL + ap)

		So(err, ShouldBeNil)

		resp, err = client.PostForm(dexURL+ap, url.Values{"login": {username}, "password": {password}})

		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)

		b, _ := ioutil.ReadAll(resp.Body)
		sb := string(b)

		So(sb, ShouldContainSubstring, "Token:")
		So(sb, ShouldContainSubstring, "Claims:")
		So(sb, ShouldContainSubstring, crossClient)
		So(sb, ShouldContainSubstring, username)
	})
}

func findAuthenticationPath(data io.ReadCloser) string {

	getHrefAttribute := func(t html.Token) (href string) {

		for _, a := range t.Attr {
			if a.Key == "href" {
				href = a.Val
			}
		}

		return
	}

	tnr := html.NewTokenizer(data)

	for {
		t := tnr.Next()

		switch {
		case t == html.ErrorToken:
			return ""
		case t == html.StartTagToken:
			t := tnr.Token()

			isAnchor := t.Data == "a"
			if isAnchor {

				ap := getHrefAttribute(t)

				if strings.Contains(ap, "/auth/local") {

					return ap
				}
			}
		}
	}
}
