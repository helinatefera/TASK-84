package request

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	CaptchaID    string `json:"captcha_id"`
	CaptchaToken string `json:"captcha_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32,alphanum"`
	Email    string `json:"email" binding:"omitempty,email,max=255"`
}

type UpdatePreferencesRequest struct {
	Locale   string `json:"locale" binding:"omitempty,max=10"`
	Timezone string `json:"timezone" binding:"omitempty,max=64"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin moderator product_analyst regular_user"`
}
