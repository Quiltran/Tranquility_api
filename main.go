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

	services.LoadJWTSettings()

	config, err := config.NewConfig()
	if err != nil {
		logger.ERROR(fmt.Errorf("error creating config: %v", err).Error())
		panic(1)
	}

	database, err := data.CreatePostgres(config.ConnectionString)
	if err != nil {
		logger.ERROR(fmt.Errorf("error creating connection to database: %v", err).Error())
		panic(1)
	}
	dbCommands := services.NewDatabaseCommands(database)

	server := app.CreateApp(database, logger)

	controllers.NewAuthController(
		logger,
		dbCommands,
		config,
	).RegisterRoutes(&server)

	mux := middleware.RequestLog(server, logger)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
		panic(1)
	}
}
