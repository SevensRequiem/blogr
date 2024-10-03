package routes

import (
	"net/http"

	auth "blogr.moe/backend/auth"
	"blogr.moe/backend/blog"
	"blogr.moe/backend/database"
	"blogr.moe/backend/home"
	"blogr.moe/backend/stripe"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return home.Home(c)
	})

	e.GET("/login", func(c echo.Context) error {
		return home.Login(c)
	})

	e.GET("/logout", func(c echo.Context) error {
		return auth.Logout(c)
	})

	e.GET("/register", func(c echo.Context) error {
		return home.Register(c)
	})

	e.GET("/dashboard", func(c echo.Context) error {
		if !auth.IsLoggedIn(c) {
			return c.Redirect(302, "/login")
		}
		return home.Dashboard(c)
	})

	e.GET("/premium", func(c echo.Context) error {
		return home.Premium(c)
	})

	// Static files
	e.Static("/assets", "assets")

	// api routes
	e.POST("/api/auth/login", auth.Login)
	e.POST("/api/auth/register", func(c echo.Context) error {
		return auth.Register(c)
	})
	e.GET("/api/auth/logout", auth.Logout)

	e.POST("/api/user/post", blog.NewBlogHandler)
	e.GET("/api/user/posts", blog.GetLatestPostsUser)

	e.GET("/u/:user/:id", func(c echo.Context) error {
		user := c.Param("user")
		id := c.Param("id")
		return home.SinglePost(c, user, id)
	})
	e.GET("/i/:user/:postid", func(c echo.Context) error {
		return blog.GetPostImage(c)
	})
	e.POST("/api/blog", blog.NewBlogHandler)
	e.DELETE("/api/user/blog/:id", blog.DeleteUserPost)
	e.GET("/api/stats", func(c echo.Context) error {
		stats, err := database.GetStats()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, stats)
	})
	e.GET("/api/stripe/checkout", stripe.GetCheckoutSession)
	e.GET("/api/stripe/success", stripe.CheckoutSuccessHandler)

}
