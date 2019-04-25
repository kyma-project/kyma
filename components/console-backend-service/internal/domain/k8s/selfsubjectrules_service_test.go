package k8s_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest/fake"
)

func TestSelfSubjectRules_Create(t *testing.T) {

	var testCases = []struct {
		caseName          string
		failingRESTClient bool
		success           bool
	}{
		{"Success", false, true},
		{"ErrorCreatingResource", true, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {

			requestData := fixSSRRRequestData()

			client := fixAuthorizationClient(requestData, testCase.failingRESTClient)
			svc := k8s.NewSelfSubjectRulesService(client)

			fakeContext := authn.WithUserInfoContext(context.TODO(), &user.DefaultInfo{Name: "fake"})

			result, err := svc.Create(fakeContext, &requestData)

			if testCase.success {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
				assert.Nil(t, result)
			}
		})
	}
}

func fixSSRRRequestData() authv1.SelfSubjectRulesReview {
	return authv1.SelfSubjectRulesReview{
		Spec: authv1.SelfSubjectRulesReviewSpec{
			Namespace: "*",
		},
	}
}

func fixAuthorizationClient(payload authv1.SelfSubjectRulesReview, failingRESTClient bool) v1.AuthorizationV1Interface {

	client := automock.AuthorizationV1Interface{}
	fakeRESTClient := &fake.RESTClient{
		NegotiatedSerializer: scheme.Codecs,
		Client:               fixHTTPClient([]byte(fmt.Sprintf("%v", payload)), failingRESTClient),
	}
	client.On("RESTClient").Return(fakeRESTClient, nil).Once()

	return &client
}
