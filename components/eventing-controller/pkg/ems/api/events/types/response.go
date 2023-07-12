package types

type Response struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
}

type PublishResponse struct {
	Response
	Successful []string `json:"successful,omitempty"`
	Failed     []string `json:"failed,omitempty"`
}

type CreateResponse struct {
	Response
	Href string `json:"href,omitempty"`
}

type UpdateResponse struct {
	Response
}

type DeleteResponse struct {
	Response
}

type UpdateStateResponse struct {
	Response
	Href string `json:"href,omitempty"`
}

type TriggerHandshake struct {
	Response
}
