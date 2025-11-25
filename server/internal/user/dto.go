package user

type CreateUser struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse is returned to the client (refresh token is in httpOnly cookie)
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` // seconds until access token expires
}

// InternalTokenResponse is used internally to pass both tokens
type InternalTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
