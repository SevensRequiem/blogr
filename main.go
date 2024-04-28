package moeblogger

import (

	// echo
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// self packages
	"moeblogger/routes"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routes.Routes(e)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
