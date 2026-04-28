package account

type Account struct {
	ID       uint   `grom:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"-"`
	Token    string `json:"-"`
}
type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RenameRequest struct {
	NewUsername string `json:"new_username"`
}
type ChangePasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
type FindByIDRequest struct {
	ID uint `json:"id"`
}
type FindByIDResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}
type FindByUsernameRequest struct {
	Username string `json:"username"`
}
type FindByUsernameResponse struct {
	ID       uint   `json:"id"`
	Usernaem string `json:"usernaem"`
}
type LoginResquest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
