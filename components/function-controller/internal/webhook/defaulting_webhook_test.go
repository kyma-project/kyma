package webhook

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDefaultingWebHook_Handle(t *testing.T) {
	type fields struct {
		configv1alpha1 *serverlessv1alpha1.DefaultingConfig
		configv1alpha2 *serverlessv1alpha2.DefaultingConfig
		client         ctrlclient.Client
		decoder        *admission.Decoder
	}
	type args struct {
		ctx context.Context
		req admission.Request
	}
	type want struct {
		operationsCount int
		statusCode      int32
	}
	scheme := runtime.NewScheme()
	_ = serverlessv1alpha2.AddToScheme(scheme)
	decoder, err := admission.NewDecoder(scheme)
	require.NoError(t, err)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Set function defaults successfully v1alpha2",
			fields: fields{
				configv1alpha2: &serverlessv1alpha2.DefaultingConfig{
					Function: serverlessv1alpha2.FunctionDefaulting{
						Replicas: serverlessv1alpha2.FunctionReplicasDefaulting{
							DefaultPreset: "S",
							Presets: map[string]serverlessv1alpha2.ReplicasPreset{
								"S": {
									Min: int32(1),
									Max: int32(1),
								},
							},
						},
						Resources: serverlessv1alpha2.FunctionResourcesDefaulting{
							DefaultPreset: "S",
							Presets: map[string]serverlessv1alpha2.ResourcesPreset{
								"S": {
									RequestCPU:    "100m",
									RequestMemory: "128Mi",
									LimitCPU:      "200m",
									LimitMemory:   "256Mi",
								},
							},
						},
					},
					BuildJob: serverlessv1alpha2.BuildJobDefaulting{
						Resources: serverlessv1alpha2.BuildJobResourcesDefaulting{
							DefaultPreset: "normal",
							Presets: map[string]serverlessv1alpha2.ResourcesPreset{
								"normal": {
									RequestCPU:    "700m",
									RequestMemory: "700Mi",
									LimitCPU:      "1100m",
									LimitMemory:   "1100Mi",
								},
							},
						},
					},
				},
				client:  fake.NewClientBuilder().Build(),
				decoder: decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: "Function", Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`{
								"apiVersion": "serverless.kyma-project.io/v1alpha2",
								"kind": "Function",
								"metadata": {
									"labels": {
										"serverless.kyma-project.io/function-resources-preset": "S"
									},
									"name": "testfunc",
									"namespace": "default"
								},
								"spec": {
									"runtime": "python39",
									"source": {
										"inline": {
											"source": "def main(event, context):\n  return \"hello world\"\n"
										}
									}
								}
							}`),
						},
					},
				},
			},
			want: want{
				// 4 patch operations added
				// add /spec/sources/inline/dependencies
				// add /spec/scaleConfig
				// add /status
				// add /metadata/creationTimestamp
				operationsCount: 4,
			},
		},
		{
			name: "Set function defaults successfully v1alpha1",
			fields: fields{
				configv1alpha1: &serverlessv1alpha1.DefaultingConfig{
					Function: serverlessv1alpha1.FunctionDefaulting{
						Replicas: serverlessv1alpha1.FunctionReplicasDefaulting{
							DefaultPreset: "S",
							Presets: map[string]serverlessv1alpha1.ReplicasPreset{
								"S": {
									Min: int32(1),
									Max: int32(1),
								},
							},
						},
						Resources: serverlessv1alpha1.FunctionResourcesDefaulting{
							DefaultPreset: "S",
							Presets: map[string]serverlessv1alpha1.ResourcesPreset{
								"S": {
									RequestCPU:    "100m",
									RequestMemory: "128Mi",
									LimitCPU:      "200m",
									LimitMemory:   "256Mi",
								},
							},
						},
					},
					BuildJob: serverlessv1alpha1.BuildJobDefaulting{
						Resources: serverlessv1alpha1.BuildJobResourcesDefaulting{
							DefaultPreset: "normal",
							Presets: map[string]serverlessv1alpha1.ResourcesPreset{
								"normal": {
									RequestCPU:    "700m",
									RequestMemory: "700Mi",
									LimitCPU:      "1100m",
									LimitMemory:   "1100Mi",
								},
							},
						},
					},
				},
				client:  fake.NewClientBuilder().Build(),
				decoder: decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: serverlessv1alpha1.FunctionKind, Version: serverlessv1alpha1.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`{
								"apiVersion": "serverless.kyma-project.io/v1alpha1",
								"kind": "Function",
								"metadata": {
									"labels": {
										"serverless.kyma-project.io/function-resources-preset": "S"
									},
									"name": "testfunc",
									"namespace": "default"
								},
								"spec": {
									"source": "def main(event, context):\n  return \"hello world\"\n"
								}
							}`),
						},
					},
				},
			},
			want: want{
				// 6 patch operations added
				// add /spec/resources
				// add /spec/buildResources
				// add /spec/minReplicas
				// add /spec/maxReplicas
				// add /status
				// add /metadata/creationTimestamp
				operationsCount: 6,
			},
		},
		{
			name: "Bad request",
			fields: fields{
				configv1alpha2: &serverlessv1alpha2.DefaultingConfig{},
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: "Function", Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`bad request`),
						},
					},
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Fail on invalid kind",
			fields: fields{

				client:  fake.NewClientBuilder().Build(),
				decoder: decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: "Function", Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`{
								"apiVersion": "serverless.kyma-project.io/v1alpha2",
								"kind": "NotFunction",
								"metadata": {
									"labels": {
										"serverless.kyma-project.io/function-resources-preset": "S"
									},
									"name": "testfunc",
									"namespace": "default"
								},
								"spec": {
									"runtime": "python39"
								}
							}`),
						},
					},
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &DefaultingWebHook{
				configAlphaV2: tt.fields.configv1alpha2,
				configAlphaV1: tt.fields.configv1alpha1,
				client:        tt.fields.client,
				decoder:       tt.fields.decoder,
				log:           zap.NewNop().Sugar(),
			}
			got := w.Handle(tt.args.ctx, tt.args.req)

			if tt.want.operationsCount != 0 {
				require.True(t, got.Allowed)
				assert.Equal(t, tt.want.operationsCount, len(got.Patches), fmt.Sprintf("%+v", got.Patches))
			}
			if tt.want.statusCode != 0 {
				require.False(t, got.Allowed)
				require.Equal(t, tt.want.statusCode, got.Result.Code)
			}
		})
	}
}
