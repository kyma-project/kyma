package webhook

import (
	"context"
	"net/http"
	"testing"

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

func TestValidatingWebHook_Handle(t *testing.T) {
	type fields struct {
		configV1Alpha2 serverlessv1alpha2.ValidationConfig
		client         ctrlclient.Client
		decoder        *admission.Decoder
	}
	type args struct {
		ctx context.Context
		req admission.Request
	}

	scheme := runtime.NewScheme()
	_ = serverlessv1alpha2.AddToScheme(scheme)
	decoder, err := admission.NewDecoder(scheme)
	require.NoError(t, err)

	tests := []struct {
		name         string
		fields       fields
		args         args
		responseCode int32
	}{
		{
			name: "Accept valid git function",
			fields: fields{
				configV1Alpha2: fixValidationConfig(),
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
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
    "name": "testfuncgit",
    "namespace": "default"
  },
  "spec": {
    "resourceConfiguration": {
      "build": {
        "resources": {
          "limits": {
            "cpu": "700m",
            "memory": "700Mi"
          },
          "requests": {
            "cpu": "200m",
            "memory": "200Mi"
          }
        }
      },
      "function": {
        "resources": {
          "limits": {
            "cpu": "400m",
            "memory": "512Mi"
          },
          "requests": {
            "cpu": "200m",
            "memory": "256Mi"
          }
        }
      }
    },
	"scaleConfig": {
		"maxReplicas": 1,
    	"minReplicas": 1
	},
    "runtime": "python39",
    "source": {
      "gitRepository": {
        "url": "test-url",
	"baseDir": "/py-handler",
	"reference": "test-ref"
      }
    }
  }
}`),
						},
					},
				},
			},
			responseCode: http.StatusOK,
		},
		{
			name: "Accept valid v1alpha2 function",
			fields: fields{
				configV1Alpha2: fixValidationConfig(),
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: serverlessv1alpha2.FunctionKind, Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(Marshall(t, ValidV1Alpha2Function())),
						},
					},
				},
			},
			responseCode: http.StatusOK,
		},
		{
			name: "Deny invalid function",
			fields: fields{
				configV1Alpha2: fixValidationConfig(),
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: "Function", Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`{"apiVersion": "serverless.kyma-project.io/v1alpha2",
								"kind": "Function",
								"metadata": {
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
			responseCode: http.StatusForbidden,
		},
		{
			name: "Bad request",
			fields: fields{
				configV1Alpha2: fixValidationConfig(),
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
			},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{Kind: "Function", Version: serverlessv1alpha2.FunctionVersion},
						Object: runtime.RawExtension{
							Raw: []byte(`{"bad request"`),
						},
					},
				},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name: "Deny on invalid kind",
			fields: fields{
				configV1Alpha2: fixValidationConfig(),
				client:         fake.NewClientBuilder().Build(),
				decoder:        decoder,
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
			responseCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ValidatingWebHook{
				configv1alpha2: &tt.fields.configV1Alpha2,
				client:         tt.fields.client,
				decoder:        tt.fields.decoder,
				log:            zap.NewNop().Sugar(),
			}
			got := w.Handle(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.responseCode, got.Result.Code)
		})
	}
}

func fixValidationConfig() serverlessv1alpha2.ValidationConfig {
	return serverlessv1alpha2.ValidationConfig{
		Function: serverlessv1alpha2.MinFunctionValues{
			Replicas: serverlessv1alpha2.MinFunctionReplicasValues{
				MinValue: int32(1),
			},
			Resources: serverlessv1alpha2.MinFunctionResourcesValues{
				MinRequestCPU:    "10m",
				MinRequestMemory: "16Mi",
			},
		},
		BuildJob: serverlessv1alpha2.MinBuildJobValues{Resources: serverlessv1alpha2.MinBuildJobResourcesValues{
			MinRequestCPU:    "200m",
			MinRequestMemory: "200Mi",
		}},
	}
}
