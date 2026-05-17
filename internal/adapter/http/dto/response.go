package dto

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is the standard envelope for error API responses.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// OK wraps data in a SuccessResponse.
func OK(data interface{}) *SuccessResponse {
	return &SuccessResponse{Data: data}
}
