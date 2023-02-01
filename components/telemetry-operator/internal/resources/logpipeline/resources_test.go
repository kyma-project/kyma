package logpipeline

import (
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestMakeDaemonSet(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	checksum := "foo"
	daemonSet := MakeDaemonSet(name, checksum)

	require.NotNil(t, daemonSet)
	require.Equal(t, daemonSet.Name, name.Name)
	require.Equal(t, daemonSet.Namespace, name.Namespace)
	require.Equal(t, daemonSet.Spec.Selector.MatchLabels, labels())
	require.Equal(t, daemonSet.Spec.Template.ObjectMeta.Labels, labels())
	require.NotEmpty(t, daemonSet.Spec.Template.Spec.Containers[0].EnvFrom)
	require.NotNil(t, daemonSet.Spec.Template.Spec.Containers[0].LivenessProbe, "liveness probe must be defined")
	require.NotNil(t, daemonSet.Spec.Template.Spec.Containers[0].ReadinessProbe, "readiness probe must be defined")

	podSecurityContext := daemonSet.Spec.Template.Spec.SecurityContext
	require.NotNil(t, podSecurityContext, "pod security context must be defined")
	require.False(t, *podSecurityContext.RunAsNonRoot, "must not run as non-root")

	containerSecurityContext := daemonSet.Spec.Template.Spec.Containers[0].SecurityContext
	require.NotNil(t, containerSecurityContext, "container security context must be defined")
	require.False(t, *containerSecurityContext.Privileged, "must not be privileged")
	require.False(t, *containerSecurityContext.AllowPrivilegeEscalation, "must not escalate to privileged")
	require.True(t, *containerSecurityContext.ReadOnlyRootFilesystem, "must use readonly fs")
}

func TestMakeConfigMap(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	cm := MakeConfigMap(name)

	require.NotNil(t, cm)
	require.Equal(t, cm.Name, name.Name)
	require.Equal(t, cm.Namespace, name.Namespace)
	require.NotEmpty(t, cm.Data["custom_parsers.conf"])
	require.NotEmpty(t, cm.Data["fluent-bit.conf"])
	require.NotEmpty(t, cm.Data["loki-labelmap.json"])
}

func TestMakeLuaConfigMap(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	cm := MakeLuaConfigMap(name)

	require.NotNil(t, cm)
	require.NotEqual(t, cm.Name, name.Name)
	require.Equal(t, cm.Name, name.Name+"-luascripts")
	require.Equal(t, cm.Namespace, name.Namespace)
	require.NotEmpty(t, cm.Data["filter-script.lua"])
}

func TestMakeServiceAccount(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	svcAcc := MakeServiceAccount(name)

	require.NotNil(t, svcAcc)
	require.Equal(t, svcAcc.Name, name.Name)
	require.Equal(t, svcAcc.Namespace, name.Namespace)
}

func TestMakeClusterRole(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	clusterRole := MakeClusterRole(name)

	require.NotNil(t, clusterRole)
	require.Equal(t, clusterRole.Name, name.Name)
}

func TestMakeClusterRoleBinding(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	clusterRoleBinding := MakeClusterRoleBinding(name)
	svcAcc := MakeServiceAccount(name)
	clusterRole := MakeClusterRole(name)

	require.NotNil(t, clusterRoleBinding)
	require.Equal(t, clusterRoleBinding.Name, name.Name)
	require.Equal(t, clusterRoleBinding.RoleRef.Name, clusterRole.Name)
	require.Equal(t, clusterRoleBinding.RoleRef.Kind, "ClusterRole")
	require.Equal(t, clusterRoleBinding.Subjects[0].Name, svcAcc.Name)
	require.Equal(t, clusterRoleBinding.Subjects[0].Kind, "ServiceAccount")
}
