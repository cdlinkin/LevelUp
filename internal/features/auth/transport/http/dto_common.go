package auth_transport_http

type RegisterRequest struct {
	FullName    string  `json:"full_name" validate:"required,min=3,max=100" example:"Roman Romanovich"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=10,max=15,startswith=+" example:"+79998887766"`
	Email       string  `json:"email" validate:"required" example:"roman@mail.ru"`
	Password    string  `json:"password" validate:"required" example:"secret123"`
}

type RegisterResponse struct {
	ID          int     `json:"id" example:"1"`
	FullName    string  `json:"full_name" example:"Roman Romanovich"`
	PhoneNumber *string `json:"phone_number" example:"+79998887766"`
	Email       string  `json:"email" example:"roman@mail.ru"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required" example:"roman@mail.ru"`
	Password string `json:"password" validate:"required" example:"secret123"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"86400"`
}
