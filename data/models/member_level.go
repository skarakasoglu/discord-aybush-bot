package models

import "time"

type MemberLevel struct{
	Id int
	Member
	ExperiencePoints int64
	LastMessageTimestamp time.Time
}
