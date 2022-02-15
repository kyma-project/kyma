package subscribed

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type Processor struct {
	SubscriptionLister *cache.GenericLister
	Config             *env.BebConfig
	Logger             *logrus.Logger
}

func (p Processor) ExtractEventsFromSubscriptions(writer http.ResponseWriter, request *http.Request) {
	eventsMap := make(map[Event]bool)
	subsList, err := (*p.SubscriptionLister).List(labels.Everything())
	if err != nil {
		p.Logger.Errorf("failed to fetch subscriptions: %v", err)
		RespondWithErrorAndLog(err, writer)
		return
	}

	appName := legacy.ParseApplicationNameFromPath(request.URL.Path)
	for _, sObj := range subsList {
		sub, err := ConvertRuntimeObjToSubscription(sObj)
		if err != nil {
			p.Logger.Errorf("failed to convert a runtime obj to a Subscription: %v", err)
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
