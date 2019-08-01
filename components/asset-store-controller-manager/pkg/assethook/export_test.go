package assethook

import (
	"context"
	"io"
	"time"
)

type Callback func(ctx context.Context, basePath, filePath string, responseBody io.Reader, messagesChan chan Message, errChan chan error)

func NewProcessor(workers int, client HttpClient, continueOnFail bool, onSuccess, onFail Callback) *processor {
	return &processor{
		workers:        workers,
		httpClient:     client,
		onSuccess:      onSuccess,
		onFail:         onFail,
		continueOnFail: continueOnFail,
		timeout:        time.Minute,
	}
}

func NewTestValidator(processor httpProcessor) Validator {
	return &validationEngine{
		processor: processor,
	}
}

func NewTestMutator(processor httpProcessor) Mutator {
	return &mutationEngine{
		processor: processor,
	}
}
