package serverless

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func TestFunctionReconciler_buildConfigMap(t *testing.T) {
	tests := []struct {
		name string
		fn   *serverlessv1alpha1.Function
		want corev1.ConfigMap
	}{
		{
			name: "should properly build configmap",
			fn: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fn-ns",
					UID:       "fn-uuid",
					Name:      "function-name",
				},
				Spec: serverlessv1alpha1.FunctionSpec{Source: "fn-source", Deps: ""},
			},
			want: corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "fn-ns",
					GenerateName: "function-name-",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
						serverlessv1alpha1.FunctionNameLabel:      "function-name",
						serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
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
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.buildConfigMap(tt.fn, runtime.GetRuntime(serverlessv1alpha1.Nodejs10))
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_buildDeployment(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}

	tests := []struct {
		name                  string
		args                  args
		rtmCfg                runtime.Config
		expectedVolumesLength int
	}{
		{
			name:                  "spec.template.labels should contain every element from spec.selector.MatchLabels, also python runtime should have volume+volumeMount with emptyDir",
			args:                  args{instance: newFixFunction("ns", "name", 1, 2)},
			rtmCfg:                runtime.GetRuntimeConfig(serverlessv1alpha1.Python38),
			expectedVolumesLength: 1,
		},
		{
			name:                  "node10 runtime does not have volume+volumeMount with emptyDir",
			args:                  args{instance: newFixFunction("ns", "name", 1, 2)},
			rtmCfg:                runtime.GetRuntimeConfig(serverlessv1alpha1.Nodejs10),
			expectedVolumesLength: 0,
		},
		{
			name:                  "node12 runtime does not have volume+volumeMount with emptyDir",
			args:                  args{instance: newFixFunction("ns", "name", 1, 2)},
			rtmCfg:                runtime.GetRuntimeConfig(serverlessv1alpha1.Nodejs12),
			expectedVolumesLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.buildDeployment(tt.args.instance, tt.rtmCfg)

			for key, value := range got.Spec.Selector.MatchLabels {
				g.Expect(got.Spec.Template.Labels[key]).To(gomega.Equal(value))
				g.Expect(got.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
				g.Expect(got.Spec.Template.Spec.Containers[0].Env).To(gomega.ContainElements(tt.rtmCfg.RuntimeEnvs))

				g.Expect(got.Spec.Template.Spec.Volumes).To(gomega.HaveLen(tt.expectedVolumesLength))
				g.Expect(got.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.HaveLen(tt.expectedVolumesLength))
				if tt.expectedVolumesLength > 0 {
					g.Expect(got.Spec.Template.Spec.Volumes[0].Name).To(gomega.Equal(got.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name))
					errs := validation.IsDNS1123Subdomain(got.Spec.Template.Spec.Volumes[0].Name)
					g.Expect(errs).To(gomega.HaveLen(0))
				}
			}
		})
	}
}

func TestFunctionReconciler_buildHorizontalPodAutoscaler(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	type wants struct {
		minReplicas int32
		maxReplicas int32
	}

	nilCase1 := newFixFunction("ns", "name", 2, 2)
	nilCase1.Spec.MinReplicas = nil
	nilCase1.Spec.MaxReplicas = nil
	nilCase2 := newFixFunction("ns", "name", 2, 2)
	nilCase2.Spec.MinReplicas = nil
	nilCase3 := newFixFunction("ns", "name", 2, 2)
	nilCase3.Spec.MaxReplicas = nil

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
			r := &FunctionReconciler{}
			got := r.buildHorizontalPodAutoscaler(tt.args.instance, "foo-bar")

			g.Expect(*got.Spec.MinReplicas).To(gomega.Equal(tt.wants.minReplicas))
			g.Expect(got.Spec.MaxReplicas).To(gomega.Equal(tt.wants.maxReplicas))
		})
	}
}

