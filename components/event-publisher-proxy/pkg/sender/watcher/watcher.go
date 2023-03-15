package watcher

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/watcher"
	"k8s.io/client-go/transport"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// buffer the event types to be sent by workers.
	buffer      = 1000000
	eventSource = "/default/sap.kyma/tunas-develop" // TODO make this configurable
)

// compile-time check for interfaces implementation.
var (
	_ watcher.UpdateNotifiable = &Sender{}
)

type Sender struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      client.Client
	config      *watcher.Config
	events      []events.Event
	process     chan bool
	queue       chan events.Event
	cleanup     sync.WaitGroup
	running     bool
	undelivered int32
	ack         int32
	nack        int32
}

func New(config *config.Config) *Sender {
	return &Sender{config: config}
}

func (s *Sender) NotifyUpdate(cm *corev1.ConfigMap) {
	//config.Map(cm, s.config)
	s.stop()
	s.init()
	s.start()
}

func (s *Sender) init() {
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = cloudevents.NewClientOrDie(t.Clone())
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.events = events.Generate(s.config)
	s.process = make(chan bool, s.config.EpsLimit)
	s.queue = make(chan events.Event, buffer)
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
}

func (s *Sender) start() {
	s.running = true
	s.queueEventsAsync()
	s.sendEventsAsync()
}

func (s *Sender) stop() {
	// recover from closing already closed channels
	defer func() { recover() }()

	s.running = false
	s.cancel()
	close(s.process)
	close(s.queue)
	s.cleanup.Wait()
}

func (s *Sender) queueEventsAsync() {
	for _, e := range s.events {
		go s.queueEvent(e)
	}
}

func (s *Sender) sendEventsAsync() {
	for i := 0; i < s.config.Workers; i++ {
		go s.sendEvents()
	}
}

func (s *Sender) queueEvent(e events.Event) {
	defer func() {
		recover()
		s.cleanup.Done()
	}()

	s.cleanup.Add(1)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// queue event immediately
	for ; s.running; <-ticker.C {
		for i := 0; i < e.Eps; i++ {
			if !s.running {
				return
			}

			s.queue <- e
		}
	}
}

func (s *Sender) sendEvents() {
	defer func() {
		recover()
		s.cleanup.Done()
	}()

	s.cleanup.Add(1)

	for s.running {
		select {
		case e := <-s.queue:
			{
				if !s.running {
					return
				}

				s.process <- true
				go s.sendEvent(e)
			}
		}
	}
}

func (s *Sender) sendEvent(e events.Event) {
	defer func() {
		if !s.running {
			return
		}

		<-s.process
	}()

	ce := cev2.NewEvent()
	ce.SetType(e.EventType)
	ce.SetSource(eventSource)
	data := map[string]interface{}{"type": e.EventType, "time": time.Now()}
	if err := ce.SetData(cev2.ApplicationJSON, data); err != nil {
		log.Printf("Failed to set cloudevent data: %s", err)
		return
	}

	result := s.client.Send(cev2.ContextWithTarget(s.ctx, s.config.PublishEndpoint), ce)
	switch {
	case cev2.IsUndelivered(result):
		{
			atomic.AddInt32(&s.undelivered, 1)
		}
	case cev2.IsACK(result):
		{
			atomic.AddInt32(&s.ack, 1)
		}
	case cev2.IsNACK(result):
		{
			atomic.AddInt32(&s.nack, 1)
		}
	}
}
