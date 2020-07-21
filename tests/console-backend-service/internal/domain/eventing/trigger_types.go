package eventing

type TriggerListQueryResponse struct {
	Triggers []Trigger
}

type TriggerEvent struct {
	Type    string  `json:"type"`
	Trigger Trigger `json:"trigger"`
}

type Trigger struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Spec      TriggerSpec   `json:"spec""`
	Status    TriggerStatus `json:"status"`
}

type TriggerSpec struct {
	Broker     string                 `json:"broker"`
	Filter     map[string]interface{} `json:"filter"`
	Subscriber Subscriber             `json:"subscriber"`
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
