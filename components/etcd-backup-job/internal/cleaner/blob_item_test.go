package cleaner

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBlobItemSortAlg(t *testing.T) {
	// given
	fixTime := func() time.Time {
		return time.Date(1994, 04, 21, 1, 1, 1, 1, time.UTC)
	}

	tests := map[string]struct {
		givenBlobs          []simpleBlobItem
		expOrderOfBlobsName []string
	}{
		"already sorted in descending way": {
			givenBlobs: []simpleBlobItem{
				{"testA", fixTime().Add(2 * time.Hour)},
				{"testB", fixTime().Add(time.Hour)},
				{"testC", fixTime().Add(time.Second)},
				{"testD", fixTime()},
			},
			expOrderOfBlobsName: []string{"testA", "testB", "testC", "testD"},
		},
		"already sorted in ascending way": {
			givenBlobs: []simpleBlobItem{
				{"testD", fixTime()},
				{"testC", fixTime().Add(time.Second)},
				{"testB", fixTime().Add(time.Hour)},
				{"testA", fixTime().Add(2 * time.Hour)},
			},
			expOrderOfBlobsName: []string{"testA", "testB", "testC", "testD"},
		},
		"unsorted": {
			givenBlobs: []simpleBlobItem{
				{"testC", fixTime().Add(time.Second)},
				{"testB", fixTime().Add(time.Hour)},
				{"testD", fixTime()},
				{"testA", fixTime().Add(2 * time.Hour)},
			},
			expOrderOfBlobsName: []string{"testA", "testB", "testC", "testD"},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			sort.Sort(byLastModifiedDesc(tc.givenBlobs))
			// then
			for idx, name := range tc.expOrderOfBlobsName {
				assert.Equal(t, tc.givenBlobs[idx].Name, name)
			}
		})
	}
}
