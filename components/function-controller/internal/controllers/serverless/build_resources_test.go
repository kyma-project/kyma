package serverless

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	rtmNodeJS16 = fnRuntime.GetRuntimeConfig(serverlessv1alpha2.NodeJs16)
	rtmNodeJS18 = fnRuntime.GetRuntimeConfig(serverlessv1alpha2.NodeJs18)
	rtmPython39 = fnRuntime.GetRuntimeConfig(serverlessv1alpha2.Python39)
)

func TestFunctionReconciler_buildConfigMap(t *testing.T) {
	tests := []struct {
		name string
		fn   *serverlessv1alpha2.Function
		want corev1.ConfigMap
	}{
		{
			name: "should properly build configmap",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fn-ns",
					UID:       "fn-uuid",
					Name:      "function-name",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
				},
			},
			want: corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "fn-ns",
					GenerateName: "function-name-",
					Labels: map[string]string{
						serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
						serverlessv1alpha2.FunctionNameLabel:      "function-name",
						serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
					},
				},
				Data: map[string]string{
					"source":       "fn-source",
					"dependencies": "{}",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			s := systemState{
				//TODO: https://github.com/kyma-project/kyma/issues/14079
				instance: *tt.fn,
			}
			got := s.buildConfigMap()
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_buildDeployment(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha2.Function
	}
	rtmCfg := runtime.GetRuntimeConfig(serverlessv1alpha2.NodeJs18)
	resourceCfg := fixResources()

	tests := []struct {
		name string
		args args
	}{
		{
			name: "deployment should contain needed elements",
			args: args{
				instance: newFixFunction("ns", "name", 1, 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				//TODO https://github.com/kyma-project/kyma/issues/14079
				instance: *tt.args.instance,
			}

			got := s.buildDeployment(buildDeploymentArgs{}, resourceCfg)

			// deployment selector labels are equal with pod labels
			for key, value := range got.Spec.Selector.MatchLabels {
				g.Expect(got.Spec.Template.Labels[key]).To(gomega.Equal(value))
			}
			g.Expect(got.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
			g.Expect(got.Spec.Template.Spec.Containers[0].Env).To(gomega.ContainElements(rtmCfg.RuntimeEnvs))

			// pod labels & annotations
			g.Expect(got.Spec.Template.ObjectMeta.Labels).To(gomega.HaveLen(4 + 3))
			g.Expect(got.Spec.Template.ObjectMeta.Labels).To(gomega.HaveKeyWithValue("foo", "bar"))
			g.Expect(got.Spec.Template.ObjectMeta.Labels).To(gomega.HaveKeyWithValue(testBindingLabel1, "foobar"))
			g.Expect(got.Spec.Template.ObjectMeta.Labels).To(gomega.HaveKeyWithValue(testBindingLabel2, testBindingLabelValue))

			g.Expect(got.Spec.Template.ObjectMeta.Annotations).To(gomega.HaveLen(1 + 1))
			g.Expect(got.Spec.Template.ObjectMeta.Annotations).To(gomega.HaveKeyWithValue("foo", "bar"))

			// volumes
			const expectedVolumeCount = 3
			g.Expect(got.Spec.Template.Spec.Volumes).To(gomega.HaveLen(expectedVolumeCount))
			g.Expect(got.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.HaveLen(expectedVolumeCount))

			for i := 0; i < expectedVolumeCount; i++ {
				g.Expect(got.Spec.Template.Spec.Volumes[i].Name).To(gomega.Equal(got.Spec.Template.Spec.Containers[0].VolumeMounts[i].Name))
				errs := validation.IsDNS1123Subdomain(got.Spec.Template.Spec.Volumes[i].Name)
				g.Expect(errs).To(gomega.HaveLen(0))

				checkSecretVolume(g, tt.args.instance.Spec.SecretMounts,
					got.Spec.Template.Spec.Volumes[i], got.Spec.Template.Spec.Containers[0].VolumeMounts[i])
			}

			g.Expect(got.Spec.Template.Spec.Containers[0].StartupProbe.SuccessThreshold).To(gomega.BeEquivalentTo(1), "documentation states that this value has to be set to 1")
		})
	}
}

func checkSecretVolume(g *gomega.WithT, secretMounts []serverlessv1alpha2.SecretMount, volume corev1.Volume, volumeMount corev1.VolumeMount) {
	if volume.Secret != nil {
		var matchingSecretMount *serverlessv1alpha2.SecretMount = nil
		for _, secretMount := range secretMounts {
			if secretMount.SecretName == volume.Secret.SecretName {
				matchingSecretMount = secretMount.DeepCopy()
				break
			}
		}
		g.Expect(matchingSecretMount).ToNot(gomega.BeNil())
		g.Expect(volumeMount.MountPath).To(gomega.Equal(matchingSecretMount.MountPath))
	}
}

func TestFunctionReconciler_buildDeploymentWithResources(t *testing.T) {
	resourceCfg := fixResources()
	resources := resourceCfg.Presets.ToResourceRequirements()
	python39Resources := resourceCfg.RuntimePresets[string(serverlessv1alpha2.NodeJs18)].ToResourceRequirements()

	customResources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("378m"),
			corev1.ResourceMemory: resource.MustParse("378Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("157m"),
			corev1.ResourceMemory: resource.MustParse("157Mi"),
		},
	}

	type args struct {
		instance          *serverlessv1alpha2.Function
		expectedResources corev1.ResourceRequirements
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "deployment should use set resources preset",
			args: args{
				instance:          newFixFunctionWithFunctionResourceProfile("ns", "name", "M"),
				expectedResources: resources["M"],
			},
		},
		{
			name: "deployment should use default resources preset",
			args: args{
				instance:          newFixFunction("ns", "name", 1, 2),
				expectedResources: resources["L"],
			},
		},
		{
			name: "deployment should use default runtime preset",
			args: args{
				instance:          newFixFunctionWithRuntime("ns", "name", serverlessv1alpha2.Python39),
				expectedResources: python39Resources["S"],
			},
		},
		{
			name: "deployment should use default runtime preset",
			args: args{
				instance:          newFixFunctionWithCustomFunctionResource("ns", "name", customResources),
				expectedResources: *customResources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				instance: *tt.args.instance,
			}

			got := s.buildDeployment(buildDeploymentArgs{}, resourceCfg)

			g.Expect(got.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
			g.Expect(got.Spec.Template.Spec.Containers[0].Resources).To(gomega.Equal(tt.args.expectedResources))
		})
	}
}

