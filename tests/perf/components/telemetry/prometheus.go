package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func portForwardToPrometheus(config *rest.Config) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", "kyma-system", "prometheus-monitoring-prometheus-0")
	hostIP := strings.TrimLeft(config.Host, "https:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{"9090"}, stopChan, readyChan, out, errOut)
	if err != nil {
		panic(err)
	}

	go func() {
		for range readyChan {
		}
		if len(errOut.String()) != 0 {
			panic(errOut.String())
		} else if len(out.String()) != 0 {
			fmt.Println(out.String())
		}
	}()

	if err = forwarder.ForwardPorts(); err != nil {
		panic(err)
	}
}
