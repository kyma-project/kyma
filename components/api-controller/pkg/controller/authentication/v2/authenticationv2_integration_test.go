package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	istioAuth "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	"k8s.io/client-go/tools/clientcmd"
)

func TestIntegrationCreateUpdateAndDeleteAuthentication(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
		return
	}

	// given

	authentication, err := authenticationFromDefaultConfig()
	if err != nil {
		t.Error(err)
		return
	}

	rules := Rules{
		{
			Type: JwtType,
			Jwt: Jwt{
				Issuer:  "https://accounts.google.com",
				JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
			},
		},
	}

	authDto := &Dto{
		MetaDto: meta.Dto{
			Name:      "httpbin-api",
			Namespace: "default",
		},
		ServiceName: "sample-app-kfvcdftg-0",
		Rules:       rules,
		AuthenticationEnabled: true,
	}

	// when

	t.Logf("CREATE: %+v", authDto)

	createdResource, err2 := authentication.Create(authDto)
	if err2 != nil {
		t.Errorf("Unable to create authentication. Root cause: %v", err2)
		return
	}
	authDto.Status.Resource = *createdResource

	t.Logf("CREATED RESOURCE: %+v", createdResource)

	rules = Rules{
		{
			Type: JwtType,
			Jwt: Jwt{
				Issuer:  "https://dex.nightly.cluster.kyma.cx",
				JwksUri: "https://dex.nightly.cluster.kyma.cx/keys",
			},
		},
	}

	t.Logf("UPDATE: %+v", authDto)

	deleteAuthentication := func() {

		t.Logf("DELETE: %+v", authDto)

		err4 := authentication.Delete(authDto)
		if err4 != nil {
			t.Errorf("Unable to delete authentication. Root cause: %v", err4)
			return
		}
	}
	defer deleteAuthentication()

	oldApiDto := &Dto{
		Status: kymaMeta.GatewayResourceStatus{
			Resource: *createdResource,
		},
		AuthenticationEnabled: true,
	}

	_, err3 := authentication.Update(oldApiDto, authDto)
	if err3 != nil {
		t.Errorf("Unable to update authentication. Root cause: %v", err3)
		return
	}
}

func authenticationFromDefaultConfig() (Interface, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to load kube config. Root cause: %v", err)
	}

	clientset := istioAuth.NewForConfigOrDie(kubeConfig)

	sampleJwtDefaultConfig := JwtDefaultConfig{
		Issuer:  "https://accounts.google.com",
		JwksUri: "https://www.googleapis.com/oauth2/v3/certs",
	}

	return New(clientset, sampleJwtDefaultConfig, true), nil
}
