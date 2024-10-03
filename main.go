package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"blogr.moe/backend/auth"
	"blogr.moe/backend/database"
	"blogr.moe/backend/routes"
	"blogr.moe/backend/utils/scheduler"
	"github.com/joho/godotenv"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	err := godotenv.Load() // Load .env file from the current directory
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	secret := os.Getenv("SECRET")
	baseUrl := os.Getenv("BASE_URL")
	if secret == "" {
		log.Fatal("SECRET is not set")
	}
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Domain:   baseUrl,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${id} ${time_rfc3339} ${remote_ip} > ${method} > ${uri} > ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://dev.blogr.moe"},
		AllowCredentials: true, // Allow credentials (cookies)
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-CSRF-Token", "Authorization", "X-CSRF-Token"},
	}))
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "cookie:csrf",
		CookieDomain:   baseUrl,
		CookieName:     "csrf",
		CookieMaxAge:   86400,
		CookieSecure:   true,
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteStrictMode,
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
	}))

	accesslog, err := os.OpenFile("access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} - ${id} [${time_rfc3339}] \"${method} ${uri} HTTP/1.1\" ${status} ${bytes_sent}\n",
		Output: accesslog, // Set the Output to the log file
	}))
	routes.RegisterRoutes(e)
	c := e.NewContext(nil, nil)

	s24h := scheduler.NewScheduler()
	s24h.ScheduleTask(scheduler.Task{
		Action: func() {
			auth.IsPaymentActive(c)
		},
		Duration: 24 * time.Hour,
	})
	go s24h.Run()
	database.GetTotalPostCount()
	database.GetTotalUserCount()

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("PORT is not set")
	}
	//backups.BackupFiles()
	//mail.TestMail()
	//queue.NewQueueManager().ProcessAll()
	e.Use(session.Middleware(store))

	e.StartTLS(":"+strconv.Itoa(port), "backend/certificates/cert.pem", "backend/certificates/key.pem")
}
