package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"blogr.moe/blog"
	"blogr.moe/database"
	"blogr.moe/logs"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var db = database.DB

type User struct {
	ID            uint   `json:"id" gorm:"primary_key"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	LastLogin     string `json:"last_login"`
	DateCreated   string `json:"date_created"`
	Reputation    int    `json:"reputation"`
	TotalViews    int    `json:"total_views"`
	Group         groups
	Blog          []blog.Blog
	Premium       bool   `json:"premium"`
	TransactionID string `json:"transaction_id"`
	SubDomain     string `json:"sub_domain"`
}

type groups struct {
	ID   uint   `json:"id" gorm:"primary_key"`
	Name string `json:"name"`
}

type UserLogin struct {
	ID       uint   `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) GetUser(username string) error {
	if err := db.Where("username = ?", username).First(&u).Error; err != nil {
		return err
	}
	return nil
}

func GetUserByUsername(username string) User {
	var user User
	db.Where("username = ?", username).First(&user)
	user = User{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Username:    user.Username,
		Avatar:      user.Avatar,
		LastLogin:   user.LastLogin,
		DateCreated: user.DateCreated,
		Reputation:  user.Reputation,
		TotalViews:  user.TotalViews,
		Group:       user.Group,
		Blog:        user.Blog,
		Premium:     user.Premium,
	}
	return user
}

func GetUser(uuid string) User {
	var user User
	db.Where("uuid = ?", uuid).First(&user)
	user = User{
		ID:          user.ID,
		UUID:        user.UUID,
		Name:        user.Name,
		Email:       user.Email,
		Username:    user.Username,
		Avatar:      user.Avatar,
		LastLogin:   user.LastLogin,
		DateCreated: user.DateCreated,
		Reputation:  user.Reputation,
		TotalViews:  user.TotalViews,
		Group:       user.Group,
		Blog:        user.Blog,
		Premium:     user.Premium,
	}
	return user
}
func checkPassword(password string, user UserLogin) bool {
	db.Where("username = ?", user.Username).First(&user)
	decrypted := DecodePassword(user.Password)
	return password == decrypted
}

func encryptPassword(password string) string {
	enc := os.Getenv("ENCRYPT_KEY")
	h := hmac.New(sha256.New, []byte(enc))
	h.Write([]byte(password))
	encpass := h.Sum(nil)
	encodedPass := base64.StdEncoding.EncodeToString(encpass)
	logs.Debug("Encoded Pass for new user")
	return encodedPass
}

func DecodePassword(encrypted_password string) string {
	enc := os.Getenv("ENCRYPT_KEY")
	decoded, err := base64.StdEncoding.DecodeString(encrypted_password)

	if err != nil {
		log.Println(err)
		return ""
	}
	h := hmac.New(sha256.New, []byte(enc))
	h.Write(decoded)
	encpass := h.Sum(nil)
	logs.Debug("Decoded Password for user")
	return base64.StdEncoding.EncodeToString(encpass)
}

func getFullUser(uuid string) User {
	var user User
	db.Where("uuid = ?", uuid).First(&user)
	user = User{
		ID:          user.ID,
		UUID:        user.UUID,
		Name:        user.Name,
		Email:       user.Email,
		Username:    user.Username,
		Avatar:      user.Avatar,
		LastLogin:   user.LastLogin,
		DateCreated: user.DateCreated,
		Reputation:  user.Reputation,
		TotalViews:  user.TotalViews,
		Group:       user.Group,
		Blog:        user.Blog,
		Premium:     user.Premium,
	}
	return user
}

func (u *UserLogin) LoginHandler(username string, password string, c echo.Context, uf *User) bool {
	if err := db.Where("username = ?", username).First(&u).Error; err != nil {
		return false
	}
	if checkPassword(password, *u) {
		sess, _ := session.Get("session", c)
		sess.Values["user"] = getFullUser(uf.UUID)
		sess.Save(c.Request(), c.Response())
		return true
	}
	return false
}

func (u *User) CreateUser(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	username := c.FormValue("username")
	password := c.FormValue("password")
	uuid := genuuid()
	encpass := encryptPassword(password)
	user := User{
		UUID:        uuid,
		Name:        name,
		Email:       email,
		Username:    username,
		DateCreated: time.Now().String(),
		Reputation:  0,
		TotalViews:  0,
		Group: groups{
			Name: "User",
		},
		Premium: false,
	}
	userlogin := UserLogin{
		UUID:     uuid,
		Username: username,
		Password: encpass,
	}
	db.Create(&user)
	db.Create(&userlogin)
	return c.JSON(http.StatusOK, user)
}

func genuuid() string {
	uuid := uuid.New()
	return uuid.String()
}

func AdminCheck(c echo.Context) bool {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	if user.Group.Name == "Admin" {
		return true
	}
	return false
}

func (u *User) UpdateUser(c echo.Context) error {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	name := c.FormValue("name")
	email := c.FormValue("email")
	avatar := c.FormValue("avatar")
	u.Name = name
	u.Email = email
	u.Avatar = avatar
	db.Model(&user).Updates(u)
	return c.JSON(http.StatusOK, user)
}

func (u *User) DeleteUser(c echo.Context) error {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	db.Delete(&user)
	return c.JSON(http.StatusOK, user)
}

func (u *User) Logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.JSON(http.StatusOK, "Logged Out")
}

func IsLoggedIn(c echo.Context) bool {
	sess, _ := session.Get("session", c)
	if sess.Values["user"] != nil {
		return true
	}
	return false
}

func GetUserSession(c echo.Context) User {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	return user
}

func GetUserSessionID(c echo.Context) string {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	return user.UUID
}

func GetUserSessionUserName(c echo.Context) string {
	sess, _ := session.Get("session", c)
	user := sess.Values["user"].(User)
	return user.Username
}
