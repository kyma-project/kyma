package subscribed

import (
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
)

const processorName = "processor"

type Processor struct {
	SubscriptionLister *cache.GenericLister
	Prefix             string
	Namespace          string
	Logger             *logger.Logger
}

func (p Processor) extractEventsFromSubscriptions(
	writer http.ResponseWriter,
	request *http.Request,
	enableNewCRDVersion bool,
) {
	eventsMap := make(map[Event]bool)
	subsList, err := (*p.SubscriptionLister).List(labels.Everything())
	if err != nil {
		p.namedLogger().Errorw("Failed to fetch subscriptions", "error", err)
		RespondWithErrorAndLog(err, writer)
		return
	}

	appName := legacy.ParseApplicationNameFromPath(request.URL.Path)
	for _, sObj := range subsList {
		if enableNewCRDVersion {
			sub, err := ConvertRuntimeObjToSubscription(sObj)

			if err != nil {
				p.namedLogger().Errorw("Failed to convert a runtime obj to a Subscription", "error", err)
				continue
			}
			if sub.Spec.Types != nil {
				eventsForSub := FilterEventTypeVersions(p.Prefix, appName, sub)
				eventsMap = AddUniqueEventsToResult(eventsForSub, eventsMap)
			}
		} else {
			sub, err := ConvertRuntimeObjToSubscriptionV1alpha1(sObj)

			if err != nil {
				p.namedLogger().Errorw("Failed to convert a runtime obj to a Subscription", "error", err)
				continue
			}
			if sub.Spec.Filter != nil {
				eventsForSub := FilterEventTypeVersionsV1alpha1(p.Prefix, p.Namespace, appName, sub.Spec.Filter)
				eventsMap = AddUniqueEventsToResult(eventsForSub, eventsMap)
			}
		}
	}
	events := ConvertEventsMapToSlice(eventsMap)
	RespondWithBody(writer, Events{
		EventsInfo: events,
	}, http.StatusOK)
}

func (p Processor) ExtractEventsFromSubscriptions(writer http.ResponseWriter, request *http.Request) {
	p.extractEventsFromSubscriptions(writer, request, true)
}

func (p Processor) ExtractEventsFromSubscriptionsV1alpha1(writer http.ResponseWriter, request *http.Request) {
	p.extractEventsFromSubscriptions(writer, request, false)
}

func (p Processor) namedLogger() *zap.SugaredLogger {
	return p.Logger.WithContext().Named(processorName)
}
