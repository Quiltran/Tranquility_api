package main

import (
	"fmt"
	"log"
	"net/http"
	"tranquility/app"
	"tranquility/config"
	"tranquility/controllers"
	"tranquility/data"
	"tranquility/middleware"
	"tranquility/services"
)

func main() {
	logger, err := services.CreateLogger("Tranquility")
	if err != nil {
		log.Fatalln(err)
		panic(1)
	}

	config, err := config.NewConfig()
	if err != nil {
		logger.ERROR(fmt.Errorf("error creating config: %v", err).Error())
		panic(1)
	}

	fileHandler := services.NewFileHandler(config.UploadPath)

	database, err := data.CreatePostgres(config.ConnectionString, fileHandler)
	if err != nil {
		logger.ERROR(fmt.Errorf("error creating connection to database: %v", err).Error())
		panic(1)
	}

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

	mux := middleware.RequestLog(server, logger)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
		panic(1)
	}
}
