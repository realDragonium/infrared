package service

type UserService interface {
	CreateUser(user *User) bool
	User(username string) User
	Users() []User
	UpdateUser(user *User)
	DeleteUser(user *User)
}

type User struct {
	Model
	Username  string `json:"username" binding:"omitempty,max=32" gorm:"type:varchar(32);unique_index;not null"`
	Password  []byte `json:"-" binding:"omitempty,max=72" gorm:"type:binary(60);not null"`
	Roles     []Role `json:"roles,omitempty" gorm:"many2many:user_role"`
	Email     string `json:"email,omitempty" binding:"omitempty,email" gorm:"type:varchar(254)"`
	IsAdmin   bool   `json:"isAdmin"`
	CanCreate bool   `json:"canCreate"`
}
