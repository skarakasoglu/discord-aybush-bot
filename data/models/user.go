package models

type User struct{
	Id int
	Username string
	Password string
	Email string
	IsActive bool
	Roles []Role
}
