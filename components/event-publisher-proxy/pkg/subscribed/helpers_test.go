package subscribed

import (
	"reflect"
	"testing"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func TestFilterEventTypeVersions(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		appName        string
		subscription   *eventingv1alpha2.Subscription
		expectedEvents []Event
	}{
		{
			name:           "should return no events when there is no subscription",
			appName:        "fooapp",
			subscription:   &eventingv1alpha2.Subscription{},
			expectedEvents: make([]Event, 0),
		}, {
			name:    "should return a slice of events when eventTypes are provided",
			appName: "foovarkes",
			subscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Source: "foovarkes",
					Types: []string{
						"order.created.v1",
						"order.created.v2",
					},
				},
			},
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v2"),
			},
		}, {
			name:    "should return no event if app name is different than subscription source",
			appName: "foovarkes",
			subscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Source: "diff-source",
					Types: []string{
						"order.created.v1",
						"order.created.v2",
					},
				},
			},
			expectedEvents: []Event{},
		}, {
			name:    "should return event types if event type consists of eventType and appName for typeMaching exact",
			appName: "foovarkes",
			subscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Source:       "/default/sap.kyma/tunas-develop",
					TypeMatching: eventingv1alpha2.TypeMatchingExact,
					Types: []string{
						"sap.kyma.custom.foovarkes.order.created.v1",
						"sap.kyma.custom.foovarkes.order.created.v2",
					},
				},
			},
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v2"),
			},
		}, {
			name:    "should return no event if app name is not part of external event types",
			appName: "foovarkes",
			subscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Source:       "/default/sap.kyma/tunas-develop",
					TypeMatching: eventingv1alpha2.TypeMatchingExact,
					Types: []string{
						"sap.kyma.custom.difffoovarkes.order.created.v1",
						"sap.kyma.custom.difffoovarkes.order.created.v2",
					},
				},
			},
			expectedEvents: []Event{},
		}, {
			name:    "should return event type only with 'sap.kyma.custom' prefix and appname",
			appName: "foovarkes",
			subscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Source:       "/default/sap.kyma/tunas-develop",
					TypeMatching: eventingv1alpha2.TypeMatchingExact,
					Types: []string{
						"foo.prefix.custom.foovarkes.order.created.v1",
						"sap.kyma.custom.foovarkes.order.created.v2",
						"sap.kyma.custom.diffvarkes.order.created.v2",
					},
				},
			},
			expectedEvents: []Event{
				NewEvent("order.created", "v2"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEvents := FilterEventTypeVersions("sap.kyma.custom", tc.appName, tc.subscription)
			if !reflect.DeepEqual(tc.expectedEvents, gotEvents) {
				t.Errorf("Received incorrect events, Wanted: %v, Got: %v", tc.expectedEvents, gotEvents)
			}
		})
	}
}

func TestBuildEventType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		appName string
		want    Event
	}{
		{
			name:    "should return no events when there is no subscription",
			appName: "order.created.v1",
			want: Event{
				Name:    "order.created",
				Version: "v1",
			},
		}, {
			name:    "should return a slice of events when eventTypes are provided",
			appName: "product.order.created.v1",
			want: Event{
				Name:    "product.order.created",
				Version: "v1",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			event := buildEvent(tc.appName)
			if !reflect.DeepEqual(tc.want, event) {
				t.Errorf("Received incorrect events, Wanted: %v, Got: %v", tc.want, event)
			}
		})
	}
}

func TestFilterEventTypeVersionsV1alpha1(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name            string
		appName         string
		eventTypePrefix string
		bebNs           string
		filters         *eventingv1alpha1.BEBFilters
		expectedEvents  []Event
	}{
		{
			name:            "should return no events when nil filters are provided",
			appName:         "fooapp",
			eventTypePrefix: "foo.prefix",
			bebNs:           "foo.bebns",
			filters:         nil,
			expectedEvents:  make([]Event, 0),
		}, {
			name:            "should return a slice of events when filters are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewEventMeshFilters(WithOneEventMeshFilter),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return multiple events in a slice when multiple filters are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewEventMeshFilters(WithMultipleEventMeshFiltersFromSameSource),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return no events when filters sources(bebNamespace) don't match",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "foo-dont-match",
			filters:         NewEventMeshFilters(WithMultipleEventMeshFiltersFromSameSource),
			expectedEvents:  []Event{},
		}, {
			name:            "should return 2 events(out of multiple) which matches two sources (bebNamespace and empty)",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "foo-match",
			filters:         NewEventMeshFilters(WithMultipleEventMeshFiltersFromDiffSource),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return 2 out 3 events in a slice when filters with different eventTypePrefix are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewEventMeshFilters(WithMultipleEventMeshFiltersFromDiffEventTypePrefix),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEvents := FilterEventTypeVersionsV1alpha1(tc.eventTypePrefix, tc.bebNs, tc.appName, tc.filters)
			if !reflect.DeepEqual(tc.expectedEvents, gotEvents) {
				t.Errorf("Received incorrect events, Wanted: %v, Got: %v", tc.expectedEvents, gotEvents)
			}
		})
	}
}

