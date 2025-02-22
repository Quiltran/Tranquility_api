package main

import (
	"context"
	"net/http"
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

	fileHandler := services.NewFileHandler(config.UploadPath)

	database, err := data.CreatePostgres(config.ConnectionString, fileHandler)
	if err != nil {
		panic(err)
	}

	websocketServer := services.NewWebsocketServer(ctx, logger)
	go websocketServer.Run()

	server := app.CreateApp(logger)

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
	).RegisterRoutes(&server)

	mux := middleware.RequestLog(server, logger)
	c := cors.AllowAll()

	if err := http.ListenAndServe(":8080", c.Handler(mux)); err != nil {
		panic(err)
	}
}
