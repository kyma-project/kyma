package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
)

// GetCleanSubjects returns a list of clean eventTypes from the unique filters in the subscription.
func GetCleanSubjects(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner) ([]string, error) {
	var filters []*eventingv1alpha1.BEBFilter
	if sub.Spec.Filter != nil {
		uniqueFilters, err := sub.Spec.Filter.Deduplicate()
		if err != nil {
			return []string{}, errors.Wrap(err, "deduplicate subscription filters failed")
		}
		filters = uniqueFilters.Filters
	}

	var cleanSubjects []string
	for _, filter := range filters {
		subject, err := GetCleanSubject(filter, cleaner)
		if err != nil {
			return []string{}, err
		}
		cleanSubjects = append(cleanSubjects, subject)
	}
	return cleanSubjects, nil
}

func createKeySuffix(subject string, queueGoupInstanceNo int) string {
	return subject + string(types.Separator) + strconv.Itoa(queueGoupInstanceNo)
}

func CreateKey(sub *eventingv1alpha1.Subscription, subject string, queueGoupInstanceNo int) string {
	return fmt.Sprintf("%s.%s", CreateKeyPrefix(sub), createKeySuffix(subject, queueGoupInstanceNo))
}

func GetCleanSubject(filter *eventingv1alpha1.BEBFilter, cleaner eventtype.Cleaner) (string, error) {
	eventType := strings.TrimSpace(filter.EventType.Value)
	if len(eventType) == 0 {
		return "", nats.ErrBadSubject
	}
	// clean the application name segment in the event-type from none-alphanumeric characters
	// return it as a NATS subject
	return cleaner.Clean(eventType)
}

func CreateKymaSubscriptionNamespacedName(key string, sub *nats.Subscription) types.NamespacedName {
	nsn := types.NamespacedName{}
	nnvalues := strings.Split(key, string(types.Separator))
	nsn.Namespace = nnvalues[0]
	nsn.Name = strings.TrimSuffix(strings.TrimSuffix(nnvalues[1], sub.Subject), ".")
	return nsn
}

// IsNatsSubAssociatedWithKymaSub checks if the NATS subscription is associated / related to Kyma subscription or not.
func IsNatsSubAssociatedWithKymaSub(natsSubKey string, natsSub *nats.Subscription, sub *eventingv1alpha1.Subscription) bool {
	return CreateKeyPrefix(sub) == CreateKymaSubscriptionNamespacedName(natsSubKey, natsSub).String()
}

func ConvertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	if err := event.Validate(); err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateKeyPrefix(sub *eventingv1alpha1.Subscription) string {
	namespacedName := types.NamespacedName{
		Namespace: sub.Namespace,
		Name:      sub.Name,
	}
	return namespacedName.String()
}

func CreateEventTypeCleaner(eventTypePrefix, applicationName string, logger *logger.Logger) eventtype.Cleaner { //nolint:unparam
	application := applicationtest.NewApplication(applicationName, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	return eventtype.NewCleaner(eventTypePrefix, applicationLister, logger)
}
