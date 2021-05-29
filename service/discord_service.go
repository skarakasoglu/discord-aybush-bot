package service

import (
	"database/sql"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
)

type DiscordService struct {
	db *sql.DB
}

func (d DiscordService) InsertDiscordAttachment(attachment models.DiscordAttachment) (int, error) {
	panic("implement me")
}

func (d DiscordService) GetDiscordAttachmentById(attachmentId string) (models.DiscordAttachment, error) {
	panic("implement me")
}

func (d DiscordService) DeleteDiscordAttachmentById(attachmentId string) (bool error) {
	panic("implement me")
}

func (d DiscordService) InsertDiscordLevel(level models.DiscordLevel) (int, error) {
	panic("implement me")
}

func (d DiscordService) UpdateDiscordLevel(level models.DiscordLevel) (bool, error) {
	panic("implement me")
}

func (d DiscordService) GetAllDiscordLevels() ([]models.DiscordLevel, error) {
	query := `SELECT dl.id "level_id", dl.required_experience_points, dl.maximum_experience_points, dr.id "d_role_id", dr.role_id, dr.name 
				FROM "discord_levels" as dl left join "discord_roles" as dr on dl.role_id = dr.role_id order by dl.id;`

	rows, err := d.db.Query(query)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var levels []models.DiscordLevel

	for rows.Next() {
		var level models.DiscordLevel
		var role models.DiscordRole

		_ = rows.Scan(&level.Id, &level.RequiredExperiencePoints, &level.MaximumExperiencePoints, &role.Id, &role.RoleId, &role.Name)
		level.DiscordRole = role

		levels = append(levels, level)
	}

	return levels, nil
}

func (d DiscordService) DeleteDiscordLevel(level int) (bool error) {
	panic("implement me")
}

func (d DiscordService) InsertDiscordMember(member models.DiscordMember) (int, error) {
	query := `INSERT INTO "discord_members"("member_id","email","username","discriminator","avatar_url","is_verified","is_bot","joined_at","is_left","guild_id") 
				VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
				ON CONFLICT(member_id) DO UPDATE SET 
				    username = excluded.username, discriminator = excluded.discriminator, avatar_url = excluded.avatar_url, is_verified = excluded.is_verified,
				    is_bot = excluded.is_bot, joined_at = excluded.joined_at, guild_id = excluded.guild_id
				RETURNING id;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing statement: %v", err)
		return -1, err
	}

	lastInsertedId := 0
	err = preparedStmt.QueryRow(member.MemberId, member.Email, member.Username, member.Discriminator, member.AvatarUrl, member.IsVerified, member.IsBot, member.JoinedAt, member.Left, member.GuildId).Scan(&lastInsertedId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return lastInsertedId, err
	}

	return lastInsertedId, nil
}

func (d DiscordService) UpdateDiscordMemberById(member models.DiscordMember) (bool, error) {
	query := `UPDATE "discord_members" SET username=$1,discriminator=$2,avatar_url=$3,is_verified=$4,is_bot=$5,joined_at=$6,is_left=$7,guild_id=$8 where member_id=$9;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStmt.Query(member.Username, member.Discriminator, member.AvatarUrl, member.IsVerified, member.IsBot, member.JoinedAt, member.Left, member.GuildId,member.MemberId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) GetDiscordMemberById(memberId string) (models.DiscordMember, error) {
	var member models.DiscordMember

	query := `SELECT id, member_id, email, username, discriminator, is_verified, 
       is_bot, joined_at, is_left, guild_id, avatar_url FROM "discord_members" where member_id = $1;`

	row := d.db.QueryRow(query, memberId)

	var email sql.NullString
	err := row.Scan(&member.Id, &member.MemberId, &email, &member.Username, &member.Discriminator, &member.IsVerified, &member.IsBot, &member.JoinedAt, &member.Left, &member.GuildId, &member.AvatarUrl)
	if err != nil {
		log.Printf("[DiscordService] Error on scanning row: %v", err)
		return member, err
	}
	member.Email = email.String

	return member, nil
}

func (d DiscordService) GetAllDiscordMembers() ([]models.DiscordMember, error) {
	panic("implement me")
}

func (d DiscordService) DeleteDiscordMemberById(memberId string) (bool, error) {
	panic("implement me")
}

func (d DiscordService) InsertDiscordMemberLevel(level models.DiscordMemberLevel) (int, error) {
	query := `INSERT INTO "discord_member_levels" (member_id, experience_points, last_message_timestamp, message_count, active_voice_minutes) VALUES($1, $2, $3, $4, $5) RETURNING id;`

	preparedStatement, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return -1, err
	}

	lastInsertedId := -1
	err = preparedStatement.QueryRow(level.MemberId, level.ExperiencePoints, level.LastMessageTimestamp, level.MessageCount, level.ActiveVoiceMinutes).Scan(&lastInsertedId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return -1, err
	}

	return lastInsertedId, nil
}

