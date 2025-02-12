package data

import (
	"context"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type guildRepo struct {
	db *sqlx.DB
}

func (g *guildRepo) GetJoinedGuilds(ctx context.Context, userId int32) ([]models.Guild, error) {
	var output []models.Guild
	rows, err := g.db.QueryContext(
		ctx,
		`SELECT g.id, g.name, g.description, g.owner_id, g.created_date, g.updated_date
		 FROM guild g
		 JOIN member m on m.guild_id = g.id
		 WHERE m.user_id = $1;`,
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var guild models.Guild
		if err := rows.Scan(&guild.ID, &guild.Name, &guild.Description, &guild.OwnerId, &guild.CreatedDate, &guild.UpdatedDate); err != nil {
			return nil, err
		}
		output = append(output, guild)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, err
}

func (g *guildRepo) GetGuildChannels(ctx context.Context, guildId, userId int32) ([]models.Channel, error) {
	var output []models.Channel
	rows, err := g.db.QueryContext(
		ctx,
		`SELECT c.id, c.name, c.message_count, c.guild_id, c.created_date, c.updated_date
		 FROM channel c
		 JOIN member m on c.guild_id = m.guild_id
		 WHERE m.user_id = $1 AND c.guild_id = $2`,
		&userId,
		&guildId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var channel models.Channel
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.MessageCount, &channel.GuildId, &channel.CreatedDate, &channel.UpdatedDate); err != nil {
			return nil, err
		}
		output = append(output, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, err
}

func (g *guildRepo) GetGuildChannel(ctx context.Context, guildId, channelId, userId int32) (*models.Channel, error) {
	var output models.Channel
	err := g.db.QueryRowxContext(
		ctx,
		`SELECT c.id, c.name, c.message_count, c.guild_id, c.created_date, c.updated_date
		 FROM channel c
		 JOIN member m on c.guild_id = m.guild_id
		 WHERE c.id = $1 AND m.user_id = $2 AND c.guild_id = $3`,
		channelId,
		userId,
		guildId,
	).StructScan(&output)

	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (g *guildRepo) GetOwnedGuilds(ctx context.Context, userId int32) ([]models.Guild, error) {
	var output []models.Guild = make([]models.Guild, 0)
	rows, err := g.db.QueryContext(
		ctx,
		`SELECT id, name, description, owner_id, created_date, updated_date
         FROM guild
         WHERE owner_id = $1;`,
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var guild models.Guild
		if err := rows.Scan(&guild.ID, &guild.Name, &guild.Description, &guild.OwnerId, &guild.CreatedDate, &guild.UpdatedDate); err != nil {
			return nil, err
		}
		output = append(output, guild)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, err
}

func (g *guildRepo) GetGuildByID(ctx context.Context, guildId, userId int32) (*models.Guild, error) {
	var output models.Guild
	err := g.db.QueryRowxContext(
		ctx,
		`SELECT g.id, g.name, g.description, g.owner_id, g.created_date, g.updated_date
		 FROM guild g
		 JOIN member as m on m.guild_id = g.id AND m.user_id = $2
		 WHERE g.id = $1;`,
		guildId,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
