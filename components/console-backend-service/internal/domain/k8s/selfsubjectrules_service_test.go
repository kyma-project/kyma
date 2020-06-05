package k8s_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			bytes, err := json.Marshal(requestData)
			client := fixAuthorizationClient(bytes, testCase.failingRESTClient)
			svc := k8s.NewSelfSubjectRulesService(client)

			fakeContext := authn.WithUserInfoContext(context.TODO(), &user.DefaultInfo{Name: "fake"})

			result, err := svc.Create(fakeContext, bytes)

			if testCase.success {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixSSRRRequestData() authv1.SelfSubjectRulesReview {
	return authv1.SelfSubjectRulesReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SelfSubjectRulesReview",
			APIVersion: "authorization.k8s.io/v1",
		},
		Spec: authv1.SelfSubjectRulesReviewSpec{
			Namespace: "*",
		},
	}
}

func fixAuthorizationClient(payload []byte, failingRESTClient bool) v1.AuthorizationV1Interface {

	client := automock.AuthorizationV1Interface{}
	fakeRESTClient := &fake.RESTClient{
		NegotiatedSerializer: scheme.Codecs,
		Client:               fixHTTPClient(payload, failingRESTClient),
	}
	client.On("RESTClient").Return(fakeRESTClient, nil).Once()

	return &client
}
