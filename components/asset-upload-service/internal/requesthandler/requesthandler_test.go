package requesthandler_test

import (
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/requesthandler"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader/automock"
	"testing"
	"time"
)

func TestRequestHandler_ServeHTTP(t *testing.T) {
	client := &automock.MinioClient{}
	buckets := bucket.SystemBucketNames{
		Private:"private",
		Public:"public",
	}

	handler := requesthandler.New(client, buckets, "https://example.com", 1*time.Second, 5)

	//TODO: Write tests
	//handler.ServeHTTP()


}