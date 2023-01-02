package jetstreamv2

import (
	"github.com/nats-io/nats.go"
)

type jetStreamContextStub struct {
	consumerInfoError error
	consumerInfo      *nats.ConsumerInfo

	addConsumerError error
	addConsumer      *nats.ConsumerInfo

	subscribe      *nats.Subscription
	subscribeError error

	updateConsumer      *nats.ConsumerInfo
	updateConsumerError error

	updateStreamInfo      *nats.StreamInfo
	updateStreamInfoError error

	streamInfo      *nats.StreamInfo
	streamInfoError error

	consumers         []*nats.ConsumerInfo
	deleteConsumerErr error

	addStreamInfo  *nats.StreamInfo
	addStreamError error
}

func (j jetStreamContextStub) StreamNameBySubject(_ string, _ ...nats.JSOpt) (string, error) {
	panic("implement me")
}

func (j jetStreamContextStub) ObjectStores(_ ...nats.ObjectOpt) <-chan nats.ObjectStoreStatus {
	panic("implement me")
}

func (j jetStreamContextStub) KeyValueStoreNames() <-chan string {
	panic("implement me")
}

func (j jetStreamContextStub) KeyValueStores() <-chan nats.KeyValueStatus {
	panic("implement me")
}

func (j jetStreamContextStub) Streams(_ ...nats.JSOpt) <-chan *nats.StreamInfo {
	panic("implement me")
}

func (j jetStreamContextStub) Consumers(_ string, _ ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	ch := make(chan *nats.ConsumerInfo, len(j.consumers))
	defer close(ch)
	for _, con := range j.consumers {
		ch <- con
	}
	return ch
}

func (j jetStreamContextStub) ObjectStoreNames(_ ...nats.ObjectOpt) <-chan string {
	panic("implement me")
}

func (j jetStreamContextStub) Publish(_ string, _ []byte, _ ...nats.PubOpt) (*nats.PubAck, error) {
	panic("implement me")
}

func (j jetStreamContextStub) PublishMsg(_ *nats.Msg, _ ...nats.PubOpt) (*nats.PubAck, error) {
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsync(_ string, _ []byte, _ ...nats.PubOpt) (nats.PubAckFuture, error) {
	panic("implement me")
}

func (j jetStreamContextStub) PublishMsgAsync(_ *nats.Msg, _ ...nats.PubOpt) (nats.PubAckFuture, error) {
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsyncPending() int {
	panic("implement me")
}

func (j jetStreamContextStub) PublishAsyncComplete() <-chan struct{} {
	panic("implement me")
}

func (j jetStreamContextStub) Subscribe(_ string, _ nats.MsgHandler, _ ...nats.SubOpt) (*nats.Subscription, error) {
	return j.subscribe, j.subscribeError
}

func (j jetStreamContextStub) SubscribeSync(_ string, _ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) ChanSubscribe(_ string, _ chan *nats.Msg, _ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) ChanQueueSubscribe(_, _ string,
	_ chan *nats.Msg, _ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) QueueSubscribe(_, _ string, _ nats.MsgHandler,
	_ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) QueueSubscribeSync(_, _ string, _ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) PullSubscribe(_, _ string, _ ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (j jetStreamContextStub) AddStream(_ *nats.StreamConfig, _ ...nats.JSOpt) (*nats.StreamInfo, error) {
	return j.addStreamInfo, j.addStreamError
}

func (j jetStreamContextStub) UpdateStream(_ *nats.StreamConfig, _ ...nats.JSOpt) (*nats.StreamInfo, error) {
	return j.updateStreamInfo, j.updateStreamInfoError
}

func (j jetStreamContextStub) DeleteStream(_ string, _ ...nats.JSOpt) error {
	panic("implement me")
}

func (j jetStreamContextStub) StreamInfo(_ string, _ ...nats.JSOpt) (*nats.StreamInfo, error) {
	return j.streamInfo, j.streamInfoError
}

func (j jetStreamContextStub) PurgeStream(_ string, _ ...nats.JSOpt) error {
	panic("implement me")
}

func (j jetStreamContextStub) StreamsInfo(_ ...nats.JSOpt) <-chan *nats.StreamInfo {
	panic("implement me")
}

func (j jetStreamContextStub) StreamNames(_ ...nats.JSOpt) <-chan string {
	panic("implement me")
}

func (j jetStreamContextStub) GetMsg(_ string, _ uint64, _ ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	panic("implement me")
}

func (j jetStreamContextStub) GetLastMsg(_, _ string, _ ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	panic("implement me")
}

func (j jetStreamContextStub) DeleteMsg(_ string, _ uint64, _ ...nats.JSOpt) error {
	panic("implement me")
}

func (j jetStreamContextStub) SecureDeleteMsg(_ string, _ uint64, _ ...nats.JSOpt) error {
	panic("implement me")
}

func (j jetStreamContextStub) AddConsumer(_ string, _ *nats.ConsumerConfig,
	_ ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return j.addConsumer, j.addConsumerError
}

func (j jetStreamContextStub) UpdateConsumer(_ string, _ *nats.ConsumerConfig,
	_ ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return j.updateConsumer, j.updateConsumerError
}

func (j *jetStreamContextStub) DeleteConsumer(_, consumer string, _ ...nats.JSOpt) error {
	if j.deleteConsumerErr != nil {
		return j.deleteConsumerErr
	}
	for i := len(j.consumers) - 1; i >= 0; i-- {
		if j.consumers[i].Name == consumer {
			j.consumers = remove(j.consumers, i)
		}
	}
	return nil
}

func remove(slice []*nats.ConsumerInfo, i int) []*nats.ConsumerInfo {
	return append(slice[:i], slice[i+1:]...)
}

func (j jetStreamContextStub) ConsumerInfo(_, _ string, _ ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return j.consumerInfo, j.consumerInfoError
}

func (j jetStreamContextStub) ConsumersInfo(_ string, _ ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	panic("implement me")
}

func (j jetStreamContextStub) ConsumerNames(_ string, _ ...nats.JSOpt) <-chan string {
	panic("implement me")
}

func (j jetStreamContextStub) AccountInfo(_ ...nats.JSOpt) (*nats.AccountInfo, error) {
	panic("implement me")
}

func (j jetStreamContextStub) KeyValue(_ string) (nats.KeyValue, error) {
	panic("implement me")
}

func (j jetStreamContextStub) CreateKeyValue(_ *nats.KeyValueConfig) (nats.KeyValue, error) {
	panic("implement me")
}

func (j jetStreamContextStub) DeleteKeyValue(_ string) error {
	panic("implement me")
}

func (j jetStreamContextStub) ObjectStore(_ string) (nats.ObjectStore, error) {
	panic("implement me")
}

func (j jetStreamContextStub) CreateObjectStore(_ *nats.ObjectStoreConfig) (nats.ObjectStore, error) {
	panic("implement me")
}

func (j jetStreamContextStub) DeleteObjectStore(_ string) error {
	panic("implement me")
}

type subscriberStub struct {
	isValid bool

	unsubscribeError error
}

func (s subscriberStub) SubscriptionSubject() string {
	return ""
}

func (s subscriberStub) IsValid() bool {
	return s.isValid
}

func (s subscriberStub) ConsumerInfo() (*nats.ConsumerInfo, error) {
	panic("implement me")
}

func (s subscriberStub) Unsubscribe() error {
	return s.unsubscribeError
}

func (s subscriberStub) SetPendingLimits(_ int, _ int) error {
	panic("implement me")
}

func (s subscriberStub) PendingLimits() (int, int, error) {
	panic("implement me")
}
