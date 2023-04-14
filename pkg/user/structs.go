package user

type SignUpRequest struct {
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
}

type AuthenticationRequest struct {
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
}

type AuthenticationResponse struct {
	Token string `json:"token"`
}

type PasswordChangeRequest struct {
	Password string `json:"password" binding:"required"`
}

type AdditionRequest struct {
	Username string `json:"username" binding:"required"`
	Admin    bool   `json:"admin"`
}

type PasswordResponse struct {
	Password string `json:"password"`
}

type AdminChangeRequest struct {
	Admin bool `json:"admin"`
}
