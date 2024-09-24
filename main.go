package main

import (

	// echo
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// self packages
	"blogr.moe/logs"
	"blogr.moe/routes"
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
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	e.Use(session.Middleware(store))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${id} ${time_rfc3339} ${remote_ip} > ${method} > ${uri} > ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("11M"))
	e.Use(middleware.RequestID())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://blogr.moe",
		},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'; img-src 'self' data: https://img.shields.io https://placehold.co https://static.cloudflareinsights.com https://achan.moe; connect-src 'self'; font-src 'self'; frame-src 'self'; object-src 'none'; media-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'self';",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "same-origin",
	}))
	baseUrl := os.Getenv("BASE_URL")
	middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "cookie:_csrf",
		CookieName:     "_csrf",
		CookieMaxAge:   86400,
		CookiePath:     "/",
		CookieDomain:   baseUrl,
		CookieSecure:   true,
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteStrictMode,
	})

	accesslog, err := os.OpenFile("access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		logs.Fatal("Failed to open access.log")
	}
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} - ${id} [${time_rfc3339}] \"${method} ${uri} HTTP/1.1\" ${status} ${bytes_sent}\n",
		Output: accesslog, // Set the Output to the log file
	}))

	// routes
	routes.RegisterRoutes(e)

	// port
	PORT := os.Getenv("PORT")
	// start server
	e.Logger.Fatal(e.Start(":" + PORT))
}
