// Code generated by mockery v2.30.16. DO NOT EDIT.

package mocks

import (
	nats "github.com/nats-io/nats.go"
	mock "github.com/stretchr/testify/mock"
)

// JetStreamContext is an autogenerated mock type for the JetStreamContext type
type JetStreamContext struct {
	mock.Mock
}

// AccountInfo provides a mock function with given fields: opts
func (_m *JetStreamContext) AccountInfo(opts ...nats.JSOpt) (*nats.AccountInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.AccountInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(...nats.JSOpt) (*nats.AccountInfo, error)); ok {
		return rf(opts...)
	}
	if rf, ok := ret.Get(0).(func(...nats.JSOpt) *nats.AccountInfo); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.AccountInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(...nats.JSOpt) error); ok {
		r1 = rf(opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddConsumer provides a mock function with given fields: stream, cfg, opts
func (_m *JetStreamContext) AddConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream, cfg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.ConsumerInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) (*nats.ConsumerInfo, error)); ok {
		return rf(stream, cfg, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) *nats.ConsumerInfo); ok {
		r0 = rf(stream, cfg, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.ConsumerInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) error); ok {
		r1 = rf(stream, cfg, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddStream provides a mock function with given fields: cfg, opts
func (_m *JetStreamContext) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, cfg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.StreamInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error)); ok {
		return rf(cfg, opts...)
	}
	if rf, ok := ret.Get(0).(func(*nats.StreamConfig, ...nats.JSOpt) *nats.StreamInfo); ok {
		r0 = rf(cfg, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.StreamInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.StreamConfig, ...nats.JSOpt) error); ok {
		r1 = rf(cfg, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChanQueueSubscribe provides a mock function with given fields: subj, queue, ch, opts
func (_m *JetStreamContext) ChanQueueSubscribe(subj string, queue string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, queue, ch)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, chan *nats.Msg, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, queue, ch, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, chan *nats.Msg, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, queue, ch, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, chan *nats.Msg, ...nats.SubOpt) error); ok {
		r1 = rf(subj, queue, ch, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChanSubscribe provides a mock function with given fields: subj, ch, opts
func (_m *JetStreamContext) ChanSubscribe(subj string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, ch)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, chan *nats.Msg, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, ch, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, chan *nats.Msg, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, ch, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, chan *nats.Msg, ...nats.SubOpt) error); ok {
		r1 = rf(subj, ch, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsumerInfo provides a mock function with given fields: stream, name, opts
func (_m *JetStreamContext) ConsumerInfo(stream string, name string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.ConsumerInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, ...nats.JSOpt) (*nats.ConsumerInfo, error)); ok {
		return rf(stream, name, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, ...nats.JSOpt) *nats.ConsumerInfo); ok {
		r0 = rf(stream, name, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.ConsumerInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, ...nats.JSOpt) error); ok {
		r1 = rf(stream, name, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsumerNames provides a mock function with given fields: stream, opts
func (_m *JetStreamContext) ConsumerNames(stream string, opts ...nats.JSOpt) <-chan string {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan string
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) <-chan string); ok {
		r0 = rf(stream, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan string)
		}
	}

	return r0
}

// Consumers provides a mock function with given fields: stream, opts
func (_m *JetStreamContext) Consumers(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan *nats.ConsumerInfo
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) <-chan *nats.ConsumerInfo); ok {
		r0 = rf(stream, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *nats.ConsumerInfo)
		}
	}

	return r0
}

// ConsumersInfo provides a mock function with given fields: stream, opts
func (_m *JetStreamContext) ConsumersInfo(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan *nats.ConsumerInfo
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) <-chan *nats.ConsumerInfo); ok {
		r0 = rf(stream, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *nats.ConsumerInfo)
		}
	}

	return r0
}

