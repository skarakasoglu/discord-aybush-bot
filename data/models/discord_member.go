package models

import "time"

type DiscordMember struct{
	Id int
	MemberId string
	GuildId string
	Email string
	Username string
	Discriminator string
	IsVerified bool
	IsBot bool
	Left bool
	JoinedAt time.Time
}