func TestFunctionReconciler_buildHorizontalPodAutoscaler(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha2.Function
	}
	type wants struct {
		minReplicas int32
		maxReplicas int32
	}

	nilCase1 := newFixFunction("ns", "name", 2, 2)
	nilCase1.Spec.ScaleConfig.MinReplicas = nil
	nilCase1.Spec.ScaleConfig.MaxReplicas = nil
	nilCase2 := newFixFunction("ns", "name", 2, 2)
	nilCase2.Spec.ScaleConfig.MinReplicas = nil
	nilCase3 := newFixFunction("ns", "name", 2, 2)
	nilCase3.Spec.ScaleConfig.MaxReplicas = nil

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values - 0 in spec",
			args: args{instance: newFixFunction("ns", "name", 0, 0)},
			wants: wants{
				minReplicas: 1,
				maxReplicas: 1,
			},
		},
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values - nil in spec",
			args: args{instance: nilCase1},
			wants: wants{
				minReplicas: 1,
				maxReplicas: 1,
			},
		},
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values, when min is set to 0",
			args: args{instance: newFixFunction("ns", "name", 2, 0)},
			wants: wants{
				minReplicas: 2,
				maxReplicas: 2,
			},
		},
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values, when min is nil",
			args: args{instance: nilCase2},
			wants: wants{
				minReplicas: 1,
				maxReplicas: 2,
			},
		},
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values, when max is set to 0",
			args: args{instance: newFixFunction("ns", "name", 0, 3)},
			wants: wants{
				minReplicas: 1,
				maxReplicas: 3,
			},
		},
		{
			name: "spec.minReplicas and spec.maxReplicas fields should contain fixed (from Function spec or default) values, when max is nil",
			args: args{instance: nilCase3},
			wants: wants{
				minReplicas: 2,
				maxReplicas: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				instance: *tt.args.instance,
				deployments: v1.DeploymentList{
					Items: []v1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{},
						},
					},
				},
			}

			got := s.buildHorizontalPodAutoscaler(0)

			g.Expect(*got.Spec.MinReplicas).To(gomega.Equal(tt.wants.minReplicas))
			g.Expect(got.Spec.MaxReplicas).To(gomega.Equal(tt.wants.maxReplicas))
		})
	}
}

func TestFunctionReconciler_internalFunctionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return only 3 correct labels",
			args: args{instance: &serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.internalFunctionLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(3))
		})
	}
}

