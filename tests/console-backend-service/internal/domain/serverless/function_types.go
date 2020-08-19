package serverless

type FunctionListQueryResponse struct {
	Functions []Function
}

type FunctionEvent struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type FunctionMetadataInput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
