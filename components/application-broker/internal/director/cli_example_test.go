// +build director_integration

package director

import (
	"context"
	"fmt"
	"os"
	"testing"

	gcli "github.com/kyma-project/kyma/components/application-broker/third_party/machinebox/graphql"
)

// Test_FindPackageInstanceAuth executes  call to a director to get an application
//
// go test -v -tags=director_integration ./internal/director/... -run=Test_FindPackageInstanceAuth -count=1
func Test_FindPackageInstanceAuth(t *testing.T) {
	directorURL := os.Getenv("DIRECTOR_URL")
	auth := os.Getenv("AUTHORIZATION")
	appID := os.Getenv("APP_ID")
	packageID := os.Getenv("PACKAGE_ID")
	instanceAuthID := os.Getenv("INSTANCE_AUTH_ID")
	if directorURL == "" {
		directorURL = "http://localhost:3000/graphql"
	}
	if auth == "" {
		auth = "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0."
	}
	if appID == "" {
		appID = "5cb9e3ea-6ff5-4570-9ef3-f882c81214a1"
	}
	if packageID == "" {
		packageID = "pid"
	}
	if instanceAuthID == "" {
		instanceAuthID = "inst-id"
	}

	cli := NewQGLClient(&authClient{
		Target:              gcli.NewClient(directorURL),
		AuthorizationHeader: auth,
	})

	out, err := cli.FindPackageInstanceAuth(context.TODO(), FindPackageInstanceAuthInput{
		ApplicationID:  appID,
		InstanceAuthID: instanceAuthID,
		PackageID:      packageID,
	})

	fmt.Printf("Output: %v\n", out)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

type authClient struct {
	Target              GraphQLClient
	AuthorizationHeader string
}

func (c *authClient) Run(ctx context.Context, req *gcli.Request, resp interface{}) error {
	req.Header.Set("Authorization", c.AuthorizationHeader)
	return c.Target.Run(ctx, req, resp)
}
