package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	prometheusNamespace = "kyma-system"
	prometheusPod       = "prometheus-monitoring-prometheus-0"
	prometheusPort      = "9090"
)

func portForwardToPrometheus(config *rest.Config) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", prometheusNamespace, prometheusPod)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{prometheusPort}, stopChan, readyChan, out, errOut)
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

	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			panic(err)
		}
	}()
}

func queryPrometheus(ctx context.Context, query string, t time.Time) (*model.Sample, error) {
	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%s", prometheusPort),
	})
	if err != nil {
		return nil, err
	}

	v1api := v1.NewAPI(client)
	result, warnings, err := v1api.Query(ctx, query, t)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	if vector, ok := result.(model.Vector); ok && len(vector) == 1 {
		return vector[0], nil
	}

	return nil, errors.New("unsupported result")
}
