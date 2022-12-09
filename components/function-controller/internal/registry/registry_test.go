package registry

import (
	"context"
	"net/url"
	"reflect"
	"testing"
)

func TestNewRegistryClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		opts    *RegistryClientOptions
		want    RegistryClient
		wantErr bool
	}{
		{
			name: "valid options",
			opts: &RegistryClientOptions{
				Username: "username",
				Password: "password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:      ctx,
				userName: "username",
				password: "password",
				url:      validURL("http://docker-registry.kyma-system.svc.cluster.local:5000"),
			},
		},
		{
			name: "valid options with URL scheme",
			opts: &RegistryClientOptions{
				Username: "username",
				Password: "password",
				URL:      "http://docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:      ctx,
				userName: "username",
				password: "password",
				url:      validURL("http://docker-registry.kyma-system.svc.cluster.local:5000"),
			},
		},
		{
			name: "Empty username",
			opts: &RegistryClientOptions{
				Username: "",
				Password: "password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRegistryClient(ctx, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistryClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRegistryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
