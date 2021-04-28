package model

type UserPermission string

const (
	UserPermissionAdmin     UserPermission = "admin"
	UserPermissionDeveloper UserPermission = "developer"
)

type User struct {
	Email       string
	IsDeveloper bool
	IsAdmin     bool
}
