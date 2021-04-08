package ptr

import (
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	testStr := "test"

	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{name: "Success", args: args{val: testStr}, want: &testStr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
