package docker

import (
	"encoding/base64"
	"errors"
	"io"
	"testing"

	"github.com/onsi/gomega"
)

func Test_registryCfgCredentials_MarshalJSON(t *testing.T) {
	type args struct {
		username       []byte
		password       []byte
		serverAddress  []byte
		provideEncoder provideEncoder
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				username:       []byte("Z1d6dDZveUU1dHhIQ3dlYWtPV2M="),
				password:       []byte("eHVwNE5GRkpIRzZvTWVZd094b09xWmpBeE92QzBOWTBOZ0lzUlVYTg=="),
				serverAddress:  []byte("registry.35.246.34.254.xip.io"),
				provideEncoder: base64.NewEncoder,
			},
			want: []byte(`{
				"auths": {
				  "registry.35.246.34.254.xip.io": {
					"auth": "WjFkNmREWnZlVVUxZEhoSVEzZGxZV3RQVjJNPTplSFZ3TkU1R1JrcElSelp2VFdWWmQwOTRiMDl4V21wQmVFOTJRekJPV1RCT1owbHpVbFZZVGc9PQ=="
				  }
				}
			}`),
			wantErr: false,
		},
		{
			name: "encoder write error",
			args: args{
				username:      []byte("Z1d6dDZveUU1dHhIQ3dlYWtPV2M="),
				password:      []byte("eHVwNE5GRkpIRzZvTWVZd094b09xWmpBeE92QzBOWTBOZ0lzUlVYTg=="),
				serverAddress: []byte("registry.35.246.34.254.xip.io"),
				provideEncoder: func(enc *base64.Encoding, w io.Writer) io.WriteCloser {
					return &failingWriterCloser{
						failType: failWrite,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "encoder close error",
			args: args{
				username:      []byte("Z1d6dDZveUU1dHhIQ3dlYWtPV2M="),
				password:      []byte("eHVwNE5GRkpIRzZvTWVZd094b09xWmpBeE92QzBOWTBOZ0lzUlVYTg=="),
				serverAddress: []byte("registry.35.246.34.254.xip.io"),
				provideEncoder: func(enc *base64.Encoding, w io.Writer) io.WriteCloser {
					return &failingWriterCloser{
						failType: failClose,
					}
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registryCfgCredentials{
				username:       tt.args.username,
				password:       tt.args.password,
				serverAddress:  tt.args.serverAddress,
				provideEncoder: tt.args.provideEncoder,
			}
			got, err := r.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("registryCfgCredentials.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tt.want) == 0 {
				return
			}
			gomega.NewWithT(t).Expect(got).To(gomega.MatchJSON(tt.want))
		})
	}
}

func TestNewRegistryCfgMarshaller(t *testing.T) {
	type args struct {
		data map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "err missing username",
			args: args{
				data: map[string][]byte{
					keyPassword:        []byte("-"),
					keyRegistryAddress: []byte("-"),
				},
			},
			wantErr: true,
		},
		{
			name: "err missing password",
			args: args{
				data: map[string][]byte{
					keyUsername:        []byte("-"),
					keyRegistryAddress: []byte("-"),
				},
			},
			wantErr: true,
		},
		{
			name: "err missing url",
			args: args{
				data: map[string][]byte{
					keyUsername: []byte("-"),
					keyPassword: []byte("-"),
				},
			},
			wantErr: true,
		},
		{
			name: "OK",
			args: args{
				data: map[string][]byte{
					keyUsername:        []byte("-"),
					keyPassword:        []byte("-"),
					keyRegistryAddress: []byte("-"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRegistryCfgMarshaler(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistryCfgMarshaler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

var (
	errWrite = errors.New("test write error")
	errClose = errors.New("test close error")
)

type failType int

const (
	failWrite failType = iota
	failClose
)

type failingWriterCloser struct {
	failType
}

func (f *failingWriterCloser) Write(p []byte) (n int, err error) {
	if f.failType != failWrite {
		return len(p), nil
	}
	return -1, errWrite
}

func (f *failingWriterCloser) Close() error {
	if f.failType != failClose {
		return nil
	}
	return errClose
}
