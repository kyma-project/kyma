package subscribed

import (
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

const processorName = "processor"

type Processor struct {
	SubscriptionLister *cache.GenericLister
	Config             *env.BebConfig
	Logger             *logger.Logger
}

func (p Processor) ExtractEventsFromSubscriptions(writer http.ResponseWriter, request *http.Request) {
	eventsMap := make(map[Event]bool)
	subsList, err := (*p.SubscriptionLister).List(labels.Everything())
	if err != nil {
		p.namedLogger().Errorw("Failed to fetch subscriptions", "error", err)
		RespondWithErrorAndLog(err, writer)
		return
	}

	appName := legacy.ParseApplicationNameFromPath(request.URL.Path)
	for _, sObj := range subsList {
		sub, err := ConvertRuntimeObjToSubscription(sObj)
		if err != nil {
			p.namedLogger().Errorw("Failed to convert a runtime obj to a Subscription", "error", err)
			continue
		}
		if sub.Spec.Filter != nil {
			eventsForSub := FilterEventTypeVersions(p.Config.EventTypePrefix, p.Config.BEBNamespace, appName, sub.Spec.Filter)
			eventsMap = AddUniqueEventsToResult(eventsForSub, eventsMap)
		}
	}
	events := ConvertEventsMapToSlice(eventsMap)
	RespondWithBody(writer, Events{
		EventsInfo: events,
	}, http.StatusOK)
}

func (p Processor) namedLogger() *zap.SugaredLogger {
	return p.Logger.WithContext().Named(processorName)
}