// CreateKeyValue provides a mock function with given fields: cfg
func (_m *JetStreamContext) CreateKeyValue(cfg *nats.KeyValueConfig) (nats.KeyValue, error) {
	ret := _m.Called(cfg)

	var r0 nats.KeyValue
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.KeyValueConfig) (nats.KeyValue, error)); ok {
		return rf(cfg)
	}
	if rf, ok := ret.Get(0).(func(*nats.KeyValueConfig) nats.KeyValue); ok {
		r0 = rf(cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.KeyValue)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.KeyValueConfig) error); ok {
		r1 = rf(cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateObjectStore provides a mock function with given fields: cfg
func (_m *JetStreamContext) CreateObjectStore(cfg *nats.ObjectStoreConfig) (nats.ObjectStore, error) {
	ret := _m.Called(cfg)

	var r0 nats.ObjectStore
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.ObjectStoreConfig) (nats.ObjectStore, error)); ok {
		return rf(cfg)
	}
	if rf, ok := ret.Get(0).(func(*nats.ObjectStoreConfig) nats.ObjectStore); ok {
		r0 = rf(cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.ObjectStore)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.ObjectStoreConfig) error); ok {
		r1 = rf(cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteConsumer provides a mock function with given fields: stream, consumer, opts
func (_m *JetStreamContext) DeleteConsumer(stream string, consumer string, opts ...nats.JSOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream, consumer)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, ...nats.JSOpt) error); ok {
		r0 = rf(stream, consumer, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteKeyValue provides a mock function with given fields: bucket
func (_m *JetStreamContext) DeleteKeyValue(bucket string) error {
	ret := _m.Called(bucket)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(bucket)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMsg provides a mock function with given fields: name, seq, opts
func (_m *JetStreamContext) DeleteMsg(name string, seq uint64, opts ...nats.JSOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, seq)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64, ...nats.JSOpt) error); ok {
		r0 = rf(name, seq, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteObjectStore provides a mock function with given fields: bucket
func (_m *JetStreamContext) DeleteObjectStore(bucket string) error {
	ret := _m.Called(bucket)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(bucket)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteStream provides a mock function with given fields: name, opts
func (_m *JetStreamContext) DeleteStream(name string, opts ...nats.JSOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) error); ok {
		r0 = rf(name, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLastMsg provides a mock function with given fields: name, subject, opts
func (_m *JetStreamContext) GetLastMsg(name string, subject string, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, subject)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.RawStreamMsg
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, ...nats.JSOpt) (*nats.RawStreamMsg, error)); ok {
		return rf(name, subject, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, ...nats.JSOpt) *nats.RawStreamMsg); ok {
		r0 = rf(name, subject, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.RawStreamMsg)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, ...nats.JSOpt) error); ok {
		r1 = rf(name, subject, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMsg provides a mock function with given fields: name, seq, opts
func (_m *JetStreamContext) GetMsg(name string, seq uint64, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, seq)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.RawStreamMsg
	var r1 error
	if rf, ok := ret.Get(0).(func(string, uint64, ...nats.JSOpt) (*nats.RawStreamMsg, error)); ok {
		return rf(name, seq, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, uint64, ...nats.JSOpt) *nats.RawStreamMsg); ok {
		r0 = rf(name, seq, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.RawStreamMsg)
		}
	}

	if rf, ok := ret.Get(1).(func(string, uint64, ...nats.JSOpt) error); ok {
		r1 = rf(name, seq, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KeyValue provides a mock function with given fields: bucket
func (_m *JetStreamContext) KeyValue(bucket string) (nats.KeyValue, error) {
	ret := _m.Called(bucket)

	var r0 nats.KeyValue
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (nats.KeyValue, error)); ok {
		return rf(bucket)
	}
	if rf, ok := ret.Get(0).(func(string) nats.KeyValue); ok {
		r0 = rf(bucket)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.KeyValue)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(bucket)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KeyValueStoreNames provides a mock function with given fields:
func (_m *JetStreamContext) KeyValueStoreNames() <-chan string {
	ret := _m.Called()

	var r0 <-chan string
	if rf, ok := ret.Get(0).(func() <-chan string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan string)
		}
	}

	return r0
}

// KeyValueStores provides a mock function with given fields:
func (_m *JetStreamContext) KeyValueStores() <-chan nats.KeyValueStatus {
	ret := _m.Called()

	var r0 <-chan nats.KeyValueStatus
	if rf, ok := ret.Get(0).(func() <-chan nats.KeyValueStatus); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan nats.KeyValueStatus)
		}
	}

	return r0
}

