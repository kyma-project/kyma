package migrator

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func Test_migrateTriggerDestination(t *testing.T) {
	type args struct {
		tr *eventingv1alpha1.Trigger
	}
	testNamespace := "test-ns"
	testSvcName := "test-name"

	testUrl, err := apis.ParseURL(fmt.Sprintf("http://%s.%s:8080/", testSvcName, testNamespace))

	if err != nil {
		t.Fatal("should parse")
	}

	tests := []struct {
		name    string
		args    args
		want    *duckv1.Destination
		wantErr bool
	}{
		{
			name:    "migrates typical trigger",
			wantErr: false,
			want: &duckv1.Destination{
				Ref: &corev1.ObjectReference{
					Kind:       "Service",
					Namespace:  testNamespace,
					Name:       testSvcName,
					APIVersion: "serving.knative.dev/v1",
				},
				URI: nil,
			},
			args: args{tr: &eventingv1alpha1.Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "random-name",
					Namespace: "random-namespace",
				},
				Spec: eventingv1alpha1.TriggerSpec{
					Subscriber: &duckv1.Destination{
						URI: testUrl,
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got, err := migrateTriggerDestination(tt.args.tr)

			g.Expect((err != nil) == tt.wantErr).To(gomega.BeTrue())
			g.Expect(got).To(gomega.BeEquivalentTo(tt.want))
		})
	}
}
