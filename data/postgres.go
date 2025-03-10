package data

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"tranquility/models"
	"tranquility/services"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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
	webAuthn         *webauthn.WebAuthn
	webAuthnSessions *services.WebAuthnSessions
}

func CreatePostgres(
	connectionString string,
	fileHandler *services.FileHandler,
	jwtHandler *services.JWTHandler,
	cloudflare *services.CloudflareService,
	pushNotification *services.PushNotificationService,
	webAuthn *webauthn.WebAuthn,
	webAuthnSessions *services.WebAuthnSessions,
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
		webAuthn:         webAuthn,
		webAuthnSessions: webAuthnSessions,
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

	if credentials.UserHandle == nil {
		if err := p.authRepo.UpdateLoginUserHandle(ctx, credentials.ID); err != nil {
			return nil, fmt.Errorf("an error occurred while updating user_handle while logging in: %v", err)
		}
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

	if validPassword := services.VerifyPasswordRequirements(user.Password); !validPassword {
		return nil, ErrInvalidPasswordFormat
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

func (p *Postgres) CreateMessage(ctx context.Context, message *models.Message, userId int32) (*models.Message, error) {
	tx, messageData, err := p.messageRepo.CreateMessage(ctx, message, userId)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while creating message: %s", err)
	}
	defer tx.Rollback()

	if messageData.AuthorAvatar != "" {
		url, err := p.fileHandler.GetFileUrl(messageData.AuthorAvatar)
		if err != nil {
			return nil, fmt.Errorf("error getting %s avatar after creating their message: %v", messageData.Author, err)
		}
		messageData.AuthorAvatar = url
	}

	if message.AttachmentIDs != nil {
		for _, attachment := range message.AttachmentIDs {
			if err := p.messageRepo.CreateAttachmentMapping(ctx, tx, messageData.ID, attachment); err != nil {
				return nil, fmt.Errorf("an error occurred while creating attachment mapping")
			}
		}
	}

	tx.Commit()

	attachments, err := p.messageRepo.GetMessageAttachment(ctx, messageData.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get message attachment while creating: %v", err)
	}
	for i := range attachments {
		url, err := p.fileHandler.GetFileUrl(attachments[i].FileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get url path for message attachment while submitting: %v", err)
		}
		messageData.Attachment = append(messageData.Attachment, url)
	}

	return messageData, nil
}

func (p *Postgres) GetChannelMessages(ctx context.Context, userId, guildId, channelId, pageNumber int32) ([]models.Message, error) {
	messages, err := p.messageRepo.GetChannelMessages(ctx, userId, guildId, channelId, pageNumber)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while getting channel messages: %v", err)
	}

	for i := range messages {
		if messages[i].AuthorAvatar != "" {
			url, err := p.fileHandler.GetFileUrl(messages[i].AuthorAvatar)
			if err != nil {
				return nil, fmt.Errorf("an error occurred collecting %s avatar while getting channel messages: %v", messages[i].Author, err)
			}
			messages[i].AuthorAvatar = url
		}
		attachments, err := p.messageRepo.GetMessageAttachment(ctx, messages[i].ID)
		if err != nil {
			return nil, fmt.Errorf("unable toget message attachment: %v", err)
		}
		for _, attachment := range attachments {
			url, err := p.fileHandler.GetFileUrl(attachment.FileName)
			if err != nil {
				return nil, fmt.Errorf("unable to get url path for message attachment: %v", err)
			}
			if messages[i].Attachment == nil {
				messages[i].Attachment = make([]string, 0)
			}
			messages[i].Attachment = append(messages[i].Attachment, url)
		}
	}

	return messages, nil
}

func (p *Postgres) RegisterUserWebAuthn(ctx context.Context, claims *models.Claims) (*protocol.CredentialCreation, error) {
	options, session, err := p.webAuthn.BeginRegistration(
		claims,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
	)
	if err != nil {
		return nil, err
	}
	sessionBytes, err := json.Marshal(&session)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while marshaling webauthn registration session: %v", err)
	}

	p.webAuthnSessions.AddSession(string(claims.ID), sessionBytes)
	return options, nil
}