func TestFunctionReconciler_functionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return fn labels + 3 internal ones",
			args: args{
				instance: &serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
						Labels: map[string]string{
							"some-key": "whatever-value",
						}},
				},
			},
			want: map[string]string{
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
				"some-key":                                "whatever-value",
			},
		}, {
			name: "should return 3 internal ones if there's no labels on fn",
			args: args{
				instance: &serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
				}},
			want: map[string]string{
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
			},
		},
		{
			name: "should not be able to override our internal labels",
			args: args{
				instance: &serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
						Labels: map[string]string{
							serverlessv1alpha2.FunctionUUIDLabel: "whatever-value",
						}},
				},
			},
			want: map[string]string{
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				//TODO https://github.com/kyma-project/kyma/issues/14079
				instance: *tt.args.instance,
			}
			got := s.functionLabels()
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_buildJob(t *testing.T) {
	g := gomega.NewWithT(t)

	// GIVEN
	cmName := "test-configmap"

	dockerCfg := DockerConfig{
		ActiveRegistryConfigSecretName: "docker-secret-name",
	}
	//nolint:gosec
	packageRegistryConfigSecretName := "pkg-config-secret"

	testCases := []struct {
		Name                 string
		Runtime              serverlessv1alpha2.Runtime
		ExpectedVolumesLen   int
		ExpectedVolumes      []expectedVolume
		ExpectedMountsLen    int
		ExpectedVolumeMounts []corev1.VolumeMount
	}{
		{
			Name:               "Success Node16",
			Runtime:            serverlessv1alpha2.NodeJs16,
			ExpectedVolumesLen: 4,
			ExpectedVolumes: []expectedVolume{
				{name: "sources", localObjectReference: cmName},
				{name: "runtime", localObjectReference: rtmNodeJS16.DockerfileConfigMapName},
				{name: "credentials", localObjectReference: dockerCfg.ActiveRegistryConfigSecretName},
				{name: "registry-config", localObjectReference: packageRegistryConfigSecretName},
			},
			ExpectedMountsLen: 5,
			ExpectedVolumeMounts: []corev1.VolumeMount{
				{Name: "sources", MountPath: "/workspace/src/package.json", SubPath: FunctionDepsKey, ReadOnly: true},
				{Name: "sources", MountPath: "/workspace/src/handler.js", SubPath: FunctionSourceKey, ReadOnly: true},
				{Name: "runtime", MountPath: "/workspace/Dockerfile", SubPath: "Dockerfile", ReadOnly: true},
				{Name: "credentials", MountPath: "/docker", ReadOnly: true},
				{Name: "registry-config", MountPath: "/workspace/registry-config/.npmrc", SubPath: ".npmrc", ReadOnly: true},
			},
		},
		{
			Name:               "Success Node18",
			Runtime:            serverlessv1alpha2.NodeJs18,
			ExpectedVolumesLen: 4,
			ExpectedVolumes: []expectedVolume{
				{name: "sources", localObjectReference: cmName},
				{name: "runtime", localObjectReference: rtmNodeJS18.DockerfileConfigMapName},
				{name: "credentials", localObjectReference: dockerCfg.ActiveRegistryConfigSecretName},
				{name: "registry-config", localObjectReference: packageRegistryConfigSecretName},
			},
			ExpectedMountsLen: 5,
			ExpectedVolumeMounts: []corev1.VolumeMount{
				{Name: "sources", MountPath: "/workspace/src/package.json", SubPath: FunctionDepsKey, ReadOnly: true},
				{Name: "sources", MountPath: "/workspace/src/handler.js", SubPath: FunctionSourceKey, ReadOnly: true},
				{Name: "runtime", MountPath: "/workspace/Dockerfile", SubPath: "Dockerfile", ReadOnly: true},
				{Name: "credentials", MountPath: "/docker", ReadOnly: true},
				{Name: "registry-config", MountPath: "/workspace/registry-config/.npmrc", SubPath: ".npmrc", ReadOnly: true},
			},
		},
		{
			Name:               "Success Python39",
			Runtime:            serverlessv1alpha2.Python39,
			ExpectedVolumesLen: 4,
			ExpectedVolumes: []expectedVolume{
				{name: "sources", localObjectReference: cmName},
				{name: "runtime", localObjectReference: rtmPython39.DockerfileConfigMapName},
				{name: "credentials", localObjectReference: dockerCfg.ActiveRegistryConfigSecretName},
				{name: "registry-config", localObjectReference: packageRegistryConfigSecretName},
			},
			ExpectedMountsLen: 5,
			ExpectedVolumeMounts: []corev1.VolumeMount{
				{Name: "sources", MountPath: "/workspace/src/requirements.txt", SubPath: FunctionDepsKey, ReadOnly: true},
				{Name: "sources", MountPath: "/workspace/src/handler.py", SubPath: FunctionSourceKey, ReadOnly: true},
				{Name: "runtime", MountPath: "/workspace/Dockerfile", SubPath: "Dockerfile", ReadOnly: true},
				{Name: "credentials", MountPath: "/docker", ReadOnly: true},
				{Name: "registry-config", MountPath: "/workspace/registry-config/pip.conf", SubPath: "pip.conf", ReadOnly: true},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			functionName := "my-function"
			s := systemState{
				instance: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{Name: functionName},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: testCase.Runtime,
						Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{}},
					},
				},
			}

			// when
			job := s.buildJob(cmName, cfg{
				docker: dockerCfg,
				fn: FunctionConfig{
					PackageRegistryConfigSecretName: "pkg-config-secret",
				},
			})

			// then
			g.Expect(job.ObjectMeta.GenerateName).To(gomega.Equal(fmt.Sprintf("%s-build-", functionName)))
			g.Expect(job.Spec.Template.Spec.Volumes).To(gomega.HaveLen(testCase.ExpectedVolumesLen))
			assertVolumes(g, job.Spec.Template.Spec.Volumes, testCase.ExpectedVolumes)

			g.Expect(job.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
			g.Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.HaveLen(testCase.ExpectedMountsLen))
			g.Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.Equal(testCase.ExpectedVolumeMounts))
		})
	}
}

