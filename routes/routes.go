package routes

import (
	"net/http"

	auth "blogr.moe/auth"
	"blogr.moe/blog"
	"blogr.moe/home"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	// Home
	e.GET("/", home.Home)
	// Static
	e.Static("/static", "static")
	e.Static("/assets", "assets")

	// Login

	e.POST("/login", func(c echo.Context) error {
		return auth.LoginHandler(c)
	})

	e.GET("/logout", func(c echo.Context) error {
		return auth.LogoutHandler(c)
	})
	e.POST("/register", func(c echo.Context) error {
		return auth.RegisterUser(c)
	})

	// Profile
	e.GET("/user/profile", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.Redirect(http.StatusFound, "/")
		}

		return home.Profile(c)
	})
	e.GET("/user/posts", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.Redirect(http.StatusFound, "/")
		}
		return home.ProfilePosts(c)
	})
	e.GET("/user/comments", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.Redirect(http.StatusFound, "/")
		}
		return home.ProfileComments(c)
	})
	e.GET("/user/settings", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.Redirect(http.StatusFound, "/")
		}
		return home.ProfileSettings(c)
	})

	// Blog
	e.GET("/blog/:id", func(c echo.Context) error {
		return blog.GetBlog(c)
	})
	e.GET("/:user/blog", home.UserBlog)
	e.GET("/blog", home.BlogList)
	e.GET("/uploads/blogs/:uuid/:file", func(c echo.Context) error {
		return blog.GetBlogFile(c)
	})

	// API
	e.GET("/api/user/blog", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.JSON(http.StatusUnauthorized, "Unauthorized")
		}
		return blog.GetUserBlogPosts(c)
	})

	e.POST("/api/user/blog", func(c echo.Context) error {
		isUser := auth.UserCheck(c)
		if !isUser {
			return c.JSON(http.StatusUnauthorized, "Unauthorized")
		}
		return blog.NewBlogHandler(c)
	})

}
