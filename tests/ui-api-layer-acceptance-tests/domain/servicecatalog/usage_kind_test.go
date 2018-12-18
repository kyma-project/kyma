// +build acceptance

package servicecatalog

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
)

type usageKind struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Group       string `json:"group"`
	Kind        string `json:"kind"`
	Version     string `json:"version"`
}

type usageKindResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type usageKindsResponse struct {
	UsageKinds []usageKind `json:"usageKinds"`
}

type usageKindResourcesResponse struct {
	UsageKindResources []usageKindResource `json:"usageKindResources"`
}

const (
	usageKindName      = "usage-kind-test"
	usageKindNamespace = "usage-kind-ns-test"
)

func TestUsageKind(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	client, cfg, err := client.NewClientWithConfig()
	require.NoError(t, err)

	bucClient, err := versioned.NewForConfig(cfg)
	require.NoError(t, err)

	deployClient, err := v1beta1.NewForConfig(cfg)
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = client.Namespaces().Create(fixNamespace(usageKindNamespace))
	require.NoError(t, err)

	t.Log("Creating UsageKind...")
	_, err = bucClient.ServicecatalogV1alpha1().UsageKinds().Create(fixUsageKind())
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting UsageKind...")
		err = bucClient.ServicecatalogV1alpha1().UsageKinds().Delete(usageKindName, &metav1.DeleteOptions{})
		assert.NoError(t, err)

		t.Log("Deleting namespace...")
		err = client.Namespaces().Delete(usageKindNamespace, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	t.Log("Querying for usageKinds...")
	var usageKindsResponse usageKindsResponse
	waiter.WaitAtMost(func() (bool, error) {
		err = c.Do(fixUsageKindsQuery(), &usageKindsResponse)
		if err != nil {
			return false, err
		}
		return usageKindExists(usageKindsResponse.UsageKinds, fixUsageKindResponse()), nil
	}, time.Second*5)

	t.Log("Creating resource for UsageKind...")
	_, err = deployClient.Deployments(usageKindNamespace).Create(fixDeployment())
	require.NoError(t, err)

	t.Log("Creating resource for UsageKind with an owner reference...")
	_, err = deployClient.Deployments(usageKindNamespace).Create(fixDeploymentWithOwnerReference())
	require.NoError(t, err)

	t.Log("Querying for usageKindResources...")
	var usageKindResourcesResponse usageKindResourcesResponse
	waiter.WaitAtMost(func() (bool, error) {
		err = c.Do(fixUsageKindsResourcesQuery(), &usageKindResourcesResponse)
		if err != nil {
			return false, err
		}
		if !usageKindResourceExists(usageKindResourcesResponse.UsageKindResources, fixUsageKindResourceResponse()) {
			return false, nil
		}
		if usageKindResourceExists(usageKindResourcesResponse.UsageKindResources, fixUsageKindResourcesShouldNotContainResponse()) {
			return false, nil
		}
		return true, nil
	}, time.Second*5)
}

func fixUsageKindsQuery() *graphql.Request {
	query := `query {
				usageKinds {
					name
					displayName
					kind
					group
					version
				}
			}`
	req := graphql.NewRequest(query)
	return req
}

func fixUsageKindsResourcesQuery() *graphql.Request {
	query := `query($usageKind: String!, $environment: String!) {
				usageKindResources(usageKind: $usageKind, environment: $environment) {
					name
					namespace
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("usageKind", usageKindName)
	req.SetVar("environment", usageKindNamespace)
	return req
}

func fixUsageKind() *v1alpha1.UsageKind {
	return &v1alpha1.UsageKind{
		ObjectMeta: metav1.ObjectMeta{
			Name: usageKindName,
		},
		Spec: v1alpha1.UsageKindSpec{
			Resource: &v1alpha1.ResourceReference{
				Kind:    "Deployment",
				Version: "v1beta1",
				Group:   "apps",
			},
			DisplayName: "Deploys",
			LabelsPath:  "spec.deployment.spec.template.metadata.labels",
		},
	}
}

func fixDeployment() *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "usage-kind-fix-a",
			Namespace: usageKindNamespace,
			Labels:    fixLabels(),
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: fixLabels(),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fixLabels(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "test",
							Image:           "test",
							ImagePullPolicy: "Never",
						},
					},
				},
			},
		},
	}
}

func fixDeploymentWithOwnerReference() *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "usage-kind-fix-b",
			Namespace: usageKindNamespace,
			Labels:    fixLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       "owner",
					Kind:       "dummy",
					APIVersion: "v8",
					UID:        "123",
				},
			},
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: fixLabels(),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fixLabels(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "test",
							Image:           "test",
							ImagePullPolicy: "Never",
						},
					},
				},
			},
		},
	}
}

func fixNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixLabels() map[string]string {
	labels := make(map[string]string)
	labels["label"] = "test"
	return labels
}

func fixUsageKindResponse() usageKind {
	return usageKind{
		Name:        usageKindName,
		DisplayName: "Deploys",
		Group:       "apps",
		Kind:        "Deployment",
		Version:     "v1beta1",
	}
}

func fixUsageKindResourceResponse() usageKindResource {
	return usageKindResource{
		Name:      "usage-kind-fix-a",
		Namespace: usageKindNamespace,
	}
}

func fixUsageKindResourcesShouldNotContainResponse() usageKindResource {
	return usageKindResource{
		Name:      "usage-kind-fix-b",
		Namespace: usageKindNamespace,
	}
}

func usageKindExists(items []usageKind, expected usageKind) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func usageKindResourceExists(items []usageKindResource, expected usageKindResource) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
