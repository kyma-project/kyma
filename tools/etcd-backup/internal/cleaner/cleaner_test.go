package cleaner_test

import (
	"context"
	"encoding/xml"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tools/etcd-backup/internal/cleaner"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/cleaner/automock"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/logger/spy"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAzureCleanerCleanSuccess(t *testing.T) {
	var (
		fixBlobPrefix = "fix-prefix"
		timeNow       = fixNowProvider()
		fixTime       = timeNow()
	)

	type given struct {
		expirationBlobTime        time.Duration
		leaveMinNewestBackupBlobs int64
		blobItems                 []azblob.BlobItem
	}
	type expected struct {
		blobNamesToDelete []string
	}

	tests := map[string]struct {
		given
		expected
	}{
		"should leave minimum newest backup even if all are expired": {
			given{
				expirationBlobTime:        time.Second,
				leaveMinNewestBackupBlobs: 3,
				blobItems: []azblob.BlobItem{
					{Name: "test-1", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-2 * time.Hour)}},
					{Name: "test-2", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-3 * time.Hour)}},
					{Name: "test-3", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-4 * time.Hour)}},
				},
			},
			expected{
				blobNamesToDelete: []string{},
			},
		},
		"should delete expired backups but leave the minimum newest ones": {
			given{
				expirationBlobTime:        time.Second,
				leaveMinNewestBackupBlobs: 2,
				blobItems: []azblob.BlobItem{
					{Name: "test-1", Properties: azblob.BlobProperties{LastModified: fixTime}},
					{Name: "test-2", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-time.Hour)}},
					{Name: "test-3", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-3 * time.Hour)}},
				},
			},
			expected{
				blobNamesToDelete: []string{"test-3"},
			},
		},
		"should always delete all expired backups if leaveMinNewestBackupBlobs is set to 0": {
			given{
				expirationBlobTime:        time.Second,
				leaveMinNewestBackupBlobs: 0,
				blobItems: []azblob.BlobItem{
					{Name: "test-1", Properties: azblob.BlobProperties{LastModified: fixTime}},
					{Name: "test-2", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-time.Hour)}},
					{Name: "test-3", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-3 * time.Hour)}},
				},
			},
			expected{
				blobNamesToDelete: []string{"test-3", "test-2"},
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			azBlobCliMock := automock.NewAzureBlobClient()
			defer azBlobCliMock.AssertExpectations(t)

			opt := azblob.ListBlobsSegmentOptions{
				MaxResults: 500,
				Prefix:     fixBlobPrefix,
			}
			azBlobCliMock.On("ListBlobsFlatSegment", mock.MatchedBy(anyContext(t)), azblob.Marker{}, opt).
				Return(fixBlobsSegmentResp(tc.given.blobItems, azBlobDoneMarker()), nil).Once()
			for _, name := range tc.expected.blobNamesToDelete {
				azBlobCliMock.On("DeleteBlockBlob", mock.MatchedBy(anyContext(t)), name, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{}).
					Return(&azblob.BlobDeleteResponse{}, nil).Once()
			}

			cfg := cleaner.Config{
				ExpirationBlobTime:        tc.given.expirationBlobTime,
				LeaveMinNewestBackupBlobs: tc.given.leaveMinNewestBackupBlobs,
			}
			sut := cleaner.NewAzure(cfg, azBlobCliMock, spy.NewLogDummy()).WithTimeProvider(timeNow)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// when
			err := sut.Clean(ctx.Done(), fixBlobPrefix)

			// then
			require.NoError(t, err)
		})
	}
}

func TestAzureCleanerCleanIterateIndefinitely(t *testing.T) {
	// given
	fixBlobPrefix := "fix-prefix"

	azBlobCliMock := automock.NewAzureBlobClient()
	defer azBlobCliMock.AssertExpectations(t)

	opt := azblob.ListBlobsSegmentOptions{
		MaxResults: 500,
		Prefix:     fixBlobPrefix,
	}
	azBlobCliMock.On("ListBlobsFlatSegment", mock.MatchedBy(anyContext(t)), azblob.Marker{}, opt).
		Return(&azblob.ListBlobsFlatSegmentResponse{
			NextMarker: azBlobNotDoneMarker(),
		}, nil)

	sut := cleaner.NewAzure(cleaner.Config{}, azBlobCliMock, spy.NewLogDummy())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// when
	err := sut.Clean(ctx.Done(), fixBlobPrefix)

	// then
	assert.EqualError(t, err, "while getting blobs information from azure container: Container contains more than 50000 blobs.")
}