func TestConvertEventsMapToSlice(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		inputMap     map[Event]bool
		wantedEvents []Event
	}{
		{
			name: "should return events from the map in a slice",
			inputMap: map[Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v2"): true,
			},
			wantedEvents: []Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v2"),
			},
		}, {
			name:         "should return no events for an empty map of events",
			inputMap:     map[Event]bool{},
			wantedEvents: []Event{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEvents := ConvertEventsMapToSlice(tc.inputMap)
			for _, event := range gotEvents {
				found := false
				for _, wantEvent := range tc.wantedEvents {
					if event == wantEvent {
						found = true
						continue
					}
				}
				if !found {
					t.Errorf("incorrect slice of events, wanted: %v, got: %v", tc.wantedEvents, gotEvents)
				}
			}
		})
	}
}

func TestAddUniqueEventsToResult(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                   string
		eventsSubSet           []Event
		givenUniqEventsAlready map[Event]bool
		wantedUniqEvents       map[Event]bool
	}{
		{
			name: "should return unique events along with the existing ones",
			eventsSubSet: []Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v1"),
			},
			givenUniqEventsAlready: map[Event]bool{
				NewEvent("bar-already-existing", "v1"): true,
			},
			wantedUniqEvents: map[Event]bool{
				NewEvent("foo", "v1"):                  true,
				NewEvent("bar", "v1"):                  true,
				NewEvent("bar-already-existing", "v1"): true,
			},
		}, {
			name: "should return unique new events from the subset provided only",
			eventsSubSet: []Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v1"),
			},
			givenUniqEventsAlready: nil,
			wantedUniqEvents: map[Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
		}, {
			name:         "should return existing unique events when an empty subset provided",
			eventsSubSet: []Event{},
			givenUniqEventsAlready: map[Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
			wantedUniqEvents: map[Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
		}, {
			name:                   "should return no unique events when an empty subset provided",
			eventsSubSet:           []Event{},
			givenUniqEventsAlready: map[Event]bool{},
			wantedUniqEvents:       map[Event]bool{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotUniqEvents := AddUniqueEventsToResult(tc.eventsSubSet, tc.givenUniqEventsAlready)
			if !reflect.DeepEqual(tc.wantedUniqEvents, gotUniqEvents) {
				t.Errorf("incorrect unique events, wanted: %v, got: %v", tc.wantedUniqEvents, gotUniqEvents)
			}
		})
	}
}

type EventMeshFilterOption func(filter *eventingv1alpha1.BEBFilters)

func NewEventMeshFilters(opts ...EventMeshFilterOption) *eventingv1alpha1.BEBFilters {
	newFilters := &eventingv1alpha1.BEBFilters{}
	for _, opt := range opts {
		opt(newFilters)
	}

	return newFilters
}

func WithOneEventMeshFilter(filters *eventingv1alpha1.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	filters.Filters = []*eventingv1alpha1.EventMeshFilter{
		NewEventMeshFilter(evSource, evType),
	}
}

func WithMultipleEventMeshFiltersFromSameSource(filters *eventingv1alpha1.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	filters.Filters = []*eventingv1alpha1.EventMeshFilter{
		NewEventMeshFilter(evSource, evType),
		NewEventMeshFilter(evSource, evType),
		NewEventMeshFilter(evSource, evType),
	}
}

func WithMultipleEventMeshFiltersFromDiffSource(filters *eventingv1alpha1.BEBFilters) {
	evSource1 := "foo-match"
	evSource2 := "/default/foo.different/kt1"
	evSource3 := "/default/foo.different2/kt1"
	evSource4 := ""
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	filters.Filters = []*eventingv1alpha1.EventMeshFilter{
		NewEventMeshFilter(evSource1, evType),
		NewEventMeshFilter(evSource2, evType),
		NewEventMeshFilter(evSource3, evType),
		NewEventMeshFilter(evSource4, evType),
	}
}

func WithMultipleEventMeshFiltersFromDiffEventTypePrefix(filters *eventingv1alpha1.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType1 := "foo.prefix.custom.foovarkes.order.created.v1"
	evType2 := "foo.prefixdifferent.custom.foovarkes.order.created.v1"
	filters.Filters = []*eventingv1alpha1.EventMeshFilter{
		NewEventMeshFilter(evSource, evType1),
		NewEventMeshFilter(evSource, evType2),
		NewEventMeshFilter(evSource, evType1),
	}
}

func NewEventMeshFilter(evSource, evType string) *eventingv1alpha1.EventMeshFilter {
	return &eventingv1alpha1.EventMeshFilter{
		EventSource: &eventingv1alpha1.Filter{
			Property: "source",
			Value:    evSource,
		},
		EventType: &eventingv1alpha1.Filter{
			Property: "type",
			Value:    evType,
		},
	}
}

func NewEvent(name, version string) Event {
	return Event{
		Name:    name,
		Version: version,
	}
}
