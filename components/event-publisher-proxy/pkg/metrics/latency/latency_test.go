package latency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBucketsProvider(t *testing.T) {
	tests := []struct {
		name string
		want *BucketsProvider
	}{
		{
			name: "default latency buckets",
			want: &BucketsProvider{
				buckets: []float64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			got := NewBucketsProvider()

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBucketsProvider_Buckets(t *testing.T) {
	tests := []struct {
		name    string
		buckets []float64
		want    []float64
	}{
		{
			name:    "nil buckets",
			buckets: nil,
			want:    nil,
		},
		{
			name:    "empty buckets",
			buckets: []float64{},
			want:    []float64{},
		},
		{
			name:    "non-empty buckets",
			buckets: []float64{1, 2, 3},
			want:    []float64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p := BucketsProvider{buckets: tt.buckets}

			// when
			got := p.Buckets()

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}
