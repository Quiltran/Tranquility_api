package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserLacksPermission = errors.New("the user doesn't have valid permissions")
	ErrDuplicateMember     = errors.New("the user is already a part of the guild")
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

func (g *guildRepo) CreateGuild(ctx context.Context, guild *models.Guild, userId int32) (*sqlx.Tx, *models.Guild, error) {
	var output models.Guild
	tx, err := g.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, nil, err
	}
	err = tx.QueryRowxContext(
		ctx,
		`INSERT INTO guild (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_date, updated_date;`,
		guild.Name,
		guild.Description,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, nil, err
	}

	return tx, &output, nil
}

func (g *guildRepo) AddGuildMember(ctx context.Context, guildId, userId int32, tx *sqlx.Tx) error {
	query := `INSERT INTO member (user_id, guild_id, user_who_added) VALUES ($1, $2, $3) RETURNING id, user_id, guild_id, user_who_added, created_date, updated_date;`
	var rows sql.Result
	var err error

	if tx != nil {
		rows, err = tx.ExecContext(ctx, query, userId, guildId, userId)
	} else {
		rows, err = g.db.ExecContext(ctx, query, userId, guildId, userId)
	}

	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("an invalid number rows were affected while inserting member")
	}

	return nil
}

func (g *guildRepo) CreateChannel(ctx context.Context, channel *models.Channel, userId int32) (*models.Channel, error) {
	var output models.Channel
	err := g.db.QueryRowxContext(
		ctx,
		`INSERT INTO channel (name, guild_id) SELECT $1, $2
		 WHERE EXISTS (SELECT 1 FROM member WHERE guild_id = $2 AND user_id = $3)
         RETURNING id, name, message_count, guild_id, created_date, updated_date;`,
		channel.Name,
		channel.GuildId,
		userId,
	).StructScan(&output)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserLacksPermission
		}
		return nil, err
	}

	return &output, nil
}

func (g *guildRepo) CreateMember(ctx context.Context, member *models.Member) (*models.Member, error) {
	var output models.Member
	err := g.db.QueryRowxContext(
		ctx,
		`INSERT INTO member (user_id, guild_id, user_who_added)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, guild_id, user_who_added, created_date, updated_date;`,
		&member.UserId,
		&member.GuildId,
		&member.UserWhoAdded,
	).StructScan(&output)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, ErrDuplicateMember
			}
		}
		return nil, err
	}

	return &output, nil
}
