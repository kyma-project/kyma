//go:build unit
// +build unit

package jetstreamv2

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/stretchr/testify/assert"
)

func TestUnitValidate_For_Errors(t *testing.T) {
	tests := []struct {
		name        string
		givenConfig backendnats.Config
		wantError   error
	}{
		{
			name:        "ErrorEmptyStream",
			givenConfig: backendnats.Config{JSStreamName: ""},
			wantError:   ErrEmptyStreamName,
		},
		{
			name:        "ErrorStreamToLong",
			givenConfig: backendnats.Config{JSStreamName: fixtureStreamNameTooLong()},
			wantError:   ErrStreamNameTooLong,
		},
		{
			name: "ErrorStorageType",
			givenConfig: backendnats.Config{
				JSStreamName:        "not-empty",
				JSStreamStorageType: "invalid-storage-type",
			},
			wantError: ErrInvalidStorageType.WithArg("invalid-storage-type"),
		},
		{
			name: "ErrorRetentionPolicy",
			givenConfig: backendnats.Config{
				JSStreamName:            "not-empty",
				JSStreamStorageType:     StorageTypeMemory,
				JSStreamRetentionPolicy: "invalid-retention-policy",
			},
			wantError: ErrInvalidRetentionPolicy.WithArg("invalid-retention-policy"),
		},
		{
			name: "ErrorDiscardPolicy",
			givenConfig: backendnats.Config{
				JSStreamName:            "not-empty",
				JSStreamStorageType:     StorageTypeMemory,
				JSStreamRetentionPolicy: RetentionPolicyInterest,
				JSStreamDiscardPolicy:   "invalid-discard-policy",
			},
			wantError: ErrInvalidDiscardPolicy.WithArg("invalid-discard-policy"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.givenConfig)
			assert.ErrorIs(t, err, tc.wantError)
		})
	}
}
func Test_GetNATSConfig(t *testing.T) {
	type args struct {
		maxReconnects int
		reconnectWait time.Duration
		envs          map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    backendnats.Config
		wantErr bool
	}{
		{name: "Empty env triggers error",
			wantErr: true,
		},
		{name: "Required values only gives valid config",
			args: args{
				envs: map[string]string{
					"NATS_URL":                 "natsurl",
					"EVENT_TYPE_PREFIX":        "etp",
					"JS_STREAM_NAME":           "jsn",
					"JS_STREAM_SUBJECT_PREFIX": "kma",
				},
				maxReconnects: 1,
				reconnectWait: 1 * time.Second,
			},
			want: backendnats.Config{
				URL:                     "natsurl",
				MaxReconnects:           1,
				ReconnectWait:           1 * time.Second,
				EventTypePrefix:         "etp",
				MaxIdleConns:            50,
				MaxConnsPerHost:         50,
				MaxIdleConnsPerHost:     50,
				IdleConnTimeout:         10 * time.Second,
				JSStreamName:            "jsn",
				JSSubjectPrefix:         "kma",
				JSStreamStorageType:     "memory",
				JSStreamReplicas:        1,
				JSStreamRetentionPolicy: "interest",
				JSStreamMaxMessages:     -1,
				JSStreamMaxBytes:        "-1",
				JSConsumerDeliverPolicy: "new",
				JSStreamDiscardPolicy:   "new",
				EnableNewCRDVersion:     false,
			},
			wantErr: false,
		},
		{name: "Envs are mapped correctly",
			args: args{
				envs: map[string]string{
					"EVENT_TYPE_PREFIX":          "etp",
					"JS_STREAM_NAME":             "jsn",
					"JS_STREAM_SUBJECT_PREFIX":   "testjsn",
					"NATS_URL":                   "natsurl",
					"MAX_IDLE_CONNS":             "1",
					"MAX_CONNS_PER_HOST":         "2",
					"MAX_IDLE_CONNS_PER_HOST":    "3",
					"IDLE_CONN_TIMEOUT":          "1s",
					"JS_STREAM_STORAGE_TYPE":     "jsst",
					"JS_STREAM_REPLICAS":         "4",
					"JS_STREAM_RETENTION_POLICY": "jsrp",
					"JS_STREAM_MAX_MSGS":         "5",
					"JS_STREAM_MAX_BYTES":        "6",
					"JS_CONSUMER_DELIVER_POLICY": "jcdp",
					"ENABLE_NEW_CRD_VERSION":     "true",
					"JS_STREAM_DISCARD_POLICY":   "jsdp",
				},
				maxReconnects: 1,
				reconnectWait: 1 * time.Second,
			},
			want: backendnats.Config{
				URL:                     "natsurl",
				MaxReconnects:           1,
				ReconnectWait:           1 * time.Second,
				EventTypePrefix:         "etp",
				MaxIdleConns:            1,
				MaxConnsPerHost:         2,
				MaxIdleConnsPerHost:     3,
				IdleConnTimeout:         1 * time.Second,
				JSStreamName:            "jsn",
				JSSubjectPrefix:         "testjsn",
				JSStreamStorageType:     "jsst",
				JSStreamReplicas:        4,
				JSStreamRetentionPolicy: "jsrp",
				JSStreamMaxMessages:     5,
				JSStreamMaxBytes:        "6",
				JSConsumerDeliverPolicy: "jcdp",
				EnableNewCRDVersion:     true,
				JSStreamDiscardPolicy:   "jsdp",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store the current environ and restore it after this test.
			// A wrongly set up environment would break this test otherwise.
			env := os.Environ()
			t.Cleanup(func() {
				for _, e := range env {
					s := strings.Split(e, "=")
					if err := os.Setenv(s[0], s[1]); err != nil {
						t.Log(err)
					}
				}
			})

			// Clean the environment to make this test reliable.
			os.Clearenv()
			for k, v := range tt.args.envs {
				t.Setenv(k, v)
			}

			got, err := backendnats.GetNATSConfig(tt.args.maxReconnects, tt.args.reconnectWait)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNATSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNATSConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func fixtureStreamNameTooLong() string {
	b := strings.Builder{}
	for i := 0; i < (jsMaxStreamNameLength + 1); i++ {
		b.WriteString("a")
	}
	streamName := b.String()
	return streamName
}
