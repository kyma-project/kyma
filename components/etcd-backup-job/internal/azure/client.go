package azure

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
	"github.com/pkg/errors"
)

const blobServiceURLPattern = "https://%s.blob.core.windows.net"

// BlobContainerClient is facade for creating and using the azure-storage-blob-go client.
// Allows easy mocking Azure client dependency in tests.
type BlobContainerClient struct {
	cURL azblob.ContainerURL
}

// NewBlobContainerClient returns client for Azure blob storage
func NewBlobContainerClient(absAccountName, absAccountKey, absContainerName string) (*BlobContainerClient, error) {
	blobSvcURL, err := url.Parse(fmt.Sprintf(blobServiceURLPattern, absAccountName))
	if err != nil {
		return nil, errors.Wrap(err, "while parsing blob storage URL")
	}

	var (
		cred     = azblob.NewSharedKeyCredential(absAccountName, absAccountKey)
		pipeline = azblob.NewPipeline(cred, azblob.PipelineOptions{})
		sURL     = azblob.NewServiceURL(*blobSvcURL, pipeline)
		cURL     = sURL.NewContainerURL(absContainerName)
	)

	return &BlobContainerClient{
		cURL: cURL,
	}, nil
}

// ListBlobsFlatSegment wraps the ContainerURL.ListBlobsFlatSegment
func (c *BlobContainerClient) ListBlobsFlatSegment(ctx context.Context, marker azblob.Marker,
	opts azblob.ListBlobsSegmentOptions) (*azblob.ListBlobsFlatSegmentResponse, error) {

	return c.cURL.ListBlobsFlatSegment(ctx, marker, opts)
}

// DeleteBlockBlob allows you to delete block blob by calling only one method.
func (c *BlobContainerClient) DeleteBlockBlob(ctx context.Context, blobName string,
	deleteOpt azblob.DeleteSnapshotsOptionType, ac azblob.BlobAccessConditions) (*azblob.BlobDeleteResponse, error) {

	blobURL := c.cURL.NewBlockBlobURL(blobName)
	return blobURL.Delete(ctx, deleteOpt, ac)
}
