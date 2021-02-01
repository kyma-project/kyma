package kubernetes

import (
	"context"
	"testing"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func Test_registryWatcher_Handle(t *testing.T) {
	decoder, err := admission.NewDecoder(runtime.NewScheme())
	if err != nil {
		t.Error(err)
	}
	type fields struct {
		Decoder *admission.Decoder
	}
	type args struct {
		ctx context.Context
		req admission.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int32
	}{
		{
			name:   "decode error - bad request",
			fields: fields{decoder},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1beta1.AdmissionRequest{
						Object: runtime.RawExtension{
							Raw: []byte("err"),
						},
					},
				},
			},
			wantCode: 400,
		},
		{
			name:   "missing username - internal error",
			fields: fields{decoder},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1beta1.AdmissionRequest{
						Object: runtime.RawExtension{
							Raw: []byte(`{
								"apiVersion": "v1",
								"kind": "Secret",
								"metadata": {
									"creationTimestamp": "2021-01-31T18:21:24Z",
									"managedFields": [
										{
											"apiVersion": "v1",
											"fieldsType": "FieldsV1",
											"fieldsV1": {
												"f:type": {}
											},
											"manager": "kubectl",
											"operation": "Update",
											"time": "2021-01-31T18:21:24Z"
										}
									],
									"name": "empty-secret",
									"namespace": "default",
									"resourceVersion": "606",
									"uid": "06549245-7a23-407a-9734-f00580adbb40"
								},
								"type": "Opaque"
							}`),
						},
					},
				},
			},
			wantCode: 500,
		},
		{
			name:   "Ok",
			fields: fields{decoder},
			args: args{
				ctx: context.Background(),
				req: admission.Request{
					AdmissionRequest: v1beta1.AdmissionRequest{
						Object: runtime.RawExtension{
							Raw: []byte(`{
								"apiVersion": "v1",
								"data": {
									"password": "dGVzdA==",
									"registry": "dGVzdA==",
									"username": "dGVzdA=="
								},
								"kind": "Secret",
								"metadata": {
									"creationTimestamp": "2021-01-31T18:41:36Z",
									"managedFields": [
										{
											"apiVersion": "v1",
											"fieldsType": "FieldsV1",
											"fieldsV1": {
												"f:data": {
													".": {},
													"f:password": {},
													"f:registry": {},
													"f:username": {}
												},
												"f:type": {}
											},
											"manager": "kubectl",
											"operation": "Update",
											"time": "2021-01-31T18:41:36Z"
										}
									],
									"name": "registry-credentials",
									"namespace": "default",
									"resourceVersion": "1455",
									"uid": "6f6018ab-0ea1-469f-afb6-1c052e7b26a9"
								},
								"type": "Opaque"
							}
							`),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registryWatcher{
				Decoder: tt.fields.Decoder,
			}
			got := r.Handle(tt.args.ctx, tt.args.req)
			if (tt.wantCode != 0) && got.Result.Code != tt.wantCode {
				t.Errorf("response result code = %v, want %v", got.Result.Code, tt.wantCode)
			}
		})
	}
}
