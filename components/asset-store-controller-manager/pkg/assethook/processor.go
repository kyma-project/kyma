package assethook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	pkgPath "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/path"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type Result struct {
	Success  bool
	Messages map[string][]Message
}

type Message struct {
	Filename string
	Message  string
}

type processor struct {
	onSuccess      func(ctx context.Context, basePath, filePath string, responseBody io.Reader, messagesChan chan Message, errChan chan error)
	onFail         func(ctx context.Context, basePath, filePath string, responseBody io.Reader, messagesChan chan Message, errChan chan error)
	workers        int
	continueOnFail bool
	timeout        time.Duration
	httpClient     HttpClient
}

//go:generate mockery -name=HttpClient -output=automock -outpkg=automock -case=underscore
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//go:generate mockery -name=httpProcessor -output=automock -outpkg=automock -case=underscore
type httpProcessor interface {
	Do(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (map[string][]Message, error)
}

func (*processor) parseParameters(metadata *runtime.RawExtension) string {
	if nil == metadata {
		return ""
	}

	return string(metadata.Raw)
}

func (p *processor) iterateFiles(files []string, filter string) (chan string, error) {
	filtered, err := pkgPath.Filter(files, filter)
	if err != nil {
		return nil, errors.Wrapf(err, "while filtering files with regex %s", filter)
	}

	fileNameChan := make(chan string, len(filtered))
	defer close(fileNameChan)
	for _, fileName := range filtered {
		fileNameChan <- fileName
	}

	return fileNameChan, nil
}

func (p *processor) Do(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (map[string][]Message, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(map[string][]Message)
	for _, service := range services {
		success, messages, err := p.doService(ctx, cancel, basePath, files, service)
		if err != nil {
			return nil, err
		}
		if !success {
			name := fmt.Sprintf("%s/%s%s", service.Namespace, service.Name, service.Endpoint)
			results[name] = messages
		}
	}

	return results, nil
}

func (p *processor) doService(ctx context.Context, cancel context.CancelFunc, basePath string, files []string, service v1alpha2.AssetWebhookService) (bool, []Message, error) {
	fileChan, err := p.iterateFiles(files, service.Filter)
	if err != nil {
		return false, nil, errors.Wrap(err, "while creating files channel")
	}
	messagesChan := make(chan Message)
	errChan := make(chan error)
	go func() {
		defer close(messagesChan)
		defer close(errChan)

		var waitGroup sync.WaitGroup
		for i := 0; i < p.workers; i++ {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				p.doFiles(ctx, cancel, basePath, service, fileChan, messagesChan, errChan)
			}()
		}
		waitGroup.Wait()
	}()

	var waitGroup sync.WaitGroup
	var errs []error
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for e := range errChan {
			errs = append(errs, e)
		}
	}()

	var messages []Message
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for msg := range messagesChan {
			messages = append(messages, msg)
		}
	}()
	waitGroup.Wait()

	if len(errs) > 0 {
		msg := errs[0].Error()
		for _, e := range errs[1:] {
			msg = fmt.Sprintf("%s, %s", msg, e.Error())
		}
		return false, nil, errors.New(msg)
	}

	if len(messages) == 0 {
		return true, nil, nil
	}
	return false, messages, nil
}

func (p *processor) doFiles(ctx context.Context, cancel context.CancelFunc, basePath string, service v1alpha2.AssetWebhookService, pathChan chan string, messagesChan chan Message, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-errChan:
			return
		case path, ok := <-pathChan:
			if !ok {
				return
			}

			p.doFile(ctx, cancel, basePath, path, service, messagesChan, errChan)
		}
	}
}

func (p *processor) doFile(ctx context.Context, cancel context.CancelFunc, basePath string, path string, service v1alpha2.AssetWebhookService, messagesChan chan Message, errChan chan error) {
	body, contentType, err := p.buildQuery(basePath, path, p.parseParameters(service.Parameters))
	if err != nil {
		errChan <- errors.Wrap(err, "while building multipart query")
		return
	}

	success, modified, rspBody, err := p.call(ctx, contentType, service.WebhookService, body)
	if err != nil {
		errChan <- errors.Wrap(err, "while sending request to webhook")
		return
	}
	defer rspBody.Close()

	if success && modified && p.onSuccess != nil {
		p.onSuccess(ctx, basePath, path, rspBody, messagesChan, errChan)
	} else if !success && p.onFail != nil {
		p.onFail(ctx, basePath, path, rspBody, messagesChan, errChan)
	}

	if !success && !p.continueOnFail {
		cancel()
	}
}

func (p *processor) buildQuery(basePath, filePath, parameters string) (io.Reader, string, error) {
	buffer := &bytes.Buffer{}
	formWriter := multipart.NewWriter(buffer)
	defer formWriter.Close()

	path := filepath.Join(basePath, filePath)
	file, err := os.Open(path)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while opening file %s", filePath)
	}
	defer file.Close()

	contentWriter, err := formWriter.CreateFormFile("content", filepath.Base(file.Name()))
	if err != nil {
		return nil, "", errors.Wrapf(err, "while creating content field for file %s", filePath)
	}

	_, err = io.Copy(contentWriter, file)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while copying file %s to content field", filePath)
	}

	err = formWriter.WriteField("parameters", parameters)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while creating parameters field for parameters %s", parameters)
	}

	return buffer, formWriter.FormDataContentType(), nil
}

func (p *processor) call(ctx context.Context, contentType string, webhook v1alpha2.WebhookService, body io.Reader) (bool, bool, io.ReadCloser, error) {
	context, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequest("POST", p.getWebhookUrl(webhook), body)
	if err != nil {
		return false, false, nil, errors.Wrap(err, "while creating request")
	}

	req.Header.Set("Content-Type", contentType)
	req.WithContext(context)

	rsp, err := p.httpClient.Do(req)
	if err != nil {
		return false, false, nil, errors.Wrapf(err, "while sending request to webhook")
	}

	switch rsp.StatusCode {
	case http.StatusOK, http.StatusUnprocessableEntity:
		success := rsp.StatusCode == http.StatusOK
		return success, success, rsp.Body, nil
	case http.StatusNotModified:
		return true, false, rsp.Body, nil
	default:
		return false, false, rsp.Body, fmt.Errorf("invalid response from %s, code: %d", req.URL, rsp.StatusCode)
	}
}

func (*processor) getWebhookUrl(service v1alpha2.WebhookService) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, service.Endpoint)
}
