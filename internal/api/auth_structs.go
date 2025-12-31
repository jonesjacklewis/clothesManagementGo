package api

type SignupRequest struct {
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

type SignupResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	IdToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
