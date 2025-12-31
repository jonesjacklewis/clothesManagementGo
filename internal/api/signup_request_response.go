package api

type SignupRequest struct {
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

type SignupResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId,omitempty"`
}
