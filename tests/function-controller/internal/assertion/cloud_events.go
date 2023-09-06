package assertion

import (
	"bytes"
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"time"
)

const EventTypeParam = "type"

/*
*
https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md
*/
type cloudEventResponse struct {
	// Required
	CeType string `json:"ce-type"`
	// Required
	CeSource string `json:"ce-source"`
	// Required
	CeSpecVersion string `json:"ce-specversion"`
	// Required
	CeID string `json:"ce-id"`
	// Optional
	CeTime string `json:"ce-time"`
	// Optional
	CeDataContentType string `json:"ce-datacontenttype"`
	// Extension field, optional.
	CeEventTypeVersion string         `json:"ce-eventtypeversion"`
	Data               cloudEventData `json:"data"`
}

type cloudEventData struct {
	Hello string `json:"hello"`
}

type cloudEventCheck struct {
	name     string
	log      *logrus.Entry
	endpoint string
	encoding cloudevents.Encoding
}

var _ executor.Step = &cloudEventCheck{}

func CloudEventReceiveCheck(log *logrus.Entry, name string, encoding cloudevents.Encoding, target *url.URL) *cloudEventCheck {
	return &cloudEventCheck{
		encoding: encoding,
		name:     name,
		log:      log.WithField(executor.LogStepKey, name),
		endpoint: target.String(),
	}
}

func (ce cloudEventCheck) Name() string {
	return ce.name
}

func (ce cloudEventCheck) Run() error {
	ceEventType := fmt.Sprintf("test-%s", ce.encoding)
	expResp := cloudEventResponse{
		CeType:             ceEventType,
		CeSource:           "contract-test",
		CeSpecVersion:      cloudevents.VersionV1,
		CeEventTypeVersion: "v1alpha2",
	}
	ceCtx, data, err := ce.createCECtx()
	if err != nil {
		return err
	}
	expResp.Data = cloudEventData{Hello: data}

	err = sentCloudEvent(ceCtx, expResp)
	if err != nil {
		return errors.Wrap(err, "while setting cloud event data")
	}

	ceResp, err := getCloudEventFromFunction(ce.endpoint, ceEventType)
	if err != nil {
		return errors.Wrap(err, "while fetching cloud event from function")
	}
	err = assertCloudEvent(ceResp, expResp)
	if err != nil {
		return errors.Wrapf(err, "while validating cloud event")
	}
	ce.log.Info("cloud event is okay")
	return nil
}

func (ce cloudEventCheck) Cleanup() error {
	return nil
}

func (ce cloudEventCheck) OnError() error {
	return nil
}

func (ce cloudEventCheck) createCECtx() (context.Context, string, error) {
	ceCtx := cloudevents.ContextWithTarget(context.Background(), ce.endpoint)
	var data = ""
	switch ce.encoding {
	case cloudevents.EncodingStructured:
		ceCtx = cloudevents.WithEncodingStructured(ceCtx)
		data = "structured"
	case cloudevents.EncodingBinary:
		ceCtx = cloudevents.WithEncodingBinary(ceCtx)
		data = "binary"
	default:
		return nil, "", errors.Errorf("Encoding not supported: %s", ce.encoding)
	}
	return ceCtx, data, nil
}

type cloudEventSendCheck struct {
	name     string
	log      *logrus.Entry
	endpoint string
}

var _ executor.Step = &cloudEventSendCheck{}

func CloudEventSendCheck(log *logrus.Entry, name string, target *url.URL) executor.Step {
	return cloudEventSendCheck{
		name:     name,
		log:      log,
		endpoint: target.String(),
	}
}

func (s cloudEventSendCheck) Name() string {
	return s.name
}

