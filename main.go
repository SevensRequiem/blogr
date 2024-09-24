package main

import (

	// echo
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// self packages
	"moe-blogger/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// echo instance
	e := echo.New()
	secret := os.Getenv("SECRET")
	//session store
	e.Use(middleware.Secure())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(secret))))

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//csrf
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
	}))

	// routes
	routes.Routes(e)

	// port
	PORT := os.Getenv("PORT")
	// start server
	e.Logger.Fatal(e.Start(":" + PORT))
}