func TestFunctionReconciler_mergeLabels(t *testing.T) {
	type args struct {
		labelsCollection []map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should work with empty slice",
			args: args{labelsCollection: []map[string]string{}},
			want: map[string]string{},
		},
		{
			name: "should work with 1 map as argument",
			args: args{labelsCollection: []map[string]string{{"key": "value"}}},
			want: map[string]string{"key": "value"},
		},
		{
			name: "should work with multiple maps",
			args: args{labelsCollection: []map[string]string{{"key": "value"}, {"key1": "value1"}, {"key2": "value2"}}},
			want: map[string]string{
				"key":  "value",
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.mergeLabels(tt.args.labelsCollection...)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_internalFunctionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return only 3 correct labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
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

func TestFunctionReconciler_servicePodLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should work on function with no labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionResourceLabel:  serverlessv1alpha1.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should work with function with some labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: serverlessv1alpha1.FunctionSpec{
					Labels: map[string]string{
						"test-some": "test-label",
					},
				}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionResourceLabel:  serverlessv1alpha1.FunctionResourceLabelDeploymentValue,
				"test-some":                               "test-label",
			},
		},
		{
			name: "Should not overwrite internal labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: serverlessv1alpha1.FunctionSpec{
					Labels: map[string]string{
						"test-some":                              "test-label",
						serverlessv1alpha1.FunctionResourceLabel: "job",
						serverlessv1alpha1.FunctionNameLabel:     "some-other-name",
					},
				}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionResourceLabel:  serverlessv1alpha1.FunctionResourceLabelDeploymentValue,
				"test-some":                               "test-label",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.podLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_functionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return fn labels + 3 internal ones",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
						Labels: map[string]string{
							"some-key": "whatever-value",
						}},
				},
			},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				"some-key":                                "whatever-value",
			},
		}, {
			name: "should return 3 internal ones if there's no labels on fn",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
				}},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
			},
		},
		{
			name: "should not be able to override our internal labels",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
						Labels: map[string]string{
							serverlessv1alpha1.FunctionUUIDLabel: "whatever-value",
						}},
				},
			},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue,
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.functionLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_buildJob(t *testing.T) {
	g := gomega.NewWithT(t)
	fnName := "my-function"
	cmName := "test-configmap"
	rtmCfg := runtime.Config{
		Runtime:                 "my-runtime",
		DependencyFile:          "deps.txt",
		FunctionFile:            "function.abap",
		DockerfileConfigMapName: "dockerfile-runtime-abap",
		RuntimeEnvs:             nil,
	}

	expectedVolumeMounts := []corev1.VolumeMount{
		{Name: "sources", MountPath: "/workspace/src/deps.txt", SubPath: FunctionDepsKey, ReadOnly: true},
		{Name: "sources", MountPath: "/workspace/src/function.abap", SubPath: FunctionSourceKey, ReadOnly: true},
		{Name: "runtime", MountPath: "/workspace/Dockerfile", SubPath: "Dockerfile", ReadOnly: true},
		{Name: "credentials", MountPath: "/docker", ReadOnly: true}}
	expectedVolumes := []expectedVolume{
		{name: "sources", localObjectReference: cmName},
		{name: "runtime", localObjectReference: rtmCfg.DockerfileConfigMapName}}

	instance := serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{Name: fnName},
		Spec:       serverlessv1alpha1.FunctionSpec{},
	}

	r := FunctionReconciler{}
	// when
	job := r.buildJob(&instance, rtmCfg, cmName)

	// then
	g.Expect(job.ObjectMeta.GenerateName).To(gomega.Equal("my-function-build-"))
	assertVolumes(g, job.Spec.Template.Spec.Volumes, expectedVolumes)

	g.Expect(job.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
	g.Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.HaveLen(4))
	g.Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(gomega.Equal(expectedVolumeMounts))
}

type expectedVolume struct {
	name                 string
	localObjectReference string
}

func assertVolumes(g *gomega.WithT, actual []corev1.Volume, expected []expectedVolume) {
	for _, expVol := range expected {
		found := false
		for _, actualVol := range actual {
			if actualVol.Name == expVol.name {
				found = true
			}
		}
		g.Expect(found).To(gomega.BeTrue(), "Volume with name: %s, not found", expVol.name)
	}
}
