package beb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	testing2 "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	// mock server endpoints.
	eventsEndpoint        = "/events"
	eventsHTTP400Endpoint = "/events400"

	// connection settings.
	maxIdleConns        = 100
	maxIdleConnsPerHost = 200
)

func TestNewHttpMessageSender(t *testing.T) {
	t.Parallel()

	client := oauth.NewClient(context.Background(), &env.BEBConfig{})
	defer client.CloseIdleConnections()

	msgSender := NewSender(eventsEndpoint, client)
	if msgSender.Target != eventsEndpoint {
		t.Errorf("Message sender target is misconfigured want: %s but got: %s", eventsEndpoint, msgSender.Target)
	}
	if msgSender.Client != client {
		t.Errorf("Message sender client is misconfigured want: %#v but got: %#v", client, msgSender.Client)
	}
}

func TestNewRequestWithTarget(t *testing.T) {
	t.Parallel()

	client := oauth.NewClient(context.Background(), &env.BEBConfig{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost})
	defer client.CloseIdleConnections()

	msgSender := NewSender(eventsEndpoint, client)

	type ctxKey struct{}
	const ctxValue = "testValue"
	ctx := context.WithValue(context.Background(), ctxKey{}, ctxValue)
	req, err := msgSender.NewRequestWithTarget(ctx, eventsEndpoint)
	if err != nil {
		t.Errorf("Failed to create a CloudEvent HTTP request with error: %v", err)
	}
	if req == nil {
		t.Error("Failed to create a CloudEvent HTTP request want new request but got nil")
		return
	}
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request has invalid method want: %s but got: %s", http.MethodPost, req.Method)
	}
	if req.URL.Path != eventsEndpoint {
		t.Errorf("HTTP request has invalid target want: %s but got: %s", eventsEndpoint, req.URL.Path)
	}
	if len(req.Header) > 0 {
		t.Error("HTTP request should be created with empty headers")
	}
	if req.Close != false {
		t.Errorf("HTTP request close is invalid want: %v but got: %v", false, req.Close)
	}
	if req.Body != nil {
		t.Error("HTTP request should be created with empty body")
	}
	if req.Context() != ctx {
		t.Errorf("HTTP request context does not match original context want: %#v, but got %#v", ctx, req.Context())
	}
	if got := req.Context().Value(ctxKey{}); got != ctxValue {
		t.Errorf("HTTP request context key:value do not match mant: %v:%v but got %v:%v", ctxKey{}, ctxValue, ctxKey{}, got)
	}
}

func TestSender_Send_Error(t *testing.T) {
	type fields struct {
		Target string
	}
	type args struct {
		// timeout is one easy way to trigger an error on sending
		timeout time.Duration
		builder *testing2.CloudEventBuilder
	}
	var tests = []struct {
		name    string
		fields  fields
		args    args
		want    sender.PublishResult
		wantErr bool
	}{
		{
			name: "valid event",
			fields: fields{
				Target: "https://127.1.1.1:12345/idontexist",
			},
			args: args{
				timeout: 1 * time.Millisecond,
				builder: testing2.NewCloudEventBuilder(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hOk := &HandlerStub{ResponseStatus: 204}
			hFail := &HandlerStub{ResponseStatus: 400}
			mux := http.NewServeMux()
			mux.HandleFunc(eventsEndpoint, hOk.ServeHTTP)
			mux.HandleFunc(eventsHTTP400Endpoint, hFail.ServeHTTP)
			server := httptest.NewServer(mux)
			s := &Sender{
				Client: server.Client(),
				Target: tt.fields.Target,
			}
			ctx, cancel := context.WithTimeout(context.Background(), tt.args.timeout)
			defer cancel()
			got, err := s.Send(ctx, tt.args.builder.Build(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Send() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestSender_Send(t *testing.T) {
	type fields struct {
		Target string
	}
	type args struct {
		ctx     context.Context
		builder *testing2.CloudEventBuilder
	}
	var tests = []struct {
		name    string
		fields  fields
		args    args
		want    sender.PublishResult
		wantErr bool
	}{
		{
			name: "valid event, backend 400",
			fields: fields{
				Target: eventsHTTP400Endpoint,
			},
			args: args{
				ctx:     context.Background(),
				builder: testing2.NewCloudEventBuilder(),
			},
			want: HTTPPublishResult{
				Status: 400,
				Body:   []byte{},
			},
			wantErr: false,
		},
		{
			name: "valid event",
			fields: fields{
				Target: eventsEndpoint,
			},
			args: args{
				ctx:     context.Background(),
				builder: testing2.NewCloudEventBuilder(),
			},
			want: HTTPPublishResult{
				Status: 204,
				Body:   []byte{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hOk := &HandlerStub{ResponseStatus: 204}
			hFail := &HandlerStub{ResponseStatus: 400}
			mux := http.NewServeMux()
			mux.HandleFunc(eventsEndpoint, hOk.ServeHTTP)
			mux.HandleFunc(eventsHTTP400Endpoint, hFail.ServeHTTP)
			server := httptest.NewServer(mux)
			target, err := url.JoinPath(server.URL, tt.fields.Target)
			assert.NoError(t, err)
			s := &Sender{
				Client: server.Client(),
				Target: target,
			}
			got, err := s.Send(tt.args.ctx, tt.args.builder.Build(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Send() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type HandlerStub struct {
	Request        http.Request
	ResponseStatus int
}

func (h *HandlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.Request = *request
	writer.WriteHeader(h.ResponseStatus)
}
