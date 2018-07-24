package v2

import (
	"testing"

	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	istioFakes "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateAuthentication_ShouldCreateNewPolicy(t *testing.T) {

	// given
	fakeIstioAuth := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioAuth)

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
	}

	// when
	gatewayResource, err := authentication.Create(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource == nil {
		t.Error("Gateway resource should not be nil.")
		return
	}

	if gatewayResource.Name != "test-api" {
		t.Error("Gateway resource name should be the same as api name.")
		return
	}
}

func TestCreateAuthentication_ShouldNotCreatePolicyIfDisabled(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioConfig)

	var dto *Dto = nil

	// when
	gatewayResource, err := authentication.Create(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource != nil {
		t.Error("Gateway resource should be nil.")
		return
	}
}

func TestCreateAuthRule_ShouldNotCreatePolicyIfRulesEmpty(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioConfig)

	dto := &Dto{
		MetaDto: meta.Dto{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		ServiceName: "dummy-service",
	}

	// when
	gatewayResource, err := authentication.Create(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource != nil {
		t.Error("Gateway resource should be nil.")
		return
	}
}

func TestUpdateAuthentication_ShouldCreateNewPolicy(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioConfig)

	oldDto := &Dto{}

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
					Issuer:  "https://accounts.google.com",
					JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
				},
			},
		},
	}

	// when
	gatewayResource, err := authentication.Update(oldDto, newDto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource == nil {
		t.Error("Gateway resource should not be nil.")
		return
	}

	if gatewayResource.Name != "test-api" {
		t.Error("Gateway resource name should be the same as api name.")
		return
	}
}

func TestUpdateAuthentication_ShouldDeleteOldPolicyIfAuthenticationDisabled(t *testing.T) {

	// given
	testAuthentication := &istioAuthApi.Policy{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Spec: &istioAuthApi.PolicySpec{},
	}
	fakeIstioConfig := istioFakes.NewSimpleClientset(testAuthentication)
	authentication := New(fakeIstioConfig)

	oldDto := &Dto{
		MetaDto: meta.Dto{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Status: kymaMeta.GatewayResourceStatus{
			Resource: kymaMeta.GatewayResource{
				Name: "test-api",
			},
		},
	}

	var newDto *Dto = nil

	// when
	gatewayResource, err := authentication.Update(oldDto, newDto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource != nil {
		t.Error("Gateway resource should be nil (should only delete old resource).")
		return
	}
}

func TestUpdateAuthentication_ShouldDeleteOldPolicyIfRulesEmpty(t *testing.T) {

	// given
	testAuthentication := &istioAuthApi.Policy{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Spec: &istioAuthApi.PolicySpec{},
	}
	fakeIstioConfig := istioFakes.NewSimpleClientset(testAuthentication)
	authentication := New(fakeIstioConfig)

	oldDto := &Dto{
		MetaDto: meta.Dto{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Status: kymaMeta.GatewayResourceStatus{
			Resource: kymaMeta.GatewayResource{
				Name: "test-api",
			},
		},
	}

	newDto := &Dto{
		MetaDto: meta.Dto{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		ServiceName: "dummy-service",
	}

	// when
	gatewayResource, err := authentication.Update(oldDto, newDto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource != nil {
		t.Error("Gateway resource should be nil (should only delete old resource).")
		return
	}
}

func TestUpdateAuthentication_ShouldDoNothingIfSRulesHasNotChanged(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioConfig)

	oldApi := &Dto{
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
		Status: kymaMeta.GatewayResourceStatus{
			Resource: kymaMeta.GatewayResource{
				Name: "test-api",
			},
		},
	}

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
					Issuer:  "https://accounts.google.com",
					JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
				},
			},
		},
	}

	// when
	gatewayResource, err := authentication.Update(oldApi, newDto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if gatewayResource != nil {
		t.Error("Gateway resource should be nil (should do nothing).")
		return
	}
}

func TestDeleteAuthentication(t *testing.T) {

	// given
	testAuthentication := &istioAuthApi.Policy{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      "test-api",
			Namespace: "test-namespace",
		},
		Spec: &istioAuthApi.PolicySpec{},
	}
	fakeIstioConfig := istioFakes.NewSimpleClientset(testAuthentication)
	authentication := New(fakeIstioConfig)

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
	}

	// when
	err := authentication.Delete(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}

func TestDeleteAuthentication_ShouldNotFailIfNil(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authentication := New(fakeIstioConfig)

	var dto *Dto = nil

	// when
	err := authentication.Delete(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}

func TestDeleteAuthentication_ShouldNotFailIfOldNameEmpty(t *testing.T) {

	// given
	fakeIstioConfig := istioFakes.NewSimpleClientset()
	authRules := New(fakeIstioConfig)

	dto := &Dto{
		MetaDto: meta.Dto{
			Namespace: "test-namepsace",
		},
	}

	// when
	err := authRules.Delete(dto)

	// then
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}
