package cleaner

import (
	"context"
	"sort"
	"time"

	pTime "github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/time"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	maxResultInSingleQuery      = 500
	maxAllowedQueriesForSegment = 100
)

// ErrBlobsLimitNumberExceeded means that in given container is more blobs than expected.
// It's used to ensure that application will not load too much blobs in RAM
var ErrBlobsLimitNumberExceeded = errors.Errorf("Container contains more than %d blobs.", maxAllowedQueriesForSegment*maxResultInSingleQuery)

//go:generate mockery -name=azureBlobClient -output=automock -outpkg=automock -case=underscore
type azureBlobClient interface {
	ListBlobsFlatSegment(context.Context, azblob.Marker, azblob.ListBlobsSegmentOptions) (*azblob.ListBlobsFlatSegmentResponse, error)
	DeleteBlockBlob(context.Context, string, azblob.DeleteSnapshotsOptionType, azblob.BlobAccessConditions) (*azblob.BlobDeleteResponse, error)
}

// AzureCleaner removes old backup files from given ABS
type AzureCleaner struct {
	log                logrus.FieldLogger
	blobCli            azureBlobClient
	minBackupBlobs     int64
	expirationBlobTime time.Duration

	nowProvider pTime.NowProvider
}

// NewAzure returns new instance of AzureCleaner
func NewAzure(cfg Config, blobCli azureBlobClient, log logrus.FieldLogger) *AzureCleaner {
	return &AzureCleaner{
		log:                log,
		blobCli:            blobCli,
		minBackupBlobs:     cfg.LeaveMinNewestBackupBlobs,
		expirationBlobTime: cfg.ExpirationBlobTime,
	}
}

// Clean removes from ABS old backup files
func (c *AzureCleaner) Clean(stopCh <-chan struct{}, blobPrefix string) error {
	c.log.Debug("Starting clean-up process")
	defer c.log.Debug("Clean-up process completed")

	ctx := c.cancellableCtxOnInterrupt(stopCh)

	c.log.Debugf("Collecting all blobs from ABS container")
	blobs, err := c.getBlobs(ctx, blobPrefix)
	if err != nil {
		return errors.Wrap(err, "while getting blobs information from azure container")
	}

	blobsToDelete := c.extractBlobsToDeletion(blobs)
	c.log.Debugf("Deleting blobs from ABS container: %+v", blobsToDelete)
	if err := c.deleteBlobs(ctx, blobsToDelete); err != nil {
		return errors.Wrap(err, "while deleting blobs from azure container")
	}

	return nil
}

func (c *AzureCleaner) getBlobs(ctx context.Context, blobPrefix string) ([]simpleBlobItem, error) {
	var simpleBlobs []simpleBlobItem

	// Iterate over the blob(s) in container.
	// Since a container may hold millions of blobs, this is done 1 segment at a time.
	// Additionally we have a guard to do not to iterate indefinitely (max blobs maxAllowedQueriesForSegment*maxResultInSingleQuery)
	for cnt, marker := 0, (azblob.Marker{}); marker.NotDone(); cnt++ {
		if cnt >= maxAllowedQueriesForSegment {
			return nil, ErrBlobsLimitNumberExceeded
		}

		// for more details about used opinions, see: https://docs.microsoft.com/pl-pl/rest/api/storageservices/list-blobs
		opt := azblob.ListBlobsSegmentOptions{
			MaxResults: maxResultInSingleQuery,
			Prefix:     blobPrefix,
		}
		listBlob, err := c.blobCli.ListBlobsFlatSegment(ctx, marker, opt)
		if err != nil {
			return nil, errors.Wrap(err, "while listing blobs from azure container")
		}

		for _, blobInfo := range listBlob.Segment.BlobItems {
			simpleBlobs = append(simpleBlobs, simpleBlobItem{
				Name:         blobInfo.Name,
				LastModified: blobInfo.Properties.LastModified,
			})
		}

		// ListBlobs returns the start of the next segment;
		// we MUST use this to get the next segment (after processing the current result segment).
		marker = listBlob.NextMarker
	}

	return simpleBlobs, nil
}

func (c *AzureCleaner) extractBlobsToDeletion(blobs []simpleBlobItem) []simpleBlobItem {
	// STEP 0: check if we have enough blobs to execute deletion process
	if int64(len(blobs)) <= c.minBackupBlobs {
		return []simpleBlobItem{}
	}

	// STEP 1: sort (newest blobs precede the oldest ones)
	sort.Sort(byLastModifiedDesc(blobs))

	// STEP 2: ensure that we always will have `minBackupBlobs` available
	blobs = blobs[c.minBackupBlobs:]

	// STEP 3: collect blobs which are too old
	thresholdTime := c.calcThresholdTime()
	var blobsToDelete []simpleBlobItem
	for _, b := range blobs {
		if b.LastModified.After(thresholdTime) {
			continue
		}
		blobsToDelete = append(blobsToDelete, b)
	}

	return blobsToDelete
}

func (c *AzureCleaner) calcThresholdTime() time.Time {
	return c.nowProvider.Now().Add(-c.expirationBlobTime)
}

// deleteBlobs deletes the blobs from ABS by given blob name. If error is returned for single file then it fail fast.
func (c *AzureCleaner) deleteBlobs(ctx context.Context, blobs []simpleBlobItem) error {
	for _, b := range blobs {
		_, err := c.blobCli.DeleteBlockBlob(ctx, b.Name, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
		if err != nil {
			return errors.Wrapf(err, "while deleting blob %q", b.Name)
		}
	}
	return nil
}

// cancellableCtxOnInterrupt returns context which will be closed when we received a signal on given stop channel
func (*AzureCleaner) cancellableCtxOnInterrupt(stop <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-stop:
			cancel()
		}
	}()

	return ctx
}
