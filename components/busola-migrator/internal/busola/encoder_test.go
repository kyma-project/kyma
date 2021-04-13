package busola

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_encodeInitString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{
			name:    "simple json",
			arg:     `{"data":"anything"}`,
			want:    "XQAAgAD__________wBAqQiGE9mRGS0n_ewjaG-DLt6IU__ksYAA",
			wantErr: false,
		}, {
			name:    "not json",
			arg:     `not json`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := encodeInitString(tt.arg)
		if tt.wantErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, tt.want, got)

		_, err = base64.RawURLEncoding.DecodeString(got)
		assert.Nil(t, err)
	}
}
