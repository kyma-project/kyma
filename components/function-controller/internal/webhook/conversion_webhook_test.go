package webhook

import (
	"testing"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestConvertingWebhook_convertFunction(t *testing.T) {
	// type fields struct {
	// 	scheme  *runtime.Scheme
	// 	client  ctrlclient.Client
	// 	decoder *conversion.Decoder
	// 	log     logr.Logger
	// }
	// type args struct {
	// 	src runtime.Object
	// 	dst runtime.Object
	// }

	client := fake.NewClientBuilder().Build()
	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	w := NewConvertingWebhook(client, scheme)
	tests := []struct {
		name string
		// fields  fields
		src runtime.Object
		wantDst runtime.Object
		// args    args
		wantErr bool
	}{{
		name: "v1alpha1 to v1alpha2 should work",
		src: ,
		dst: ,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := w.convertFunction(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("ConvertingWebhook.convertFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
