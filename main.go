package main

import (

	// echo
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// self packages
	"moe-blogger/routes"
)

func main() {
	// echo instance
	e := echo.New()

	//session store
	e.Use(middleware.Secure())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("3ASIRFGHSRIFGaerwsgkhwerngi456E00R670IA0NR76G0I078K0WR768G0A56R0K580I680G"))))

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//csrf
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
	}))

	// routes
	routes.Routes(e)

	// start server
	e.Logger.Fatal(e.Start(":1323"))
}
