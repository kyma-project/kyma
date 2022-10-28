package test

import (
	"context"
	"log"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getSubscriptionAssert fetches a subscription using the lookupKey and allows making assertions on it.
func getSubscriptionAssert(g *gomega.GomegaWithT, ctx context.Context, subscription *eventingv1alpha2.Subscription) gomega.AsyncAssertion {
	return g.Eventually(func() *eventingv1alpha2.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := emTestEnsemble.k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("fetch subscription %s failed: %v", lookupKey.String(), err)
			return &eventingv1alpha2.Subscription{}
		}
		return subscription
	}, bigTimeOut, bigPollingInterval)
}

// getAPIRuleForASvcAssert fetches an apiRule for a given service and allows making assertions on it.
func getAPIRuleForASvcAssert(g *gomega.GomegaWithT, ctx context.Context, svc *corev1.Service) gomega.AsyncAssertion {
	return g.Eventually(func() apigatewayv1beta1.APIRule {
		apiRules, err := getAPIRulesList(ctx, svc)
		g.Expect(err).Should(gomega.BeNil())
		return filterAPIRulesForASvc(apiRules, svc)
	}, smallTimeOut, smallPollingInterval)
}

// getAPIRuleAssert fetches an apiRule and allows making assertions on it.
func getAPIRuleAssert(g *gomega.GomegaWithT, ctx context.Context, apiRule *apigatewayv1beta1.APIRule) gomega.AsyncAssertion {
	return g.Eventually(func() apigatewayv1beta1.APIRule {
		fetchedApiRule, err := getAPIRule(ctx, apiRule)
		if err != nil {
			return apigatewayv1beta1.APIRule{}
		}
		return *fetchedApiRule
	}, bigTimeOut, bigPollingInterval)
}
