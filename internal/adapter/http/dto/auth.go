package dto

// LoginRequest is the request body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse is the response body for a successful login.
type LoginResponse struct {
	Token string `json:"token"`
}
