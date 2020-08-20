package sender

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
)

var noRetries = RetryConfig{
	RetryMax: 0,
	CheckRetry: func(ctx context.Context, resp *nethttp.Response, err error) (bool, error) {
		return false, nil
	},
	Backoff: func(attemptNum int, resp *nethttp.Response) time.Duration {
		return 0
	},
}

// ConnectionArgs allow to configure connection parameters to the underlying
// HTTP Client transport.
type ConnectionArgs struct {
	// MaxIdleConns refers to the max idle connections, as in net/http/transport.
	MaxIdleConns int
	// MaxIdleConnsPerHost refers to the max idle connections per host, as in net/http/transport.
	MaxIdleConnsPerHost int
}

func (ca *ConnectionArgs) ConfigureTransport(transport *nethttp.Transport) {
	if ca == nil {
		return
	}
	transport.MaxIdleConns = ca.MaxIdleConns
	transport.MaxIdleConnsPerHost = ca.MaxIdleConnsPerHost
}

type HttpMessageSender struct {
	Client *nethttp.Client
	Target string
}

func NewHttpMessageSender(connectionArgs *ConnectionArgs, target string, httpClient *nethttp.Client) (*HttpMessageSender, error) {
	// Add connection options to the default transport.
	var base = nethttp.DefaultTransport.(*nethttp.Transport).Clone()
	connectionArgs.ConfigureTransport(base)
	return &HttpMessageSender{Client: httpClient, Target: target}, nil
}

func (s *HttpMessageSender) NewCloudEventRequest(ctx context.Context) (*nethttp.Request, error) {
	return nethttp.NewRequestWithContext(ctx, "POST", s.Target, nil)
}

func (s *HttpMessageSender) NewCloudEventRequestWithTarget(ctx context.Context, target string) (*nethttp.Request, error) {
	return nethttp.NewRequestWithContext(ctx, "POST", target, nil)
}

func (s *HttpMessageSender) Send(req *nethttp.Request) (*nethttp.Response, error) {
	return s.Client.Do(req)
}

// CheckRetry specifies a policy for handling retries. It is called
// following each request with the response and error values returned by
// the http.Client. If CheckRetry returns false, the Client stops retrying
// and returns the response to the caller. If CheckRetry returns an error,
// that error value is returned in lieu of the error from the request. The
// Client will close any response body when retrying, but if the retry is
// aborted it is up to the CheckRetry callback to properly close any
// response body before returning.
type CheckRetry func(ctx context.Context, resp *nethttp.Response, err error) (bool, error)

// Backoff specifies a policy for how long to wait between retries.
// It is called after a failing request to determine the amount of time
// that should pass before trying again.
type Backoff func(attemptNum int, resp *nethttp.Response) time.Duration

type RetryConfig struct {

	// Maximum number of retries
	RetryMax int

	CheckRetry CheckRetry

	Backoff Backoff
}

func (s *HttpMessageSender) SendWithRetries(req *nethttp.Request, config *RetryConfig) (*nethttp.Response, error) {
	if config == nil {
		return s.Send(req)
	}

	retryableClient := retryablehttp.Client{
		HTTPClient:   s.Client,
		RetryWaitMin: defaultRetryWaitMin,
		RetryWaitMax: defaultRetryWaitMax,
		RetryMax:     config.RetryMax,
		CheckRetry:   retryablehttp.CheckRetry(config.CheckRetry),
		Backoff: func(_, _ time.Duration, attemptNum int, resp *nethttp.Response) time.Duration {
			return config.Backoff(attemptNum, resp)
		},
		ErrorHandler: func(resp *nethttp.Response, err error, numTries int) (*nethttp.Response, error) {
			return resp, err
		},
	}

	return retryableClient.Do(&retryablehttp.Request{Request: req})
}

func NoRetries() RetryConfig {
	return noRetries
}

//func RetryConfigFromDeliverySpec(spec interface{}) (RetryConfig, error) {
//
//	retryConfig := NoRetries()
//
//	if spec.Retry != nil {
//		retryConfig.RetryMax = int(*spec.Retry)
//	}
//
//	if spec.BackoffPolicy != nil && spec.BackoffDelay != nil {
//
//		delay, err := period.Parse(*spec.BackoffDelay)
//		if err != nil {
//			return retryConfig, fmt.Errorf("failed to parse Spec.BackoffDelay: %w", err)
//		}
//
//		delayDuration, _ := delay.Duration()
//		switch *spec.BackoffPolicy {
//		case duckv1.BackoffPolicyExponential:
//			retryConfig.Backoff = func(attemptNum int, resp *nethttp.Response) time.Duration {
//				return time.Duration(math.Pow(float64(delayDuration*2), float64(attemptNum)))
//			}
//		case duckv1.BackoffPolicyLinear:
//			retryConfig.Backoff = func(attemptNum int, resp *nethttp.Response) time.Duration {
//				return delayDuration
//			}
//		}
//	}
//
//	return retryConfig, nil
//}
