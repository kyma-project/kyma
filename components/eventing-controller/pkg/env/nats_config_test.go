package env

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetNatsConfig(t *testing.T) {
	type args struct {
		maxReconnects int
		reconnectWait time.Duration
		envs          map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    NatsConfig
		wantErr bool
	}{
		{name: "Empty env triggers error",
			wantErr: true,
		},
		{name: "Required values only gives valid config",
			args: args{
				envs: map[string]string{
					"EVENT_TYPE_PREFIX": "etp",
					"JS_STREAM_NAME":    "jsn",
				},
				maxReconnects: 1,
				reconnectWait: 1 * time.Second,
			},
			want: NatsConfig{
				URL:                     "nats.nats.svc.cluster.local",
				MaxReconnects:           1,
				ReconnectWait:           1 * time.Second,
				EventTypePrefix:         "etp",
				MaxIdleConns:            50,
				MaxConnsPerHost:         50,
				MaxIdleConnsPerHost:     50,
				IdleConnTimeout:         10 * time.Second,
				JSStreamName:            "jsn",
				JSStreamStorageType:     "memory",
				JSStreamReplicas:        1,
				JSStreamRetentionPolicy: "interest",
				JSStreamMaxMessages:     -1,
				JSStreamMaxBytes:        -1,
				JSConsumerDeliverPolicy: "new",
				EnableNewCRDVersion:     false,
			},
			wantErr: false,
		},
		{name: "Envs are mapped correctly",
			args: args{
				envs: map[string]string{
					"EVENT_TYPE_PREFIX":          "etp",
					"JS_STREAM_NAME":             "jsn",
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
				},
				maxReconnects: 1,
				reconnectWait: 1 * time.Second,
			},
			want: NatsConfig{
				URL:                     "natsurl",
				MaxReconnects:           1,
				ReconnectWait:           1 * time.Second,
				EventTypePrefix:         "etp",
				MaxIdleConns:            1,
				MaxConnsPerHost:         2,
				MaxIdleConnsPerHost:     3,
				IdleConnTimeout:         1 * time.Second,
				JSStreamName:            "jsn",
				JSStreamStorageType:     "jsst",
				JSStreamReplicas:        4,
				JSStreamRetentionPolicy: "jsrp",
				JSStreamMaxMessages:     5,
				JSStreamMaxBytes:        6,
				JSConsumerDeliverPolicy: "jcdp",
				EnableNewCRDVersion:     true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store the current environ and restore it after this test. A wrongly set up environment would break this test otherwise.
			env := os.Environ()
			t.Cleanup(func() {
				for _, e := range env {
					s := strings.Split(e, "=")
					os.Setenv(s[0], s[1])
				}
			})

			// Clean the environment to make this test reliable.
			os.Clearenv()
			for k, v := range tt.args.envs {
				t.Setenv(k, v)
			}

			got, err := GetNatsConfig(tt.args.maxReconnects, tt.args.reconnectWait)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNatsConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNatsConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
