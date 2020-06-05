package service

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type MySQLConfig struct {
	Address  string
	Database string
	Username string
	Password string
}

type MySQL struct {
	*gorm.DB
}

func NewMySQL(cfg MySQLConfig) (Service, error) {
	connStr := fmt.Sprintf(
		"%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Address,
		cfg.Database,
	)

	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	db.LogMode(false)
	db.SingularTable(true)

	db.
		Set("gorm:table_options", "ENGINE=InnoDB").
		Set("gorm:table_options", "CHARSET=utf8mb4").
		AutoMigrate(&User{}, &Role{}, &Proxy{}, &Permission{})

	db.Create(&PermissionNone)
	db.Create(&PermissionView)
	db.Create(&PermissionEdit)

	return &MySQL{DB: db}, nil
}

// UserService
func (db MySQL) CreateUser(user *User) bool {
	db.Create(user)
	return !db.NewRecord(*user)
}

func (db MySQL) User(username string) User {
	var user User
	db.Where("username = ?", username).First(&user)
	return user
}

func (db MySQL) Users() []User {
	var users []User
	db.Find(&users)
	return users
}

func (db MySQL) UpdateUser(user *User) {
	db.Save(user)
}

func (db MySQL) DeleteUser(user *User) {
	db.Unscoped().Delete(user)
}
