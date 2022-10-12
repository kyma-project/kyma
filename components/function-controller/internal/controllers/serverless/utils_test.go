package serverless

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestUtils_mergeLabels(t *testing.T) {
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
			got := mergeLabels(tt.args.labelsCollection...)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
