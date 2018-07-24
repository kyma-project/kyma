package actors

import (
	"log"
	"time"

	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/handlers"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/stanutil"
	trc "github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"k8s.io/apimachinery/pkg/types"
)

type subscriptionCRWithNATSStreamingSubscription struct {
	subscriptionCR            *v1alpha1.Subscription
	natsStreamingSubscription *stan.Subscription
}

const (
	actionChannelSize                       = 16
	retryConnectionAndSubscriptionsInterval = 30 * time.Second
)

type action func()

type SubscriptionsSupervisorInterface interface {
	PoisonPill()
	IsRunning() bool
	IsNATSConnected() bool
	StartSubscriptionReq(sub *v1alpha1.Subscription)
	StopSubscriptionReq(sub *v1alpha1.Subscription)
}

// SubscriptionsSupervisor manages the state of NATS Streaming subscriptions
type SubscriptionsSupervisor struct {
	running       bool
	actionChan    chan action
	opts          *opts.Options
	stanConn      *stan.Conn
	subscriptions map[types.UID]*subscriptionCRWithNATSStreamingSubscription
	mhf           *handlers.MessageHandlerFactory
}

// StartSubscriptionsSupervisor starts supervisor actor
func StartSubscriptionsSupervisor(opts *opts.Options, tracer *trc.Tracer) *SubscriptionsSupervisor {
	log.Print("starting subscriptions supervisor actor")
	ch := make(chan action, actionChannelSize)
	mhf := handlers.NewMessageHandlerFactory(opts, tracer)
	s := &SubscriptionsSupervisor{running: true, actionChan: ch, opts: opts, mhf: mhf}

	s.initialize()
	go s.actorLoop(ch)
	return s
}

func (s *SubscriptionsSupervisor) initialize() {
	log.Print("initializing subscriptions supervisor actor")
	s.subscriptions = make(map[types.UID]*subscriptionCRWithNATSStreamingSubscription)

	s.connectToNATSStreamingCluster()
}

func (s *SubscriptionsSupervisor) connectToNATSStreamingCluster() {
	var err error

	if s.stanConn, err = stanutil.Connect(s.opts.NatsStreamingClusterID, s.opts.ClientID, s.opts.NatsURL); err != nil {
		log.Fatalf("Can't connect to NATS Streaming: %v\n", err)
	}
}

func pingInterval(i time.Duration) nats.Option {
	return func(o *nats.Options) error {
		o.PingInterval = i
		return nil
	}
}

func (s *SubscriptionsSupervisor) actorLoop(ch <-chan action) {
	timer := time.NewTimer(retryConnectionAndSubscriptionsInterval)

	for s.running {
		select {
		case action := <-ch:
			action()
		case <-timer.C:
			s.retryNATSStreamingSubscriptions()
			timer = time.NewTimer(retryConnectionAndSubscriptionsInterval) // TODO improve, schedule new retry only if needed
		}
	}
}

// PoisonPill requests actor to stop
func (s *SubscriptionsSupervisor) PoisonPill() {
	s.actionChan <- func() {
		s.running = false
		close(s.actionChan)
		stanutil.Close(s.stanConn)
	}
}

// IsRunning returns the status of supervisor actor
func (s *SubscriptionsSupervisor) IsRunning() bool {
	resultChan := make(chan bool, 0)
	s.actionChan <- func() {
		resultChan <- s.running
	}
	return <-resultChan
}

// IsNATSConnected returns the status of NATS connection
func (s *SubscriptionsSupervisor) IsNATSConnected() bool {
	resultChan := make(chan bool, 0)
	s.actionChan <- func() {
		resultChan <- stanutil.IsConnected(s.stanConn)
	}
	return <-resultChan
}

// ReconnectToNATSStreaming ....
func (s *SubscriptionsSupervisor) ReconnectToNATSStreaming() {
	s.actionChan <- func() {
		log.Println("ReconnectToNATSStreaming() :: try to reconnet to NATS Streaming")
		stanutil.Close(s.stanConn)
		stanutil.Connect(s.opts.NatsStreamingClusterID, s.opts.ClientID, s.opts.NatsURL)
	}
}

// StartSubscriptionReq message models the request to start handling EventBus subscription
func (s *SubscriptionsSupervisor) StartSubscriptionReq(sub *v1alpha1.Subscription) {
	s.actionChan <- func() {
		log.Print("handling EventBus Subscription start request")
		if val, present := s.subscriptions[sub.UID]; !present || val == nil {
			s.subscriptions[sub.UID] = &subscriptionCRWithNATSStreamingSubscription{subscriptionCR: sub}
			s.retryNATSStreamingSubscriptions()
		} else { // present && val != nil
			if val.natsStreamingSubscription != nil {
				log.Print("subscription is already being handled")
			} else {
				log.Print("already aware of the subscription, handling will be retried later")
			}
		}
	}
}

// StopSubscriptionReq message models the request to stop handling EventBus subscription
func (s *SubscriptionsSupervisor) StopSubscriptionReq(sub *v1alpha1.Subscription) {
	s.actionChan <- func() {
		log.Printf("handling EventBus Subscription %s stop request", sub.Name)
		if val, present := s.subscriptions[sub.UID]; present {
			if val.natsStreamingSubscription != nil {
				s.stopSubscription(sub)
			}

			// cleanup references
			delete(s.subscriptions, sub.UID)
		} else {
			log.Printf("there was no subscription %s", sub.Name)
		}
	}
}

func (s *SubscriptionsSupervisor) retryNATSStreamingSubscriptions() {
	if s.stanConn != nil {
		for _, subscriptionCRWithNATSStreamingSubscription := range s.subscriptions {
			if subscriptionCRWithNATSStreamingSubscription.natsStreamingSubscription == nil {
				s.handleSubscription(subscriptionCRWithNATSStreamingSubscription.subscriptionCR)
			}
		}
	}
}

func (s *SubscriptionsSupervisor) stopSubscription(sub *v1alpha1.Subscription) {
	log.Printf("Stopping handling EventBus subscription %s", sub.Name)
	// Close NATS Streaming subscription for the EventBus subscription being stopped
	subscriptionCRWithNATSStreamingSubscription := s.subscriptions[sub.UID]
	if err := (*subscriptionCRWithNATSStreamingSubscription.natsStreamingSubscription).Unsubscribe(); err != nil {
		log.Printf("error unsubscribing NATS Streaming subscription: %v", err)
	}
}

func (s *SubscriptionsSupervisor) handleSubscription(sub *v1alpha1.Subscription) {
	log.Printf("attempting establishing NATS Streaming subscription for %s", sub.Name)

	msgHandler := s.mhf.NewMsgHandler(sub, s.opts)
	topic := common.FromSubscriptionSpec(sub.SubscriptionSpec).Encode()

	natsStreamingSub, err := (*s.stanConn).QueueSubscribe(
		topic, s.opts.QueueGroup, msgHandler,
		stan.DurableName(sub.Name),
		stan.SetManualAckMode(),
		stan.AckWait(s.opts.AckWait),
		stan.MaxInflight(sub.SubscriptionSpec.MaxInflight))

	if err != nil {
		log.Printf("failed to subscribe: %v", err)
	} else {
		s.subscriptions[sub.UID].natsStreamingSubscription = &natsStreamingSub
		log.Print("established NATS Streaming subscription")
	}
}