func (s cloudEventSendCheck) Run() error {
	eventData := cloudEventData{
		Hello: "send-event",
	}

	out, err := json.Marshal(&eventData)
	if err != nil {
		return errors.Wrap(err, "while marshaling eventData to send")
	}

	resp, err := http.DefaultClient.Post(s.endpoint, "application/json", bytes.NewReader(out))
	if err != nil {
		return errors.Wrap(err, "while sending eventData")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Expected: %d, got: %d status code from eventData request", http.StatusAccepted, resp.StatusCode)
	}

	event, err := getCloudEventFromFunction(s.endpoint, "send-check")
	if err != nil {
		return errors.Wrap(err, "while getting saved cloud event")
	}
	expected := cloudEventResponse{
		CeType:             "send-check",
		CeSource:           "function",
		CeSpecVersion:      cloudevents.VersionV1,
		CeEventTypeVersion: "v1alpha2",
		Data:               eventData,
	}
	err = assertCloudEvent(event, expected)
	if err != nil {
		return errors.Wrap(err, "while doing assertion on cloud event")
	}

	return nil
}

func (s cloudEventSendCheck) Cleanup() error {
	return nil
}

func (s cloudEventSendCheck) OnError() error {
	return nil
}

func sentCloudEvent(ceCtx context.Context, expResp cloudEventResponse) error {
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		return errors.Wrap(err, "while creating cloud event client")
	}
	event := cloudevents.NewEvent()
	err = event.SetData(cloudevents.ApplicationJSON, expResp.Data)
	if err != nil {
		return errors.Wrap(err, "while setting data on cloud event")
	}
	event.SetSource(expResp.CeSource)
	event.SetType(expResp.CeType)
	event.SetSpecVersion(expResp.CeSpecVersion)
	event.SetExtension("eventtypeversion", expResp.CeEventTypeVersion)

	result := c.Send(ceCtx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.Wrap(result, "while sending cloud event")
	}
	return nil
}

func getCloudEventFromFunction(endpoint, eventType string) (cloudEventResponse, error) {
	req := &http.Request{}
	fnURL, err := url.Parse(endpoint)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while parsing function url")
	}
	q := fnURL.Query()
	q.Add(EventTypeParam, eventType)
	fnURL.RawQuery = q.Encode()

	req.URL = fnURL
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while doing GET request to function")
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while reading response body")
	}
	fmt.Println("GET response:\n", string(out))

	ceResp := cloudEventResponse{}
	err = json.Unmarshal(out, &ceResp)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while unmarshalling response")
	}
	return ceResp, nil
}

func assertCloudEvent(response cloudEventResponse, expectedResponse cloudEventResponse) error {
	var errJoined error

	if expectedResponse.Data.Hello != response.Data.Hello {
		err := errors.Errorf("Expected %s, got %s in cloud event data", expectedResponse.Data.Hello, response.Data.Hello)
		errJoined = goerrors.Join(err)
	}

	_, err := time.Parse(time.RFC3339, response.CeTime)
	if err != nil {
		errJoined = goerrors.Join(errors.Wrap(err, "while validating date"))
	}

	if response.CeSource != expectedResponse.CeSource {
		errJoined = goerrors.Join(errors.Errorf("expected source %s, got: %s", expectedResponse.CeSource, response.CeSource))
	}

	if response.CeType != expectedResponse.CeType {
		errJoined = goerrors.Join(errors.Errorf("expected type %s, got: %s", expectedResponse.CeType, response.CeType))
	}

	if response.CeSpecVersion != expectedResponse.CeSpecVersion {
		errJoined = goerrors.Join(errors.Errorf("expected spec version %s, got: %s", expectedResponse.CeSpecVersion, response.CeSpecVersion))
	}

	if response.CeEventTypeVersion != expectedResponse.CeEventTypeVersion {
		errJoined = goerrors.Join(errors.Errorf("expected event type version %s, got: %s", expectedResponse.CeEventTypeVersion, response.CeEventTypeVersion))
	}

	_, err = uuid.Parse(response.CeID)
	if err != nil {
		errJoined = goerrors.Join(errors.Errorf("expected UUID, got: %s", response.CeID))
	}
	return errJoined
}