func (d DiscordService) UpdateDiscordMemberLevelById(level models.DiscordMemberLevel) (bool, error) {
	query := `UPDATE "discord_member_levels" SET experience_points = $1, last_message_timestamp = $2, message_count = $3, active_voice_minutes = $4 where member_id = $5;`

	preparedStatement, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStatement.Exec(level.ExperiencePoints, level.LastMessageTimestamp, level.MessageCount, level.ActiveVoiceMinutes, level.MemberId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) GetAllDiscordMemberLevels() ([]models.DiscordMemberLevel, error) {
	query := `
			SELECT dml.id, dm.username, dm.discriminator, dml.experience_points, dml.last_message_timestamp,dml.message_count, dml.active_voice_minutes, 
		   dm.member_id, dm.guild_id, cdl.id "current_level",  cdl.required_experience_points "current_level_required", 
		   cdl.maximum_experience_points "current_level_maximum", 
			cdr.id "current_role_id", cdr.role_id "current_role_role_id", cdr.name "current_role_name",
			ndl.id "next_level", ndl.required_experience_points "next_level_required", ndl.maximum_experience_points "next_level_maximum",
			ndr.id "next_role_id", ndr.role_id "next_role_role_id", ndr.name "next_role_name"
			FROM "discord_member_levels" as dml 
			inner join "discord_members" as dm on dm.member_id = dml.member_id
			inner join "discord_levels" as cdl on dml.experience_points between cdl.required_experience_points and (cdl.maximum_experience_points - 1)
			inner join "discord_levels" as ndl on cdl.maximum_experience_points = ndl.required_experience_points
			inner join "discord_roles" as cdr on cdr.role_id = cdl.role_id
			inner join "discord_roles" as ndr on ndr.role_id = ndl.role_id
			WHERE is_left = false
			ORDER BY dml.experience_points DESC;
	`

	rows, err := d.db.Query(query)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var memberLevels []models.DiscordMemberLevel

	for rows.Next() {
		var memberLevel models.DiscordMemberLevel
		var member models.DiscordMember
		var currentLevel models.DiscordLevel
		var nextLevel models.DiscordLevel

		err = rows.Scan(&memberLevel.Id, &member.Username, &member.Discriminator, &memberLevel.ExperiencePoints,
			&memberLevel.LastMessageTimestamp, &memberLevel.MessageCount, &memberLevel.ActiveVoiceMinutes, &member.MemberId,
			&member.GuildId, &currentLevel.Id, &currentLevel.RequiredExperiencePoints, &currentLevel.MaximumExperiencePoints,
			&currentLevel.DiscordRole.Id, &currentLevel.DiscordRole.RoleId, &currentLevel.DiscordRole.Name,
			&nextLevel.Id, &nextLevel.RequiredExperiencePoints, &nextLevel.MaximumExperiencePoints,
			&nextLevel.DiscordRole.Id, &nextLevel.DiscordRole.RoleId, &nextLevel.DiscordRole.Name)

		if err != nil {
			log.Printf("[DiscordService] Error on scanning the row: %v", err)
		}

		memberLevel.DiscordMember = member
		memberLevel.CurrentLevel = currentLevel
		memberLevel.NextLevel = nextLevel

		memberLevels = append(memberLevels, memberLevel)
	}

	return memberLevels, nil
}

func (d DiscordService) DeleteDiscordMemberLevelById(memberLevelId int) (bool, error) {
	panic("implement me")
}

func (d DiscordService) InsertDiscordMemberMessage(message models.DiscordMemberMessage) (int, error) {
	panic("implement me")
}

func (d DiscordService) DeleteDiscordMemberMessage(message models.DiscordMemberMessage) (bool, error) {
	panic("implement me")
}

func (d DiscordService) GetDiscordMemberMessagesByMemberId(memberId string) ([]models.DiscordMemberMessage, error) {
	panic("implement me")
}

func (d DiscordService) InsertDiscordRole(role models.DiscordRole) (int, error) {
	query := `INSERT INTO "discord_roles"("role_id","name") VALUES($1,$2) RETURNING id;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return -1, err
	}

	lastInsertedId := 0
	err = preparedStmt.QueryRow(role.RoleId, role.Name).Scan(&lastInsertedId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return lastInsertedId, err
	}

	return lastInsertedId, nil
}

func (d DiscordService) UpdateDiscordRoleById(role models.DiscordRole) (bool, error) {
	query := `UPDATE "discord_roles" SET name=$1 where role_id=$2;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStmt.Query(role.Name, role.RoleId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) GetAllDiscordRoles() ([]models.DiscordRole, error) {
	panic("implement me")
}

func (d DiscordService) DeleteDiscordRoleById(roleId string) (bool, error) {
	query := `DELETE FROM "discord_roles" where role_id=$1;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStmt.Query(roleId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) InsertDiscordTextChannel(channel models.DiscordTextChannel) (int, error) {
	query := `INSERT INTO "discord_text_channels"("channel_id","name", "is_nsfw", "created_at") VALUES($1,$2,$3,$4) RETURNING id;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return -1, err
	}

	lastInsertedId := 0
	err = preparedStmt.QueryRow(channel.ChannelId, channel.Name, channel.IsNsfw, channel.CreatedAt).Scan(&lastInsertedId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return lastInsertedId, err
	}

	return lastInsertedId, nil
}

func (d DiscordService) UpdateDiscordTextChannelById(channel models.DiscordTextChannel) (bool, error) {
	query := `UPDATE "discord_text_channels" SET name=$1,is_nsfw=$2 where channel_id=$3;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStmt.Query(channel.Name, channel.IsNsfw, channel.ChannelId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) GetAllDiscordTextChannels() ([]models.DiscordTextChannel, error) {
	panic("implement me")
}

func (d DiscordService) DeleteDiscordTextChannelById(channelId string) (bool, error) {
	query := `DELETE FROM "discord_text_channels" where channel_id=$1;`

	preparedStmt, err := d.db.Prepare(query)
	if err != nil {
		log.Printf("[DiscordService] Error on preparing the statement: %v", err)
		return false, err
	}

	_, err = preparedStmt.Query(channelId)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the prepared statement: %v", err)
		return false, err
	}

	return true, nil
}

func (d DiscordService) InsertDiscordLevelUpMessage(message models.DiscordLevelUpMessage) (int, error) {
	panic("implement me")
}

func (d DiscordService) UpdateDiscordLevelUpMessage(message models.DiscordLevelUpMessage) (bool, error) {
	panic("implement me")
}

func (d DiscordService) GetAllDiscordLevelUpMessages() ([]models.DiscordLevelUpMessage, error) {
	var messages []models.DiscordLevelUpMessage

	query := `SELECT * FROM "discord_level_up_messages";`

	rows, err := d.db.Query(query)
	if err != nil {
		log.Printf("[DiscordService] Error on executing the query: %v", err)
		return messages, err
	}
	defer rows.Close()

	for rows.Next() {
		var message models.DiscordLevelUpMessage
		rows.Scan(&message.Id, &message.Content)

		messages = append(messages, message)
	}

	return messages, nil
}

func (d DiscordService) DeleteDiscordLevelUpMessageById(id int) (bool, error) {
	panic("implement me")
}

func NewDiscordService(db *sql.DB) *DiscordService{
	return &DiscordService{db: db}
}
