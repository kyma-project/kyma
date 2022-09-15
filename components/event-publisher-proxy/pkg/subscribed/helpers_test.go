//go:build unit
// +build unit

package subscribed_test

import (
	"reflect"
	"testing"

	eventing "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func TestFilterEventTypeVersions(t *testing.T) {
	testCases := []struct {
		name            string
		appName         string
		eventTypePrefix string
		bebNs           string
		filters         *eventing.BEBFilters
		expectedEvents  []subscribed.Event
	}{
		{
			name:            "should return no events when nil filters are provided",
			appName:         "fooapp",
			eventTypePrefix: "foo.prefix",
			bebNs:           "foo.bebns",
			filters:         nil,
			expectedEvents:  make([]subscribed.Event, 0),
		}, {
			name:            "should return a slice of events when filters are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewBEBFilters(WithOneBEBFilter),
			expectedEvents: []subscribed.Event{
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return multiple events in a slice when multiple filters are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromSameSource),
			expectedEvents: []subscribed.Event{
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
			expectedEvents:  []subscribed.Event{},
		}, {
			name:            "should return 2 events(out of multiple) which matches two sources (bebNamespace and empty)",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "foo-match",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromDiffSource),
			expectedEvents: []subscribed.Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		}, {
			name:            "should return 2 out 3 events in a slice when filters with different eventTypePrefix are provided",
			appName:         "foovarkes",
			eventTypePrefix: "foo.prefix.custom",
			bebNs:           "/default/foo.kyma/kt1",
			filters:         NewBEBFilters(WithMultipleBEBFiltersFromDiffEventTypePrefix),
			expectedEvents: []subscribed.Event{
				NewEvent("order.created", "v1"),
				NewEvent("order.created", "v1"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEvents := subscribed.FilterEventTypeVersions(tc.eventTypePrefix, tc.bebNs, tc.appName, tc.filters)
			if !reflect.DeepEqual(tc.expectedEvents, gotEvents) {
				t.Errorf("Received incorrect events, Wanted: %v, Got: %v", tc.expectedEvents, gotEvents)
			}
		})
	}
}

func TestConvertEventsMapToSlice(t *testing.T) {
	testCases := []struct {
		name         string
		inputMap     map[subscribed.Event]bool
		wantedEvents []subscribed.Event
	}{
		{
			name: "should return events from the map in a slice",
			inputMap: map[subscribed.Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v2"): true,
			},
			wantedEvents: []subscribed.Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v2"),
			},
		}, {
			name:         "should return no events for an empty map of events",
			inputMap:     map[subscribed.Event]bool{},
			wantedEvents: []subscribed.Event{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEvents := subscribed.ConvertEventsMapToSlice(tc.inputMap)
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
		eventsSubSet           []subscribed.Event
		givenUniqEventsAlready map[subscribed.Event]bool
		wantedUniqEvents       map[subscribed.Event]bool
	}{
		{
			name: "should return unique events along with the existing ones",
			eventsSubSet: []subscribed.Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v1"),
			},
			givenUniqEventsAlready: map[subscribed.Event]bool{
				NewEvent("bar-already-existing", "v1"): true,
			},
			wantedUniqEvents: map[subscribed.Event]bool{
				NewEvent("foo", "v1"):                  true,
				NewEvent("bar", "v1"):                  true,
				NewEvent("bar-already-existing", "v1"): true,
			},
		}, {
			name: "should return unique new events from the subset provided only",
			eventsSubSet: []subscribed.Event{
				NewEvent("foo", "v1"),
				NewEvent("bar", "v1"),
			},
			givenUniqEventsAlready: nil,
			wantedUniqEvents: map[subscribed.Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
		}, {
			name:         "should return existing unique events when an empty subset provided",
			eventsSubSet: []subscribed.Event{},
			givenUniqEventsAlready: map[subscribed.Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
			wantedUniqEvents: map[subscribed.Event]bool{
				NewEvent("foo", "v1"): true,
				NewEvent("bar", "v1"): true,
			},
		}, {
			name:                   "should return no unique events when an empty subset provided",
			eventsSubSet:           []subscribed.Event{},
			givenUniqEventsAlready: map[subscribed.Event]bool{},
			wantedUniqEvents:       map[subscribed.Event]bool{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotUniqEvents := subscribed.AddUniqueEventsToResult(tc.eventsSubSet, tc.givenUniqEventsAlready)
			if !reflect.DeepEqual(tc.wantedUniqEvents, gotUniqEvents) {
				t.Errorf("incorrect unique events, wanted: %v, got: %v", tc.wantedUniqEvents, gotUniqEvents)
			}
		})
	}
}

type BEBFilterOption func(filter *eventing.BEBFilters)

func NewBEBFilters(opts ...BEBFilterOption) *eventing.BEBFilters {
	newFilters := &eventing.BEBFilters{}
	for _, opt := range opts {
		opt(newFilters)
	}

	return newFilters
}

func WithOneBEBFilter(bebFilters *eventing.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventing.BEBFilter{
		NewBEBFilter(evSource, evType),
	}

}

func WithMultipleBEBFiltersFromSameSource(bebFilters *eventing.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventing.BEBFilter{
		NewBEBFilter(evSource, evType),
		NewBEBFilter(evSource, evType),
		NewBEBFilter(evSource, evType),
	}
}

func WithMultipleBEBFiltersFromDiffSource(bebFilters *eventing.BEBFilters) {
	evSource1 := "foo-match"
	evSource2 := "/default/foo.different/kt1"
	evSource3 := "/default/foo.different2/kt1"
	evSource4 := ""
	evType := "foo.prefix.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventing.BEBFilter{
		NewBEBFilter(evSource1, evType),
		NewBEBFilter(evSource2, evType),
		NewBEBFilter(evSource3, evType),
		NewBEBFilter(evSource4, evType),
	}
}

func WithMultipleBEBFiltersFromDiffEventTypePrefix(bebFilters *eventing.BEBFilters) {
	evSource := "/default/foo.kyma/kt1"
	evType1 := "foo.prefix.custom.foovarkes.order.created.v1"
	evType2 := "foo.prefixdifferent.custom.foovarkes.order.created.v1"
	bebFilters.Filters = []*eventing.BEBFilter{
		NewBEBFilter(evSource, evType1),
		NewBEBFilter(evSource, evType2),
		NewBEBFilter(evSource, evType1),
	}
}

func NewBEBFilter(evSource, evType string) *eventing.BEBFilter {
	return &eventing.BEBFilter{
		EventSource: &eventing.Filter{
			Property: "source",
			Value:    evSource,
		},
		EventType: &eventing.Filter{
			Property: "type",
			Value:    evType,
		},
	}
}

func NewEvent(name, version string) subscribed.Event {
	return subscribed.Event{
		Name:    name,
		Version: version,
	}
}
