package delivery

// Types for request and responses
type (
	// CreateRegisterRequest struct
	CreateRegisterRequest struct {
		Email     string `json:"email"`
		Passwords string `json:"passwords"`
	}
	// CreateRegisterResponse struct
	CreateRegisterResponse struct {
		Status string `json:"status"`
	}
	// CreateLoginRequest struct
	CreateLoginRequest struct {
		Email     string `json:"email"`
		Passwords string `json:"passwords"`
	}
	// CreateLoginResponse struct
	CreateLoginResponse struct {
		Status string `json:"status"`
	}
)