func TestFunctionReconciler_buildJobWithResources(t *testing.T) {
	resourceCfg := fixResources()
	cfg := cfg{
		fn: FunctionConfig{
			ResourceConfig: ResourceConfig{BuildJob: BuildJobResourceConfig{resourceCfg}},
		},
	}

	resources := resourceCfg.Presets.ToResourceRequirements()
	python39Resources := resourceCfg.RuntimePresets[string(serverlessv1alpha2.NodeJs18)].ToResourceRequirements()

	customResources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("378m"),
			corev1.ResourceMemory: resource.MustParse("378Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("157m"),
			corev1.ResourceMemory: resource.MustParse("157Mi"),
		},
	}

	type args struct {
		instance          *serverlessv1alpha2.Function
		expectedResources corev1.ResourceRequirements
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "job should have resources profile preset",
			args: args{
				instance:          newFixFunctionWithBuildResourceProfile("ns", "name", "M"),
				expectedResources: resources["M"],
			},
		},
		{
			name: "job should have default resources preset",
			args: args{
				instance:          newFixFunction("ns", "name", 1, 2),
				expectedResources: resources["L"],
			},
		},
		{
			name: "job should have default runtime preset",
			args: args{
				instance:          newFixFunctionWithRuntime("ns", "name", serverlessv1alpha2.Python39),
				expectedResources: python39Resources["S"],
			},
		},
		{
			name: "job should have custom resources",
			args: args{
				instance:          newFixFunctionWithCustomBuildResource("ns", "name", customResources),
				expectedResources: *customResources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				instance: *tt.args.instance,
			}

			got := s.buildJob("configmap", cfg)

			g.Expect(got.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
			g.Expect(got.Spec.Template.Spec.Containers[0].Resources).To(gomega.Equal(tt.args.expectedResources))
		})
	}
}

type expectedVolume struct {
	name                 string
	localObjectReference string
}

func assertVolumes(g *gomega.WithT, actual []corev1.Volume, expected []expectedVolume) {
	for _, expVol := range expected {
		found := false
		for _, actualVol := range actual {
			if actualVol.Name == expVol.name &&
				(actualVol.Secret != nil && actualVol.Secret.SecretName == expVol.localObjectReference) ||
				(actualVol.ConfigMap != nil && actualVol.ConfigMap.LocalObjectReference.Name == expVol.localObjectReference) {
				found = true
			}
		}
		g.Expect(found).To(gomega.BeTrue(), "Volume with name: %s, referencing object: %s not found", expVol.name, expVol.localObjectReference)
	}
}

func fixResources() Resources {
	return Resources{
		Presets: map[string]Resource{
			"L": {
				RequestCPU:    Quantity{resource.MustParse("100m")},
				RequestMemory: Quantity{resource.MustParse("100Mi")},
				LimitCPU:      Quantity{resource.MustParse("200m")},
				LimitMemory:   Quantity{resource.MustParse("200Mi")},
			},
			"M": {
				RequestCPU:    Quantity{resource.MustParse("50m")},
				RequestMemory: Quantity{resource.MustParse("50Mi")},
				LimitCPU:      Quantity{resource.MustParse("100m")},
				LimitMemory:   Quantity{resource.MustParse("100Mi")},
			},
		},
		DefaultPreset: "L",
		RuntimePresets: map[string]Preset{
			string(serverlessv1alpha2.Python39): map[string]Resource{
				"S": {
					RequestCPU:    Quantity{resource.MustParse("135m")},
					RequestMemory: Quantity{resource.MustParse("135Mi")},
					LimitCPU:      Quantity{resource.MustParse("246m")},
					LimitMemory:   Quantity{resource.MustParse("246Mi")},
				},
			},
		},
	}
}
