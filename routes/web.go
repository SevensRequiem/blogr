package routes

import (
	"github.com/labstack/echo/v4"
)

func Routes(e *echo.Echo) {
	// Home
	e.GET("/", home.Home)
	e.GET("/home", home.Home)

	// Auth
	e.GET("/login", home.Login)
	e.POST("/login", auth.Login)
	e.GET("/logout", auth.Logout)

	// Blog
	e.GET("/blog", blog.Render)
	e.GET("/blog/:id", blog.RenderPost)

	// RSS
	e.GET("/rss", rss.RSS)

	// Static
	e.Static("/static", "static")
	e.Static("/assets", "assets")
}
