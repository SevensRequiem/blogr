package routes

import (
	"moe-blogger/auth"
	"moe-blogger/home"

	"github.com/labstack/echo/v4"
)

func Routes(e *echo.Echo) {
	// Home
	e.GET("/", home.Home)

	// Auth
	e.GET("/login", auth.LoginHandler)
	e.GET("/auth/callback", auth.CallbackHandler)
	e.GET("/logout", auth.LogoutHandler)

	// Blog
	e.GET("/blog", home.BlogHandler)
	e.GET("/blog/:id", home.BlogPostHandler)

	// Static
	e.Static("/static", "static")
	e.Static("/assets", "assets")
}
