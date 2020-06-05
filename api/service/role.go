package service

type Role struct {
	Model
	Name        string     `json:"description" binding:"required,max=32" gorm:"varchar(32);unique;not null"`
	Description string     `json:"description" binding:"max=240" gorm:"varchar(240)"`
	Permission  Permission `json:"permission" gorm:"default:1"`
}
