package subscribed

import (
	"reflect"
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func TestFilterEventTypeVersions(t *testing.T) {
	testCases := []struct {
		name            string
		appName         string
		eventTypePrefix string
		bebNs           string
		filters         *eventingv1alpha1.BebFilters
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
			filters:         NewBEBFilters(WithOneBEBFilter),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return multiple events in a slice when multiple filters are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromSameSource),
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
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromSameSource),
			expectedEvents:  []Event{},
		}, {
			name:            "should return 1 event(out of multiple) which matches one source(bebNamespace)",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "foo-match",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromDiffSource),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return 2 out 3 events in a slice when filters with different eventTypePrefix are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromDiffEventTypePrefix),
			expectedEvents: []Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotEvents := FilterEventTypeVersions(tc.eventTypePrefix, tc.bebNs, tc.appName, tc.filters)
			if !reflect.DeepEqual(tc.expectedEvents, gotEvents) {
				t.Errorf("Received incorrect events, Wanted: %v, Got: %v", tc.expectedEvents, gotEvents)
			}

		})
	}
}

func TestConvertEventsMapToSlice(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			gotUniqEvents := AddUniqueEventsToResult(tc.eventsSubSet, tc.givenUniqEventsAlready)
			if !reflect.DeepEqual(tc.wantedUniqEvents, gotUniqEvents) {
				t.Errorf("incorrect unique events, wanted: %v, got: %v", tc.wantedUniqEvents, gotUniqEvents)
			}
		})
	}
}

type BEBFilterOption func(filter *eventingv1alpha1.BebFilters)

func NewBEBFilters(opts ...BEBFilterOption) *eventingv1alpha1.BebFilters {
	newFilters := &eventingv1alpha1.BebFilters{}
	for _, opt := range opts {
		opt(newFilters)
	}

	return newFilters
}

func WithOneBEBFilter(bebFilters *eventingv1alpha1.BebFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventingv1alpha1.BebFilter{
		NewBEBFilter(evSource, evType),
	}

}

func WithMultipleBEBFiltersFromSameSource(bebFilters *eventingv1alpha1.BebFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventingv1alpha1.BebFilter{
		NewBEBFilter(evSource, evType),
		NewBEBFilter(evSource, evType),
		NewBEBFilter(evSource, evType),
	}
}

func WithMultipleBEBFiltersFromDiffSource(bebFilters *eventingv1alpha1.BebFilters) {
	evSource1 := "foo-match"
	evSource2 := "/default/foo.different/kt1"
	evSource3 := "/default/foo.different2/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventingv1alpha1.BebFilter{
		NewBEBFilter(evSource1, evType),
		NewBEBFilter(evSource2, evType),
		NewBEBFilter(evSource3, evType),
	}
}

func WithMultipleBEBFiltersFromDiffEventTypePrefix(bebFilters *eventingv1alpha1.BebFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType1 := "foo.prefix.custom.foovarkes.order.created.v1"
	evType2 := "foo.prefixdifferent.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventingv1alpha1.BebFilter{
		NewBEBFilter(evSource, evType1),
		NewBEBFilter(evSource, evType2),
		NewBEBFilter(evSource, evType1),
	}
}

func NewBEBFilter(evSource, evType string) *eventingv1alpha1.BebFilter {
	return &eventingv1alpha1.BebFilter{
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
