package subscribed

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	kneventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	fakeeventingclient "knative.dev/eventing/pkg/client/injection/client/fake"

	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
)

func Test_stuff(t *testing.T) {
	t.Parallel()

	// test cases
	tests := []struct {
		name   string
		source string
		given  []Event // used to create triggers
		want   Events
	}{
		{
			name:   "application with no triggers",
			source: "application1",
			given:  []Event{},
			want:   Events{EventsInfo: []Event{}},
		},
		{
			name:   "application with many triggers",
			source: "application2",
			given: []Event{
				{Name: "event1", Version: "v1"},
				{Name: "event2", Version: "v2"},
				{Name: "event3", Version: "v3"},
				{Name: "event4", Version: "v4"},
				{Name: "event5", Version: "v5"},
				{Name: "event6", Version: "v6"},
				{Name: "event7", Version: "v7"},
				{Name: "event8", Version: "v8"},
			},
			want: Events{
				EventsInfo: []Event{
					{Name: "event1", Version: "v1"},
					{Name: "event2", Version: "v2"},
					{Name: "event3", Version: "v3"},
					{Name: "event4", Version: "v4"},
					{Name: "event5", Version: "v5"},
					{Name: "event6", Version: "v6"},
					{Name: "event7", Version: "v7"},
					{Name: "event8", Version: "v8"},
				},
			},
		},
		{
			name:   "application with few triggers",
			source: "application3",
			given: []Event{
				{Name: "event1", Version: "v1"},
				{Name: "event2", Version: "v2"},
				{Name: "event3", Version: "v3"},
			},
			want: Events{
				EventsInfo: []Event{
					{Name: "event1", Version: "v1"},
					{Name: "event2", Version: "v2"},
					{Name: "event3", Version: "v3"},
				},
			},
		},
	}

	// prepare an array of trigger runtime objects
	objects := make([]runtime.Object, 0)
	for _, test := range tests {
		for _, event := range test.given {
			// create trigger from the given test info
			trigger := newTrigger(test.source, event.Name, event.Version)
			objects = append(objects, trigger)
		}
	}

	// prepare a fake events client
	_, client := fakeeventingclient.With(context.Background(), objects...)
	eventsClient := NewEventsClient(client)

	// run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := retry.Do(func() error {
				// get the subscribed events for an application
				got, err := eventsClient.GetSubscribedEvents(test.source)
				if err != nil {
					return err
				}

				// fail if the returned result is not as expected
				if !containSameEvents(test.want.EventsInfo, got.EventsInfo) {
					return fmt.Errorf("returned events are not matching the expected result")
				}

				return nil
			},
				retry.Attempts(10),
				retry.Delay(2*time.Second),
				retry.DelayType(retry.FixedDelay),
				retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
			)

			if err != nil {
				t.Fatalf("test: %s failed: %v", test.name, err)
			}
		})
	}
}

func newTrigger(source, eventType, eventTypeVersion string) *kneventingv1alpha1.Trigger {
	return &kneventingv1alpha1.Trigger{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s", source, eventType, eventTypeVersion),
		},
		Spec: kneventingv1alpha1.TriggerSpec{
			Filter: &kneventingv1alpha1.TriggerFilter{
				Attributes: &kneventingv1alpha1.TriggerFilterAttributes{
					keySource:           source,
					keyEventType:        eventType,
					keyEventTypeVersion: eventTypeVersion,
				},
			},
		},
		Status: kneventingv1alpha1.TriggerStatus{},
	}
}

func containSameEvents(events ...[]Event) bool {
	// key mapper
	key := func(evt *Event) string { return fmt.Sprintf("%s-%s", evt.Name, evt.Version) }

	// map of key counts
	m := make(map[string]int)

	// increment key count per event
	for _, evts := range events {
		for _, evt := range evts {
			m[key(&evt)] += 1
		}
	}

	// check counts per key to be equal to the length of the given events
	for _, v := range m {
		if v < len(events) {
			return false
		}
	}

	return true
}
