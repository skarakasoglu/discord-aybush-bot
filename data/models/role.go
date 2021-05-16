package models

type Role struct{
	Id int
	Name string
	Permissions []Permission
}