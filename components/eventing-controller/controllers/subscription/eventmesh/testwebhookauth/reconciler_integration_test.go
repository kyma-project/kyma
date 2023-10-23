package testwebhookauth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apigatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_UpdateWebhookAuthConfig(t *testing.T) {
	t.Parallel()

	////
	// before updating the credentials
	////

	// setup
	err := setupSuite()
	require.NoError(t, err)
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	// ensure namespace created
	namespace := getTestNamespace()
	name := fmt.Sprintf("test-resource-%s", namespace)
	ensureNamespaceCreated(ctx, t, namespace)

	// ensure subscriber service is created
	subscriberService := reconcilertesting.NewSubscriberSvc(name, namespace)
	ensureK8sResourceCreated(ctx, t, subscriberService)

	// ensure Kyma subscription is created
	kymaSubscription := reconcilertesting.NewSubscription(
		name,
		namespace,
		reconcilertesting.WithDefaultSource(),
		reconcilertesting.WithOrderCreatedV1Event(),
		reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, name)),
	)
	ensureK8sResourceCreated(ctx, t, kymaSubscription)
	getSubscriptionAssert(ctx, g, kymaSubscription).Should(reconcilertesting.HaveNoneEmptyAPIRuleName())

	// ensure APIRule is created
	apiRule := &apigatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kymaSubscription.Status.Backend.APIRuleName,
			Namespace: kymaSubscription.Namespace,
		},
	}
	getAPIRuleAssert(ctx, g, apiRule).Should(reconcilertesting.HaveNotEmptyAPIRule())
	ensureAPIRuleStatusUpdatedWithStatusReady(ctx, t, apiRule)

	// ensure hashes are computed
	getSubscriptionAssert(ctx, g, kymaSubscription).Should(
		gomega.And(
			reconcilertesting.HaveSubscriptionReady(),
			reconcilertesting.HaveNonZeroEv2Hash(),
			reconcilertesting.HaveNonZeroEventMeshHash(),
			reconcilertesting.HaveNonZeroEventMeshLocalHash(),
			reconcilertesting.HaveNonZeroWebhookAuthHash(),
		),
	)
	webhookAuthHashBefore := kymaSubscription.Status.Backend.WebhookAuthHash

	// ensure EventMesh subscription is created
	eventMeshSubscription := getEventMeshSubFromMock(kymaSubscription.Name, kymaSubscription.Namespace)
	g.Expect(eventMeshSubscription).ShouldNot(gomega.BeNil())

	// counts EventMesh mock requests before changing the credentials
	var uri = getEventMeshSubKeyForMock(kymaSubscription.Name, kymaSubscription.Namespace)
	deleteRequestsBefore := emTestEnsemble.eventMeshMock.CountRequests(http.MethodDelete, uri)
	patchRequestsBefore := emTestEnsemble.eventMeshMock.CountRequests(http.MethodPatch, uri)

	////
	// update the credentials
	////

	// set the updated credentials
	updatedCredentials := &backendeventmesh.OAuth2ClientCredentials{
		ClientID:     credentials.ClientID + "-updated",
		ClientSecret: credentials.ClientSecret + "-updated",
		TokenURL:     credentials.TokenURL,
		CertsURL:     credentials.CertsURL,
	}
	setCredentials(updatedCredentials)

	// ensure Kyma subscription is updated
	kymaSubscription = kymaSubscription.DeepCopy()
	kymaSubscription.Labels = map[string]string{"reconcile": "true"}
	ensureK8sSubscriptionUpdated(ctx, t, kymaSubscription)

	// ensure hashes are computed
	getSubscriptionAssert(ctx, g, kymaSubscription).Should(
		gomega.And(
			reconcilertesting.HaveSubscriptionReady(),
			reconcilertesting.HaveNonZeroEv2Hash(),
			reconcilertesting.HaveNonZeroEventMeshHash(),
			reconcilertesting.HaveNonZeroEventMeshLocalHash(),
			reconcilertesting.HaveNonZeroWebhookAuthHash(),
		),
	)
	webhookAuthHashAfter := kymaSubscription.Status.Backend.WebhookAuthHash

	////
	// after updating the credentials
	////

	// counts EventMesh mock requests after changing the credentials
	deleteRequestsAfter := emTestEnsemble.eventMeshMock.CountRequests(http.MethodDelete, uri)
	patchRequestsAfter := emTestEnsemble.eventMeshMock.CountRequests(http.MethodPatch, uri)

	// ensure expected EventMesh mock requests
	require.NotEqual(t, webhookAuthHashBefore, webhookAuthHashAfter)
	require.Equal(t, deleteRequestsBefore, deleteRequestsAfter)
	require.Equal(t, patchRequestsBefore, 0)
	require.Equal(t, patchRequestsAfter, 1)

	// cleanup
	require.NoError(t, tearDownSuite())
}
