package promAPI

// To parse JSON response from http://monitoring-prometheus.kyma-system:9090/api/v1/targets/metadata
type TargetsMetaDataResponse struct {
	Status    string           `json:"status"`
	Data      []TargetMetaData `json:"data"`
	ErrorType string           `json:"errorType"`
	Error     string           `json:"error"`
}

type TargetMetaData struct {
	Target Target `json:"target"`
}

type Target struct {
	Endpoint  string `json:endpoint`
	Instance  string `json:instancce`
	Job       string `json:"job"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Service   string `json:"service"`
}