// ObjectStore provides a mock function with given fields: bucket
func (_m *JetStreamContext) ObjectStore(bucket string) (nats.ObjectStore, error) {
	ret := _m.Called(bucket)

	var r0 nats.ObjectStore
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (nats.ObjectStore, error)); ok {
		return rf(bucket)
	}
	if rf, ok := ret.Get(0).(func(string) nats.ObjectStore); ok {
		r0 = rf(bucket)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.ObjectStore)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(bucket)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ObjectStoreNames provides a mock function with given fields: opts
func (_m *JetStreamContext) ObjectStoreNames(opts ...nats.ObjectOpt) <-chan string {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan string
	if rf, ok := ret.Get(0).(func(...nats.ObjectOpt) <-chan string); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan string)
		}
	}

	return r0
}

// ObjectStores provides a mock function with given fields: opts
func (_m *JetStreamContext) ObjectStores(opts ...nats.ObjectOpt) <-chan nats.ObjectStoreStatus {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan nats.ObjectStoreStatus
	if rf, ok := ret.Get(0).(func(...nats.ObjectOpt) <-chan nats.ObjectStoreStatus); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan nats.ObjectStoreStatus)
		}
	}

	return r0
}

// Publish provides a mock function with given fields: subj, data, opts
func (_m *JetStreamContext) Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, data)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.PubAck
	var r1 error
	if rf, ok := ret.Get(0).(func(string, []byte, ...nats.PubOpt) (*nats.PubAck, error)); ok {
		return rf(subj, data, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, []byte, ...nats.PubOpt) *nats.PubAck); ok {
		r0 = rf(subj, data, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.PubAck)
		}
	}

	if rf, ok := ret.Get(1).(func(string, []byte, ...nats.PubOpt) error); ok {
		r1 = rf(subj, data, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PublishAsync provides a mock function with given fields: subj, data, opts
func (_m *JetStreamContext) PublishAsync(subj string, data []byte, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, data)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 nats.PubAckFuture
	var r1 error
	if rf, ok := ret.Get(0).(func(string, []byte, ...nats.PubOpt) (nats.PubAckFuture, error)); ok {
		return rf(subj, data, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, []byte, ...nats.PubOpt) nats.PubAckFuture); ok {
		r0 = rf(subj, data, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.PubAckFuture)
		}
	}

	if rf, ok := ret.Get(1).(func(string, []byte, ...nats.PubOpt) error); ok {
		r1 = rf(subj, data, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PublishAsyncComplete provides a mock function with given fields:
func (_m *JetStreamContext) PublishAsyncComplete() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// PublishAsyncPending provides a mock function with given fields:
func (_m *JetStreamContext) PublishAsyncPending() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// PublishMsg provides a mock function with given fields: m, opts
func (_m *JetStreamContext) PublishMsg(m *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, m)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.PubAck
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.Msg, ...nats.PubOpt) (*nats.PubAck, error)); ok {
		return rf(m, opts...)
	}
	if rf, ok := ret.Get(0).(func(*nats.Msg, ...nats.PubOpt) *nats.PubAck); ok {
		r0 = rf(m, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.PubAck)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.Msg, ...nats.PubOpt) error); ok {
		r1 = rf(m, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PublishMsgAsync provides a mock function with given fields: m, opts
func (_m *JetStreamContext) PublishMsgAsync(m *nats.Msg, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, m)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 nats.PubAckFuture
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.Msg, ...nats.PubOpt) (nats.PubAckFuture, error)); ok {
		return rf(m, opts...)
	}
	if rf, ok := ret.Get(0).(func(*nats.Msg, ...nats.PubOpt) nats.PubAckFuture); ok {
		r0 = rf(m, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.PubAckFuture)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.Msg, ...nats.PubOpt) error); ok {
		r1 = rf(m, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PullSubscribe provides a mock function with given fields: subj, durable, opts
func (_m *JetStreamContext) PullSubscribe(subj string, durable string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, durable)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, durable, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, durable, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, ...nats.SubOpt) error); ok {
		r1 = rf(subj, durable, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PurgeStream provides a mock function with given fields: name, opts
func (_m *JetStreamContext) PurgeStream(name string, opts ...nats.JSOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) error); ok {
		r0 = rf(name, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// QueueSubscribe provides a mock function with given fields: subj, queue, cb, opts
func (_m *JetStreamContext) QueueSubscribe(subj string, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, queue, cb)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, nats.MsgHandler, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, queue, cb, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, nats.MsgHandler, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, queue, cb, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, nats.MsgHandler, ...nats.SubOpt) error); ok {
		r1 = rf(subj, queue, cb, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueueSubscribeSync provides a mock function with given fields: subj, queue, opts
func (_m *JetStreamContext) QueueSubscribeSync(subj string, queue string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, queue)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, queue, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, queue, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, ...nats.SubOpt) error); ok {
		r1 = rf(subj, queue, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SecureDeleteMsg provides a mock function with given fields: name, seq, opts
func (_m *JetStreamContext) SecureDeleteMsg(name string, seq uint64, opts ...nats.JSOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, seq)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64, ...nats.JSOpt) error); ok {
		r0 = rf(name, seq, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StreamInfo provides a mock function with given fields: stream, opts
func (_m *JetStreamContext) StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.StreamInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) (*nats.StreamInfo, error)); ok {
		return rf(stream, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) *nats.StreamInfo); ok {
		r0 = rf(stream, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.StreamInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...nats.JSOpt) error); ok {
		r1 = rf(stream, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StreamNameBySubject provides a mock function with given fields: _a0, _a1
func (_m *JetStreamContext) StreamNameBySubject(_a0 string, _a1 ...nats.JSOpt) (string, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) (string, error)); ok {
		return rf(_a0, _a1...)
	}
	if rf, ok := ret.Get(0).(func(string, ...nats.JSOpt) string); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, ...nats.JSOpt) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StreamNames provides a mock function with given fields: opts
func (_m *JetStreamContext) StreamNames(opts ...nats.JSOpt) <-chan string {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan string
	if rf, ok := ret.Get(0).(func(...nats.JSOpt) <-chan string); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan string)
		}
	}

	return r0
}

