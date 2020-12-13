package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Detail     string `json:"detail"`
	Error      int    `json:"error"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
