package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"tranquility/app"
	"tranquility/config"
	"tranquility/controllers"
	"tranquility/data"
	"tranquility/middleware"
	"tranquility/services"

	"github.com/rs/cors"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	logger, err := services.CreateLogger("Tranquility")
	if err != nil {
		panic(err)
	}

	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	webAuthn, err := services.NewWebauthn(config.WebAuthnConfig)
	if err != nil {
		panic("unable to create webAuthn object")
	}
	webAuthnSessions := services.NewWebAuthnSessions()

	go func() {

		timer := time.NewTicker(time.Minute)

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				clearedSessionCount := webAuthnSessions.ClearExpiredSessions()
				if clearedSessionCount > 0 {
					logger.INFO(fmt.Sprintf("%d WebAuthn sessions have been cleared", clearedSessionCount))
				}
			}
		}
	}()

	fileHandler := services.NewFileHandler(config.UploadPath)
	jwtHandler := services.NewJWTHandler(config.JWTConfig)
	cloudflare := services.NewCloudflareService(config.TurnstileSecret, logger)
	pushNotification := services.NewPushNotificationService(config.PushNotificationConfig, logger)

	database, err := data.CreatePostgres(
		config.ConnectionString,
		fileHandler,
		jwtHandler,
		cloudflare,
		pushNotification,
		webAuthn,
		webAuthnSessions,
	)
	if err != nil {
		panic(err)
	}

	websocketServer := services.NewWebsocketServer(ctx, logger)
	go websocketServer.Run()

	server := app.CreateApp(logger, jwtHandler)

	controllers.NewAuthController(
		logger,
		database,
	).RegisterRoutes(&server)
	controllers.NewAttachmentController(
		logger,
		fileHandler,
		database,
	).RegisterRoutes(&server)
	controllers.NewGuildController(
		logger,
		database,
	).RegisterRoutes(&server)
	controllers.NewMessageController(
		logger,
		database,
	).RegisterRoutes(&server)
	controllers.NewWebsocketController(
		database,
		logger,
		websocketServer,
		config.JWTConfig.Audience,
		pushNotification,
	).RegisterRoutes(&server)
	controllers.NewMemberController(
		logger,
		database,
	).RegisterRoutes(&server)
	controllers.NewPushNotificationController(
		logger,
		database,
	).RegisterRoutes(&server)
	controllers.NewProfileController(
		logger,
		database,
	).RegisterRoutes(&server)

	mux := middleware.RequestLog(server, logger)
	c := cors.New(cors.Options{
		AllowedOrigins: config.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "DELETE", "HEAD"},
		AllowedHeaders: []string{"*"},
	})

	logger.INFO(fmt.Sprintf("allowing origins %s", config.AllowedOrigins))

	if err := http.ListenAndServe(":8080", c.Handler(mux)); err != nil {
		panic(err)
	}
}