// Streams provides a mock function with given fields: opts
func (_m *JetStreamContext) Streams(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan *nats.StreamInfo
	if rf, ok := ret.Get(0).(func(...nats.JSOpt) <-chan *nats.StreamInfo); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *nats.StreamInfo)
		}
	}

	return r0
}

// StreamsInfo provides a mock function with given fields: opts
func (_m *JetStreamContext) StreamsInfo(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan *nats.StreamInfo
	if rf, ok := ret.Get(0).(func(...nats.JSOpt) <-chan *nats.StreamInfo); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *nats.StreamInfo)
		}
	}

	return r0
}

// Subscribe provides a mock function with given fields: subj, cb, opts
func (_m *JetStreamContext) Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj, cb)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, nats.MsgHandler, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, cb, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, nats.MsgHandler, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, cb, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, nats.MsgHandler, ...nats.SubOpt) error); ok {
		r1 = rf(subj, cb, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeSync provides a mock function with given fields: subj, opts
func (_m *JetStreamContext) SubscribeSync(subj string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, subj)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...nats.SubOpt) (*nats.Subscription, error)); ok {
		return rf(subj, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, ...nats.SubOpt) *nats.Subscription); ok {
		r0 = rf(subj, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...nats.SubOpt) error); ok {
		r1 = rf(subj, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateConsumer provides a mock function with given fields: stream, cfg, opts
func (_m *JetStreamContext) UpdateConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, stream, cfg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.ConsumerInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) (*nats.ConsumerInfo, error)); ok {
		return rf(stream, cfg, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) *nats.ConsumerInfo); ok {
		r0 = rf(stream, cfg, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.ConsumerInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *nats.ConsumerConfig, ...nats.JSOpt) error); ok {
		r1 = rf(stream, cfg, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStream provides a mock function with given fields: cfg, opts
func (_m *JetStreamContext) UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, cfg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *nats.StreamInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error)); ok {
		return rf(cfg, opts...)
	}
	if rf, ok := ret.Get(0).(func(*nats.StreamConfig, ...nats.JSOpt) *nats.StreamInfo); ok {
		r0 = rf(cfg, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.StreamInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(*nats.StreamConfig, ...nats.JSOpt) error); ok {
		r1 = rf(cfg, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewJetStreamContext creates a new instance of JetStreamContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewJetStreamContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *JetStreamContext {
	mock := &JetStreamContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