func TestAzureCleanerCleanFailures(t *testing.T) {
	t.Run("on listing blobs from container", func(t *testing.T) {
		// given
		var (
			fixBlobPrefix = "fix-prefix"
			fixErr        = errors.New("fix Err")
		)

		azBlobCliMock := automock.NewAzureBlobClient()
		defer azBlobCliMock.AssertExpectations(t)

		opt := azblob.ListBlobsSegmentOptions{
			MaxResults: 500,
			Prefix:     fixBlobPrefix,
		}
		azBlobCliMock.On("ListBlobsFlatSegment", mock.MatchedBy(anyContext(t)), azblob.Marker{}, opt).
			Return(nil, fixErr).Once()

		sut := cleaner.NewAzure(cleaner.Config{}, azBlobCliMock, spy.NewLogDummy())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// when
		err := sut.Clean(ctx.Done(), fixBlobPrefix)

		// then
		assert.EqualError(t, err, "while getting blobs information from azure container: while listing blobs from azure container: fix Err")
	})

	t.Run("on deleting blob from container", func(t *testing.T) {
		// given
		var (
			fixBlobPrefix = "fix-prefix"
			fixErr        = errors.New("fix Err")
			timeNow       = fixNowProvider()
			fixTime       = timeNow()
		)

		azBlobCliMock := automock.NewAzureBlobClient()
		defer azBlobCliMock.AssertExpectations(t)

		opt := azblob.ListBlobsSegmentOptions{
			MaxResults: 500,
			Prefix:     fixBlobPrefix,
		}
		blobs := []azblob.BlobItem{
			{Name: "test-1", Properties: azblob.BlobProperties{LastModified: fixTime.Add(-3 * time.Hour)}},
		}
		azBlobCliMock.On("ListBlobsFlatSegment", mock.MatchedBy(anyContext(t)), azblob.Marker{}, opt).
			Return(fixBlobsSegmentResp(blobs, azBlobDoneMarker()), nil).Once()
		azBlobCliMock.On("DeleteBlockBlob", mock.MatchedBy(anyContext(t)), "test-1", azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{}).
			Return(nil, fixErr).Once()

		cfg := cleaner.Config{
			ExpirationBlobTime: time.Second,
		}
		sut := cleaner.NewAzure(cfg, azBlobCliMock, spy.NewLogDummy()).WithTimeProvider(timeNow)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// when
		err := sut.Clean(ctx.Done(), fixBlobPrefix)

		// then
		assert.EqualError(t, err, `while deleting blobs from azure container: while deleting blob "test-1": fix Err`)
	})

}

func fixBlobsSegmentResp(b []azblob.BlobItem, m azblob.Marker) *azblob.ListBlobsFlatSegmentResponse {
	return &azblob.ListBlobsFlatSegmentResponse{
		Segment: azblob.BlobFlatList{
			BlobItems: b,
		},
		NextMarker: m,
	}
}

func anyContext(t *testing.T) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		return assert.Implements(t, (*context.Context)(nil), obj)
	}
}

func azBlobDoneMarker() azblob.Marker {
	data := []byte(`<?xml version="1.0" encoding="utf-8"?>  
<EnumerationResults ServiceEndpoint="http://myaccount.blob.core.windows.net/"  ContainerName="mycontainer">  
  <Blobs>  
    <Blob>  
      <Name>blob-name</name>    
    </Blob>  
  </Blobs>  
</EnumerationResults>`)

	var m azblob.Marker
	xml.Unmarshal(data, &m)
	return m
}

func azBlobNotDoneMarker() azblob.Marker {
	return azblob.Marker{}
}

func fixNowProvider() func() time.Time {
	return func() time.Time {
		return time.Date(1994, 04, 21, 1, 1, 1, 1, time.UTC)
	}
}
