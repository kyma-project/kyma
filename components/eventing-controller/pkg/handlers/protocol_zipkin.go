package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

type ProtocolZipkin struct {
	protocol cehttp.Protocol
}

func NewProtocolZipkin(proto *cehttp.Protocol) ProtocolZipkin {
	p := &ProtocolZipkin{
		protocol: *proto,
	}
	return *p
}

func (p *ProtocolZipkin) Request(ctx context.Context, m binding.Message, transformers ...binding.Transformer) (binding.Message, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil Context")
	} else if m == nil {
		return nil, fmt.Errorf("nil Message")
	}

	var err error
	defer func() { _ = m.Finish(err) }()

	req := p.makeRequest(ctx)

	if p.protocol.Client == nil || req == nil || req.URL == nil {
		return nil, fmt.Errorf("not initialized: %#v", p)
	}

	if err = cehttp.WriteRequest(ctx, m, req, transformers...); err != nil {
		return nil, err
	}

	if header := req.Header.Get("ce-xb3traceid"); header != "" {
		req.Header.Add("x-b3-traceid", header)
	}
	if header := req.Header.Get("ce-xb3spanid"); header != "" {
		req.Header.Add("x-b3-spanid", header)
	}
	if header := req.Header.Get("ce-xb3sampled"); header != "" {
		req.Header.Add("x-b3-sampled", header)
	}

	return p.do(ctx, req)
}

func (p *ProtocolZipkin) makeRequest(ctx context.Context) *http.Request {
	// TODO: support custom headers from context?
	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		// TODO: HeaderFrom(ctx),
	}

	if p.protocol.RequestTemplate != nil {
		req.Method = p.protocol.RequestTemplate.Method
		req.URL = p.protocol.RequestTemplate.URL
		req.Close = p.protocol.RequestTemplate.Close
		req.Host = p.protocol.RequestTemplate.Host
		copyHeadersEnsure(p.protocol.RequestTemplate.Header, &req.Header)
	}

	if p.protocol.Target != nil {
		req.URL = p.protocol.Target
	}

	// Override the default request with target from context.
	if target := cecontext.TargetFrom(ctx); target != nil {
		req.URL = target
	}
	return req.WithContext(ctx)
}

// Ensure to is a non-nil map before copying
func copyHeadersEnsure(from http.Header, to *http.Header) {
	if len(from) > 0 {
		if *to == nil {
			*to = http.Header{}
		}
		copyHeaders(from, *to)
	}
}

func copyHeaders(from, to http.Header) {
	if from == nil || to == nil {
		return
	}
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

func (p *ProtocolZipkin) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	return p.protocol.Send(ctx, m, transformers...)
}

func (p *ProtocolZipkin) Receive(ctx context.Context) (binding.Message, error) {
	return p.protocol.Receive(ctx)
}

func (p *ProtocolZipkin) Respond(ctx context.Context) (binding.Message, protocol.ResponseFn, error) {
	return p.protocol.Respond(ctx)
}
func (p *ProtocolZipkin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.protocol.ServeHTTP(rw, req)
}
