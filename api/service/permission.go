package service

type Permission struct {
	Model
	Name string `json:"name" gorm:"varchar(4)"`
}

var (
	PermissionNone = Permission{Name: "None"}
	PermissionView = Permission{Name: "View"}
	PermissionEdit = Permission{Name: "Edit"}
)
