package controllers

import "tranquility/app"

type Controller interface {
	RegisterRoutes(*app.App)
}
