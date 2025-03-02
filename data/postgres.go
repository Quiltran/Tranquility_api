package data

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"tranquility/models"
	"tranquility/services"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	ErrAttachmentNotFound = errors.New("attachment was not found while deleting")
)

type Postgres struct {
	authRepo
	attachmentRepo
	guildRepo
	messageRepo
	memberRepo
	notificationRepo
	fileHandler      *services.FileHandler
	jwtHandler       *services.JWTHandler
	cloudflare       *services.CloudflareService
	pushNotification *services.PushNotificationService
}

func CreatePostgres(
	connectionString string,
	fileHandler *services.FileHandler,
	jwtHandler *services.JWTHandler,
	cloudflare *services.CloudflareService,
	pushNotification *services.PushNotificationService,
) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		authRepo:         authRepo{db},
		attachmentRepo:   attachmentRepo{db},
		guildRepo:        guildRepo{db},
		messageRepo:      messageRepo{db},
		memberRepo:       memberRepo{db},
		notificationRepo: notificationRepo{db},
		fileHandler:      fileHandler,
		jwtHandler:       jwtHandler,
		cloudflare:       cloudflare,
		pushNotification: pushNotification,
	}, nil
}

func (p *Postgres) Login(ctx context.Context, user *models.AuthUser, ip string) (*models.AuthUser, error) {
	if user.Password == "" {
		return nil, ErrMissingPassword
	}

	if ok, err := p.cloudflare.VerifyTurnstile(user.Turnstile, ip); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("turnstile was rejected while logging in")
	}

	credentials, err := p.authRepo.Login(ctx, user)
	if err != nil {
		return nil, err
	}

	if ok, err := services.VerifyPassword(user.Password, credentials.Password); err != nil {
		return nil, fmt.Errorf("an error occurred while verifying password: %v", err)
	} else if !ok {
		return nil, ErrInvalidCredentials
	}

	authToken, err := p.jwtHandler.GenerateToken(credentials)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while generating token: %v", err)
	}
	credentials.Token = authToken
	credentials.ClearAuth()

	return credentials, nil
}

func (p *Postgres) Register(ctx context.Context, user *models.AuthUser, ip string) (*models.AuthUser, error) {
	if user.Password == "" || user.ConfirmPassword == "" {
		return nil, ErrInvalidCredentials
	}

	if ok, err := p.cloudflare.VerifyTurnstile(user.Turnstile, ip); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("turnstile was rejected")
	}

	password, err := services.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("an error occurred hashing password while registering user: %v", err)
	}

	user.Password = password
	output, err := p.authRepo.Register(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while registering user: %v", err)
	}

	return output, nil
}

func (p *Postgres) RefreshToken(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	if user.ID == 0 || user.RefreshToken == "" {
		return nil, ErrInvalidCredentials
	}

	credentials, err := p.authRepo.RefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := p.jwtHandler.GenerateToken(credentials)
	if err != nil {
		return nil, err
	}
	credentials.Token = token
	return credentials, nil
}

func (p *Postgres) CreateAttachment(ctx context.Context, file *multipart.File, attachment *models.Attachment) (*models.Attachment, error) {
	outputName, outputPath, err := p.fileHandler.StoreFile(file, attachment.FileName)
	if err != nil {
		return nil, err
	}

	attachment.FileName = outputName
	attachment.FilePath = outputPath

	output, err := p.attachmentRepo.CreateAttachment(ctx, attachment)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (p *Postgres) DeleteAttachment(ctx context.Context, fileId, userId int32) error {
	transaction, fileName, err := p.attachmentRepo.DeleteAttachment(ctx, fileId, userId)
	if err != nil {
		return err
	}
	if fileName == "" {
		return ErrAttachmentNotFound
	}

	err = p.fileHandler.DeleteFile(fileName)
	if err != nil {
		return err
	}

	transaction.Commit()
	return nil
}

func (p *Postgres) GetJoinedGuilds(ctx context.Context, userId int32) ([]models.Guild, error) {
	guilds, err := p.guildRepo.GetJoinedGuilds(ctx, userId)
	if err != nil {
		return nil, err
	}

	for i := range guilds {
		channels, err := p.guildRepo.GetGuildChannels(ctx, guilds[i].ID, userId)
		if err != nil {
			return nil, err
		}
		members, err := p.GetGuildMembers(ctx, guilds[i].ID, userId)
		if err != nil {
			return nil, err
		}

		guilds[i].Channels = channels
		guilds[i].Members = members
	}

	return guilds, nil
}

func (p *Postgres) CreateGuild(ctx context.Context, guild *models.Guild, userId int32) (*models.Guild, error) {
	tx, guild, err := p.guildRepo.CreateGuild(ctx, guild, userId)
	if err != nil {
		return nil, err
	}

	if err = p.addGuildMember(ctx, guild.ID, userId, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("rollback error: %v, original error: %v", rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return guild, nil
}

func (p *Postgres) GetMembers(ctx context.Context, guildId int32) ([]models.Member, error) {
	if guildId == -1 {
		return p.memberRepo.GetMembers(ctx)
	} else {
		return p.memberRepo.GetNotAddedMembers(ctx, guildId)
	}
}

func (p *Postgres) SaveUserPushInformation(ctx context.Context, registration *webpush.Subscription, userId int32) error {
	myReg := &models.PushNotificationRegistration{
		Endpoint: registration.Endpoint,
		Keys: struct {
			P256dh string "json:\"p256dh\""
			Auth   string "json:\"auth\""
		}{
			P256dh: registration.Keys.P256dh,
			Auth:   registration.Keys.Auth,
		},
	}
	tx, err := p.notificationRepo.SaveUserPushInformation(ctx, myReg, userId)
	if err != nil {
		return fmt.Errorf("an error occurred while saving user push notification information: %v", err)
	}
	defer tx.Rollback()

	message := models.NewPushNotificationMessage("Notifications Registered", "You have been registered for push notifications.", "/", nil).
		WithAction("Open", "open").
		WithAction("Close", "close")
	if err := p.pushNotification.SimplePush(registration, message); err != nil {
		return fmt.Errorf("an error occurred while sending registered push notification: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while commiting user's push notification info: %v", err)
	}
	return nil
}
