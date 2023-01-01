package registry

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func Test_basicClientWithOptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		opts    *RegistryClientOptions
		want    *registryClient
		wantErr bool
	}{
		{
			name: "valid options",
			opts: &RegistryClientOptions{
				Username: "test-user",
				Password: "test-password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:      ctx,
				username: "test-user",
				password: "test-password",
				url:      validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000"),
				logger:   logr.Discard(),
			},
		},
		{
			name: "valid options with URL scheme",
			opts: &RegistryClientOptions{
				Username: "test-user",
				Password: "test-password",
				URL:      "http://docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:      ctx,
				username: "test-user",
				password: "test-password",
				url:      validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000"),
				logger:   logr.Discard(),
			},
		},
		{
			name: "Empty username",
			opts: &RegistryClientOptions{
				Username: "",
				Password: "test-password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := basicClientWithOptions(ctx, tt.opts, logr.Discard())
			if (err != nil) != tt.wantErr {
				t.Errorf("basicClientWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basicClientWithOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	require.NoError(t, err)

	return u
}
