package jetstreamv2

import (
	"github.com/nats-io/nats.go"
)

type jetStreamContextStub struct {
	consumerInfoError error
	consumerInfo      *nats.ConsumerInfo

	addConsumerError error
	addConsumer      *nats.ConsumerInfo
}

func (j jetStreamContextStub) Streams(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) Consumers(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ObjectStoreNames(opts ...nats.ObjectOpt) <-chan string {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ObjectStores(opts ...nats.ObjectOpt) <-chan nats.ObjectStore {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PublishMsg(m *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsync(subj string, data []byte, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PublishMsgAsync(m *nats.Msg, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsyncPending() int {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsyncComplete() <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) SubscribeSync(subj string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ChanSubscribe(subj string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) QueueSubscribe(subj, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) QueueSubscribeSync(subj, queue string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) PullSubscribe(subj, durable string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	panic("really implement me")
}

func (j jetStreamContextStub) UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) DeleteStream(name string, opts ...nats.JSOpt) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	panic("really implement me")
}

func (j jetStreamContextStub) PurgeStream(name string, opts ...nats.JSOpt) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) StreamsInfo(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) StreamNames(opts ...nats.JSOpt) <-chan string {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) GetMsg(name string, seq uint64, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) GetLastMsg(name, subject string, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) DeleteMsg(name string, seq uint64, opts ...nats.JSOpt) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) SecureDeleteMsg(name string, seq uint64, opts ...nats.JSOpt) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) AddConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return j.addConsumer, j.addConsumerError
}

func (j jetStreamContextStub) UpdateConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) DeleteConsumer(stream, consumer string, opts ...nats.JSOpt) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ConsumerInfo(stream, name string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return j.consumerInfo, j.consumerInfoError
}

func (j jetStreamContextStub) ConsumersInfo(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ConsumerNames(stream string, opts ...nats.JSOpt) <-chan string {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) AccountInfo(opts ...nats.JSOpt) (*nats.AccountInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) KeyValue(bucket string) (nats.KeyValue, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) CreateKeyValue(cfg *nats.KeyValueConfig) (nats.KeyValue, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) DeleteKeyValue(bucket string) error {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) ObjectStore(bucket string) (nats.ObjectStore, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) CreateObjectStore(cfg *nats.ObjectStoreConfig) (nats.ObjectStore, error) {
	//TODO implement me
	panic("implement me")
}

func (j jetStreamContextStub) DeleteObjectStore(bucket string) error {
	//TODO implement me
	panic("implement me")
}
