package responses

// ApplicationResponse Model
type ApplicationResponse struct {
	Status  int         `json:"status"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
