package serverless

type FunctionListQueryResponse struct {
	Functions []Function
}

type FunctionEvent struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	UID          string            `json:"UID"`
	Labels       map[string]string `json:"labels"`
	Source       string            `json:"source"`
	Dependencies string            `json:"dependencies"`
	Env          []FunctionEnv     `json:"env"`
	Replicas     FunctionReplicas  `json:"replicas"`
	Resources    FunctionResources `json:"resources"`
	Status       FunctionStatus    `json:"status"`
}

type FunctionMetadataInput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type FunctionEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FunctionReplicas struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

type FunctionResources struct {
	Limits   ResourceValues `json:"limits"`
	Requests ResourceValues `json:"requests"`
}

type ResourceValues struct {
	Memory *string `json:"memory"`
	CPU    *string `json:"cpu"`
}

type FunctionStatus struct {
	Phase   string  `json:"phase"`
	Reason  *string `json:"reason"`
	Message *string `json:"message"`
}
