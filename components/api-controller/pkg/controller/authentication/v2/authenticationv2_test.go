package v2

import (
	"testing"

	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	istioFakes "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateAuthentication(t *testing.T) {

	fakeIstioAuth := istioFakes.NewSimpleClientset()
	fakeJwtDefaultConfig := JwtDefaultConfig{
		Issuer:  "https://accounts.google.com",
		JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
	}
	authentication := New(fakeIstioAuth, fakeJwtDefaultConfig)

	t.Run("Should create policy with custom auth configuration", func(t *testing.T) {

		dto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			ServiceName: "dummy-service",
			Rules: Rules{
				{
					Type: JwtType,
					Jwt: Jwt{
						Issuer:  "https://accounts.google.com",
						JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
					},
				},
			},
			AuthenticationEnabled: true,
		}

		gatewayResource, err := authentication.Create(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource == nil {
			t.Error("Gateway resource should not be nil.")
		}

		if gatewayResource.Name != "test-api" {
			t.Error("Gateway resource name should be the same as api name.")
		}
	})

	t.Run("Should create new policy with default jwt config if authentication enabled and rules not provided", func(t *testing.T) {
		dto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api2",
				Namespace: "test-namespace",
			},
			ServiceName:           "dummy-service",
			AuthenticationEnabled: true,
		}

		gatewayResource, err := authentication.Create(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource == nil {
			t.Error("Gateway resource should not be nil.")
		} else if gatewayResource.Name != "test-api2" {
			t.Error("Gateway resource name should be the same as api name.")
		}
	})

	t.Run("Should not create policy if authentication disabled", func(t *testing.T) {
		var dto *Dto = &Dto{
			AuthenticationEnabled: false,
		}

		gatewayResource, err := authentication.Create(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource != nil {
			t.Error("Gateway resource should be nil.")
		}
	})
}

func TestUpdateAuthentication(t *testing.T) {

	fakeIstioConfig := istioFakes.NewSimpleClientset()
	fakeJwtDefaultConfig := JwtDefaultConfig{
		Issuer:  "https://accounts.google.com",
		JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
	}
	authentication := New(fakeIstioConfig, fakeJwtDefaultConfig)

	oldDto := &Dto{
		AuthenticationEnabled: false,
	}

	customIssuer := "https://test.issuer.com"

	t.Run("Should create new policy if authentication was disabled and now is enabled", func(t *testing.T) {
		newDto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			ServiceName: "dummy-service",
			Rules: Rules{
				{
					Type: JwtType,
					Jwt: Jwt{
						Issuer:  customIssuer,
						JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
					},
				},
			},
			AuthenticationEnabled: true,
		}

		gatewayResource, err := authentication.Update(oldDto, newDto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource == nil {
			t.Error("Gateway resource should not be nil.")
		} else if gatewayResource.Name != "test-api" {
			t.Error("Gateway resource name should be the same as api name.")
		}

		oldDto = newDto // set oldDto for next test
		oldDto.Status.Resource = *gatewayResource
	})

	t.Run("Should do nothing if nothing has changed", func(t *testing.T) {
		newDto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			ServiceName: "dummy-service",
			Rules: Rules{
				{
					Type: JwtType,
					Jwt: Jwt{
						Issuer:  customIssuer,
						JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
					},
				},
			},
			AuthenticationEnabled: true,
		}

		gatewayResource, err := authentication.Update(oldDto, newDto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource.Version != oldDto.Status.Resource.Version {
			t.Error("Gateway resource should be nil (should do nothing).")
		}
	})

	t.Run("Should update policy with default jwt config if rules was custom and now rules are empty and authentication is enabled", func(t *testing.T) {
		newDto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			ServiceName:           "dummy-service",
			AuthenticationEnabled: true,
			Rules: Rules{},
		}

		gatewayResource, err := authentication.Update(oldDto, newDto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource == nil {
			t.Error("Gateway resource should not be nil.")
		} else if gatewayResource.Name != "test-api" {
			t.Error("Gateway resource name should be the same as api name.")
		}

		oldDto = newDto // set oldDto for next test
		oldDto.Status.Resource = *gatewayResource
	})

	t.Run("Should delete old policy if authentication disabled", func(t *testing.T) {
		var newDto *Dto = &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			ServiceName:           "dummy-service",
			AuthenticationEnabled: false,
		}

		gatewayResource, err := authentication.Update(oldDto, newDto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if gatewayResource != nil {
			t.Error("Gateway resource should be nil (should only delete old resource).")
		}
	})
}

func TestDeleteAuthentication(t *testing.T) {

	testAuthentication := &istioAuthApi.Policy{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Spec: &istioAuthApi.PolicySpec{},
	}
	fakeIstioConfig := istioFakes.NewSimpleClientset(testAuthentication)
	fakeJwtDefaultConfig := JwtDefaultConfig{
		Issuer:  "https://accounts.google.com",
		JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
	}
	authentication := New(fakeIstioConfig, fakeJwtDefaultConfig)

	t.Run("Should delete Policy if exists", func(t *testing.T) {
		dto := &Dto{
			MetaDto: meta.Dto{
				Name:      "test-api",
				Namespace: "test-namespace",
			},
			Status: kymaMeta.GatewayResourceStatus{
				Resource: kymaMeta.GatewayResource{
					Name:    "test-api",
					Version: "1",
				},
			},
			AuthenticationEnabled: false,
		}

		err := authentication.Delete(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("Should not fail if Policy doesn't exists", func(t *testing.T) {
		var dto *Dto = &Dto{
			AuthenticationEnabled: false,
		}

		err := authentication.Delete(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("Should not fail if old Policy name is empty", func(t *testing.T) {
		dto := &Dto{
			MetaDto: meta.Dto{
				Namespace: "test-namepsace",
			},
			AuthenticationEnabled: false,
		}

		err := authentication.Delete(dto)

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