func (p *Postgres) CompleteWebauthnRegister(ctx context.Context, claims *models.Claims, r *http.Request) error {
	sessionBytes, err := p.webAuthnSessions.GetSession(string(claims.ID))
	if err != nil {
		return fmt.Errorf("an error occurred while collecting webauthn session to complete registration: %v", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionBytes, &session); err != nil {
		return fmt.Errorf("an error occurred to unmarshal webauthn registration session in order to complete: %v", err)
	}

	credential, err := p.webAuthn.FinishRegistration(claims, session, r)
	if err != nil {
		return fmt.Errorf("an error occurred while finishing webauthn registration: %v", err)
	}

	if err := p.authRepo.saveWebAuthnCredential(ctx, credential, claims.ID); err != nil {
		return fmt.Errorf("an error occurred while saving webauthn credentials to the database after completing registration: %v", err)
	}

	return nil
}

func (p *Postgres) BeginWebAuthnLogin(ctx context.Context) (string, *protocol.CredentialAssertion, error) {
	sessionIdBytes, err := services.GenerateWebAuthnID()
	if err != nil {
		return "", nil, fmt.Errorf("an error occurred while generating the sessionID to be used for webauthn login request: %v", err)
	}
	sessionId := base64.StdEncoding.EncodeToString(sessionIdBytes)
	options, session, err := p.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		return "", nil, err
	}

	sessionBytes, err := json.Marshal(&session)
	if err != nil {
		return "", nil, fmt.Errorf("an error occurred while marshaling webauthn login session: %v", err)
	}
	p.webAuthnSessions.AddSession(sessionId, sessionBytes)

	return sessionId, options, nil
}

func (p *Postgres) CompleteWebAuthnLogin(ctx context.Context, sessionId string, r *http.Request) (*models.AuthUser, error) {
	sessionBytes, err := p.webAuthnSessions.GetSession(sessionId)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while getting webauthn login session to complete: %v", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionBytes, &session); err != nil {
		return nil, fmt.Errorf("an error occurred to unmarshal webauthn login session in order to complete: %v", err)
	}

	var userCredentials *models.AuthUser
	_, err = p.webAuthn.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (claims webauthn.User, err error) {
			user, claims, err := p.authRepo.getWebAuthnCredential(ctx, rawID, userHandle)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf("no credential was found while completing webauthn login")
				}
				return nil, fmt.Errorf("an error occurred while finding webauthn login user: %v", err)
			}
			userCredentials = user
			return claims, nil
		},
		session,
		r,
	)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while finishing webauthn discoverable login: %v", err)
	}

	authToken, err := p.jwtHandler.GenerateToken(userCredentials)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while generating token: %v", err)
	}
	userCredentials.Token = authToken
	userCredentials.ClearAuth()

	return userCredentials, nil
}

func (p *Postgres) GetUserProfile(ctx context.Context, userId int32) (*models.Profile, error) {
	profile, err := p.authRepo.GetUserProfile(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while getting user's profile information: %v", err)
	}

	if profile.AvatarURL != nil {
		url, err := p.fileHandler.GetFileUrl(*profile.AvatarURL)
		if err != nil {
			return profile, fmt.Errorf("an error occurred while getting the avatar url")
		}
		profile.AvatarURL = &url
	}

	return profile, nil
}

func (p *Postgres) UpdateUserProfile(ctx context.Context, profile *models.Profile, userId int32) (*models.Profile, error) {
	tx, err := p.authRepo.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while beginning tx before profile update for %d: %v", userId, err)
	}
	defer tx.Rollback()

	if err := p.authRepo.UpdateUserProfile(ctx, tx, profile, userId); err != nil {
		return nil, fmt.Errorf("an error occurred while updating %d profile: %v", userId, err)
	}

	if profile.AvatarID != nil {
		replacedFile, err := p.authRepo.CreateProfileMapping(ctx, tx, userId, *profile.AvatarID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("an invalid profile was attempted to be updated. no profile was found: %d", userId)
			} else if !errors.Is(err, ErrDuplicateProfileAttachment) {
				return nil, fmt.Errorf("an error occurred while updating %d user avatar: %v", userId, err)
			}
		}
		if replacedFile != nil {
			p.fileHandler.DeleteFile(*replacedFile)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("an error occurred while commiting %d profile update: %v", userId, err)
	}

	profile, err = p.authRepo.GetUserProfile(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("an error occurred getting %d profile after update: %v", userId, err)
	}
	if profile.AvatarURL != nil {
		url, err := p.fileHandler.GetFileUrl(*profile.AvatarURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get url of new profile picture after updating profile for %d: %v", userId, err)
		}
		profile.AvatarURL = &url
	}
	return profile, nil
}
