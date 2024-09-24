package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"encoding/gob"
	"encoding/json"

	"moe-blogger/database"
)

var oauthConf *oauth2.Config

var db = database.DB

func init() {
	gob.Register(User{})

	DiscordClientID := os.Getenv("DISCORD_CLIENT_ID")
	DiscordClientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	DiscordRedirectURI := os.Getenv("DISCORD_REDIRECT_URI")
	oauthConf = &oauth2.Config{
		ClientID:     DiscordClientID,
		ClientSecret: DiscordClientSecret,
		RedirectURL:  DiscordRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
		Scopes: []string{"identify"},
	}
	// Assuming database.DB is a *gorm.DB instance and properly initialized
	db.AutoMigrate(&User{})
	userid := 228343232520519680
	db = db.Exec("UPDATE users SET groups = ? WHERE id = ?", "admin", userid)
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Groups      string `json:"groups"`
	DateCreated string `json:"date_created"`
	DoesExist   bool   `json:"does_exist"`
}

type LoggedInUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	IsLoggedIn bool   `json:"is_logged_in"`
}

func CallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	client := oauthConf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	user := User{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	fmt.Println(user)
	db := database.DB
	if err != nil {
		log.Fatal(err)
	}
	err = db.Where("id = ?", user.ID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			db = db.Create(&user)
			if db.Error != nil {
				return c.JSON(http.StatusInternalServerError, db.Error)
			}
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	sess, err := session.Get("session", c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Failed to get session: %s", err.Error()))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values["user"] = user

	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Failed to save session: %s", err.Error()))
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
func LoginHandler(c echo.Context) error {
	url := oauthConf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func LogoutHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{MaxAge: -1}
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func AdminCheck(c echo.Context) bool {
	db := database.DB

	sess, err := session.Get("session", c)
	if err != nil {
		return false
	}

	userSessionValue, ok := sess.Values["user"]
	if !ok {
		return false
	}

	user, ok := userSessionValue.(User)
	if !ok {
		return false
	}

	var userFromDB User
	if err := db.Where("id = ?", user.ID).First(&userFromDB).Error; err != nil {
		return false
	}

	if !strings.Contains(userFromDB.Groups, "admin") {
		return false
	}

	return true
}
