package knative

import (
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func Test_getNewestGeneration(t *testing.T) {
	type args struct {
		revisions []servingv1.Revision
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "should return highest generation from label",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "3"}}},
				},
			},
			want: 3,
		},
		{
			name: "should return error if even one revision lacks proper label",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"random-label": "3"}}},
				},
			},
			want:    -1,
			wantErr: true,
		},
		{
			name: "should return error generation is not a number",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "NaN"}}},
				},
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)

			got, err := getNewestGeneration(tt.args.revisions)
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_getOldRevisionSelector(t *testing.T) {
	// serverless.kyma-project.io/uuid=uid,serving.knative.dev/configurationGeneration!=3
	type args struct {
		instance  *serverlessv1alpha1.Function
		revisions []servingv1.Revision
	}
	tests := []struct {
		name    string
		args    args
		want    labels.Selector
		wantErr bool
	}{
		{
			name: "properly parses revisions and instances",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "3"}}},
				},
			},
			wantErr: false,
			want: func() labels.Selector {
				lbl, _ := labels.Parse("serverless.kyma-project.io/uuid=uid,serving.knative.dev/configurationGeneration!=3")
				return lbl
			}(),
		},
		{
			name: "fails with incorrect labels from revisions",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{serverless.CfgGenerationLabel: "ups"}}},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)

			r := &ServiceReconciler{}
			got, err := r.getOldRevisionSelector(tt.args.instance, tt.args.revisions)

			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			if got != nil {
				g.Expect(got).To(gomega.Equal(tt.want))
			} else {
				g.Expect(got).To(gomega.BeNil())
			}

		})
	}
}
