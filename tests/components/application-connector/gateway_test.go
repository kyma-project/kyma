package main_test

import (
	"context"
	"net/http"
	"testing"

	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestSimpleCases(t *testing.T) {
	cfg, err := rest.InClusterConfig()
	require.Nil(t, err)

	cl, err := cli.NewForConfig(cfg)

	require.Nil(t, err)

	app, err := cl.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "test-app", v1.GetOptions{})

	require.Nil(t, err)

	for _, service := range app.Spec.Services {
		t.Run(service.DisplayName, func(t *testing.T) {
			for _, entry := range service.Entries {
				if entry.Type != "API" {
					t.Log("Skipping event entry")
					continue
				}

				t.Log("Calling", entry.CentralGatewayUrl)
				res, err := http.Get(entry.CentralGatewayUrl)
				assert.Nil(t, err)
				assert.Equal(t, 200, res.StatusCode)
			}
		})
	}
}
