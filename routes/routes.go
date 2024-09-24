package routes

import (
	"blogr.moe/home"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	// Home
	e.GET("/", home.Home)
	// Static
	e.Static("/static", "static")
	e.Static("/assets", "assets")
}
