package handlers

import (
	"net/url"
	"testing"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/nats-io/nats.go"
)

func TestConvertMsgToCE(t *testing.T) {
	validJsonCE := `{"data":"test","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","time":"2020-12-29T12:33:38.882056251Z","type":"sap.kyma.custom.varkes.order.created.v1"}`
	dataContentType := "application/json"
	testCases := []struct {
		natsMsg            *nats.Msg
		expectedCloudEvent *cev2event.Event
		expectedErr        error
	}{{
		natsMsg: &nats.Msg{
			Subject: "foo",
			Reply:   "",
			Header:  nil,
			Data:    []byte(validJsonCE),
			Sub:     nil,
		},
		expectedCloudEvent: &cev2event.Event{
			Context: &cev2event.EventContextV1{
				Type: "sap.kyma.custom.varkes.order.created.v1",
				Source: types.URIRef{
					URL: url.URL{
						Path: "source",
					},
				},
				ID:              "id",
				DataContentType: &dataContentType,
			},
			DataEncoded: []byte("\"test\""),
			DataBase64:  false,
			FieldErrors: nil,
		},
		expectedErr: nil,
	},
	}
	for _, tc := range testCases {
		gotCE, err := convertMsgToCE(tc.natsMsg)
		if err != nil && tc.expectedErr == nil {
			t.Errorf("should not give error, got: %v", err)
		}
		if err != nil && tc.expectedErr.Error() != err.Error() {
			t.Errorf("received wrong error, expected: %v got: %v", tc.expectedErr, err)
		}
		if !(gotCE.Subject() == tc.expectedCloudEvent.Subject()) ||
			!(gotCE.ID() == tc.expectedCloudEvent.ID()) ||
			!(gotCE.DataContentType() == tc.expectedCloudEvent.DataContentType()) ||
			!(gotCE.Source() == tc.expectedCloudEvent.Source()) ||
			!(string(gotCE.Data()) == string(tc.expectedCloudEvent.Data())) {
			t.Errorf("received wrong cloudevent, expected: %v got: %v", tc.expectedCloudEvent, gotCE)
		}
	}
}
