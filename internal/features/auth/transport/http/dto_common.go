package auth_transport_http

type RegisterRequest struct {
	FullName    string  `json:"full_name" validate:"required,min=3,max=100"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=10,max=15,startswith=+"`
	Email       string  `json:"email" validate:"required" `
	Password    string  `json:"password" validate:"required"`
}

type RegisterResponse struct {
	ID          int     `json:"id"`
	FullName    string  `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Email       string  `json:"email"`
}
