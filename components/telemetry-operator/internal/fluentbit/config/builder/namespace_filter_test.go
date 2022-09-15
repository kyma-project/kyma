package builder

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateNamespaceGrepFilterIncludeNamespaces(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces: v1alpha1.InputNamespaces{
					Include: []string{"namespace1", "namespace2"},
				},
			}},
		},
	}

	expected := `[FILTER]
    name  grep
    match logpipeline1.*
    regex $kubernetes['namespace_name'] namespace1|namespace2

`
	actual := createNamespaceGrepFilter(logPipeline)
	require.Equal(t, expected, actual)
}

func TestCreateNamespaceGrepFilterExcludeNamespaces(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces: v1alpha1.InputNamespaces{
					Exclude: []string{"namespace1", "namespace2"}},
			}},
		},
	}

	expected := `[FILTER]
    name    grep
    match   logpipeline1.*
    exclude $kubernetes['namespace_name'] namespace1|namespace2

`
	actual := createNamespaceGrepFilter(logPipeline)
	require.Equal(t, expected, actual)
}

func TestCreateNamespaceGrepFilterSystemNamespacesExcluded(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{ObjectMeta: metav1.ObjectMeta{Name: "logpipeline1"}}

	expected := `[FILTER]
    name    grep
    match   logpipeline1.*
    exclude $kubernetes['namespace_name'] kyma-system|kyma-integration|kube-system|istio-system|compass-system

`
	actual := createNamespaceGrepFilter(logPipeline)
	require.Equal(t, expected, actual)
}

func TestCreateNamespaceGrepFilterSystemNamespacesIncluded(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces: v1alpha1.InputNamespaces{
					System: true},
			}},
		},
	}

	actual := createNamespaceGrepFilter(logPipeline)
	require.Equal(t, "", actual)
}
