package models

import "time"

type Member struct{
	Id int
	MemberId string
	Email string
	Username string
	Discriminator string
	IsVerified bool
	IsBot bool
	Left bool
	JoinedAt time.Time
}