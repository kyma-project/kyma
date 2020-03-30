package eventing

type TriggerListQueryResponse struct {
	Triggers []Trigger
}

type TriggerEvent struct {
	Type    string  `json:"type"`
	Trigger Trigger `json:"trigger"`
}

type Trigger struct {
	Name             string                 `json:"name"`
	Namespace        string                 `json:"namespace"`
	Broker           string                 `json:"broker"`
	FilterAttributes map[string]interface{} `json:"filterAttributes"`
	Subscriber       Subscriber             `json:"subscriber"`
	Status           TriggerStatus          `json:"status"`
}

type Subscriber struct {
	URI *string        `json:"uri"`
	Ref *SubscriberRef `json:"ref"`
}

type SubscriberRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
}

type TriggerStatus struct {
	Reason []string `json:"reason"`
	Status string   `json:"status"`
}
