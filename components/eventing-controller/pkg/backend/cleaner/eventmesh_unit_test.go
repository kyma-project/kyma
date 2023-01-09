package cleaner //nolint:testpackage

import (
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

func Test_CleanEventType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "success if the given event has more than two segments",
			givenEventType: "Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:  "Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
		},
		{
			name:           "success if the given event have non-alphanumeric characters",
			givenEventType: "Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:  "Segment1Part1Part2.Segment2Part1Part2.v1",
		},
		{
			name:           "success if the given event have less than three segments",
			givenEventType: "Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä",
			wantEventType:  "Segment1Part1Part2.Segment2Part1Part2",
		},
	}

	defaultLogger, err1 := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err1)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cleaner := NewEventMeshCleaner(defaultLogger)
			eventType, err := cleaner.CleanEventType(tc.givenEventType)
			require.NoError(t, err)
			require.Equal(t, tc.wantEventType, eventType)
		})
	}
}

func Test_CleanSource(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		givenSource string
		wantSource  string
	}{
		{
			name:        "success if the given event have non-alphanumeric characters",
			givenSource: "Source1-Part1-Part2-Ä",
			wantSource:  "Source1Part1Part2",
		},
		{
			name:        "success if the given source has more than one segment",
			givenSource: "Source1-Part1-Part2-Ä.Source2-Part1-Part2-Ä",
			wantSource:  "Source1Part1Part2Source2Part1Part2",
		},
		{
			name:        "success if the given source has more than two segments",
			givenSource: "Source1-Part1-Part2-Ä.Source2-Part1-Part2-Ä.Source3-Part1-Part2-Ä",
			wantSource:  "Source1Part1Part2Source2Part1Part2Source3Part1Part2",
		},
	}

	defaultLogger, err1 := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err1)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cleaner := NewEventMeshCleaner(defaultLogger)
			eventType, _ := cleaner.CleanSource(tc.givenSource)
			require.Equal(t, tc.wantSource, eventType)
		})
	}
}
