package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_getResources(t *testing.T) {
	type args struct {
		res map[corev1.ResourceName]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[corev1.ResourceName]resource.Quantity
		wantErr bool
	}{
		{
			name:    "Parses Gi and m",
			wantErr: false,
			want: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse("5Gi"),
				corev1.ResourceMemory: resource.MustParse("5000m"),
			},
			args: args{res: map[corev1.ResourceName]string{
				corev1.ResourceCPU:    "5Gi",
				corev1.ResourceMemory: "5000m",
			}},
		},
		{
			name: "Parses Mi and whole number for cpu",
			want: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse("5Mi"),
				corev1.ResourceMemory: resource.MustParse("5"),
			},
			wantErr: false,
			args: args{
				res: map[corev1.ResourceName]string{
					corev1.ResourceCPU:    "5Mi",
					corev1.ResourceMemory: "5",
				},
			},
		},
		{
			name:    "errors on unconventional data",
			want:    nil,
			wantErr: true,
			args: args{
				res: map[corev1.ResourceName]string{
					corev1.ResourceCPU:    "5Blabla",
					corev1.ResourceMemory: "5000mili",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getResources(tt.args.res)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getResources() got = %v, want %v", got, tt.want)
			}
		})
	}
}
