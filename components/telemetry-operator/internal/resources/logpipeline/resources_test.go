package logpipeline

import (
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strconv"
	"testing"
)

func TestMakeDaemonSet(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	checksum := "foo"
	ds := DaemonSetConfig{
		FluentBitImage:              "foo-fluenbit",
		FluentBitConfigPrepperImage: "foo-configprepper",
		ExporterImage:               "foo-exporter",
		PriorityClassName:           "foo-prio-class",
		CPULimit:                    resource.MustParse(".25"),
		MemoryLimit:                 resource.MustParse("400Mi"),
		CPURequest:                  resource.MustParse(".1"),
		MemoryRequest:               resource.MustParse("100Mi"),
	}

	expectedVolMounts := []corev1.VolumeMount{
		{MountPath: "/fluent-bit/etc", Name: "shared-fluent-bit-config"},
		{MountPath: "/fluent-bit/etc/fluent-bit.conf", Name: "config", SubPath: "fluent-bit.conf"},
		{MountPath: "/fluent-bit/etc/dynamic/", Name: "dynamic-config"},
		{MountPath: "/fluent-bit/etc/dynamic-parsers/", Name: "dynamic-parsers-config"},
		{MountPath: "/fluent-bit/etc/custom_parsers.conf", Name: "config", SubPath: "custom_parsers.conf"},
		{MountPath: "/fluent-bit/etc/loki-labelmap.json", Name: "config", SubPath: "loki-labelmap.json"},
		{MountPath: "/fluent-bit/scripts/filter-script.lua", Name: "luascripts", SubPath: "filter-script.lua"},
		{MountPath: "/var/log", Name: "varlog"},
		{MountPath: "/var/lib/docker/containers", Name: "varlibdockercontainers", ReadOnly: true},
		{MountPath: "/data", Name: "varfluentbit"},
		{MountPath: "/files", Name: "dynamic-files"},
	}
	daemonSet := MakeDaemonSet(name, checksum, ds)

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

	volMounts := daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts
	require.True(t, reflect.DeepEqual(volMounts, expectedVolMounts))
}

func TestMakeMetricsService(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	service := MakeMetricsService(name)

	require.NotNil(t, service)
	require.Equal(t, service.Name, "telemetry-fluent-bit-metrics")
	require.Equal(t, service.Namespace, name.Namespace)
	require.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	require.Len(t, service.Spec.Ports, 1)

	require.Contains(t, service.Annotations, "prometheus.io/scrape")
	require.Contains(t, service.Annotations, "prometheus.io/port")
	require.Contains(t, service.Annotations, "prometheus.io/scheme")
	require.Contains(t, service.Annotations, "prometheus.io/path")

	port, err := strconv.Atoi(service.Annotations["prometheus.io/port"])
	require.NoError(t, err)
	require.Equal(t, int32(port), service.Spec.Ports[0].Port)
}

func TestMakeExporterMetricsService(t *testing.T) {
	name := types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "kyma-system"}
	service := MakeExporterMetricsService(name)

	require.NotNil(t, service)
	require.Equal(t, service.Name, "telemetry-fluent-bit-exporter-metrics")
	require.Equal(t, service.Namespace, name.Namespace)
	require.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	require.Len(t, service.Spec.Ports, 1)

	require.Contains(t, service.Annotations, "prometheus.io/scrape")
	require.Contains(t, service.Annotations, "prometheus.io/port")
	require.Contains(t, service.Annotations, "prometheus.io/scheme")

	port, err := strconv.Atoi(service.Annotations["prometheus.io/port"])
	require.NoError(t, err)
	require.Equal(t, int32(port), service.Spec.Ports[0].Port)
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
	expectedRules := []v1.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{""}, Resources: []string{"namespaces", "pods"}}}

	require.NotNil(t, clusterRole)
	require.Equal(t, clusterRole.Name, name.Name)
	require.Equal(t, clusterRole.Rules, expectedRules)
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
