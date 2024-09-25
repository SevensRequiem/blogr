package auth

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"blogr.moe/database"
	"blogr.moe/logs"
	"blogr.moe/utils/mail"
	"blogr.moe/utils/queue"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

var (
	db         = database.DB
	ExpiryTime = 72 * time.Hour

	saltSize = 32
	keyLen   = 64
)
var manager = queue.NewQueueManager()
var q = manager.GetQueue("auth", 1000)

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
	GroupID       uint   `json:"group_id"`
	Group         Group  `json:"group" gorm:"foreignKey:GroupID"`
	Premium       bool   `json:"premium"`
	VerifiedEmail bool   `json:"verified_email"`
	TransactionID string `json:"transaction_id"`
}

type Group struct {
	Admin     bool `json:"admin"`
	Moderator bool `json:"moderator"`
	User      bool `json:"user"`
}

type UserLogin struct {
	ID       uint   `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Register struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	db.AutoMigrate(&User{}, &UserLogin{})
	ensureAdminUser()
}

func ensureAdminUser() {
	var user User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.Error("Error checking for admin user: ", err)
			return
		}
		createAdminUser()
	}
}

func createAdminUser() {
	encpass, err := encryptPassword("admin")
	if err != nil {
		logs.Error("Failed to encrypt password: ", err)
		return
	}
	uuid := genuuid()
	admin := User{
		UUID:        uuid,
		Email:       "admin@blogr.moe",
		Username:    "admin",
		DateCreated: time.Now().Format(time.RFC3339),
		Reputation:  0,
		TotalViews:  0,
		Group:       Group{Admin: true, Moderator: false, User: true},
		Premium:     true,
	}

	if err := db.Create(&admin).Error; err != nil {
		logs.Error("Failed to create admin user: ", err)
		return
	}

	adminlogin := UserLogin{
		UUID:     uuid,
		Username: "admin",
		Password: encpass,
	}

	if err := db.Create(&adminlogin).Error; err != nil {
		logs.Error("Failed to create admin login: ", err)
	}
}

func (g *Group) Scan(value interface{}) error {
	if value == nil {
		*g = Group{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported data type: %T", value)
	}
	return json.Unmarshal(bytes, g)
}

func (g Group) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// RegisterUser registers a new user
func RegisterUser(c echo.Context) error {
	var register Register
	if err := c.Bind(&register); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if register.Email == "" || register.Username == "" || register.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
	}

	encpass, err := encryptPassword(register.Password)
	if err != nil {
		logs.Error("Failed to encrypt password: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	uuid := genuuid()
	user := User{
		UUID:        uuid,
		Email:       register.Email,
		Username:    register.Username,
		DateCreated: time.Now().Format(time.RFC3339),
		Reputation:  0,
		TotalViews:  0,
		Group:       Group{Admin: false, Moderator: false, User: true},
		Premium:     false,
	}

	if err := db.Create(&user).Error; err != nil {
		logs.Error("Failed to create user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	userlogin := UserLogin{
		UUID:     uuid,
		Username: register.Username,
		Password: encpass,
	}

	if err := db.Create(&userlogin).Error; err != nil {
		logs.Error("Failed to create user login: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	mail.AddMailToQueue(register.Email, "Welcome to Blogr!", "Thank you for registering with Blogr!")
	q.Enqueue(func() {
		VerifyEmail(register.Email)
	})
	return c.JSON(http.StatusOK, map[string]string{"success": "user created"})
}

func VerifyEmail(email string) {
	user, err := GetUserByEmail(email)
	if err != nil {
		logs.Error("Failed to get user by email: ", err)
		return
	}
	uuid := user.UUID
	baseUrl := os.Getenv("BASE_URL")
	mail.AddMailToQueue(email, "Verify your email", "Click the link to verify your email: "+baseUrl+"/verify/"+uuid)
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func VerifyHandler(c echo.Context) error {
	uuid := c.Param("uuid")
	var user User
	if err := db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		logs.Error("Failed to query user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	user.VerifiedEmail = true
	if err := db.Save(&user).Error; err != nil {
		logs.Error("Failed to update user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"success": "email verified"})
}

// LoginHandler is a handler for the login route

func LoginHandler(c echo.Context) error {
	var login Login

	if err := c.Bind(&login); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if login.Username == "" || login.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
	}

	var userlogin UserLogin
	if err := db.Where("username = ?", login.Username).First(&userlogin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		logs.Error("Failed to query userlogin: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	if !comparePasswords(userlogin.Password, login.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	var user User
	if err := db.Where("uuid = ?", userlogin.UUID).First(&user).Error; err != nil {
		logs.Error("Failed to query user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	session, err := session.Get("session", c)
	if err != nil {
		logs.Error("Failed to get session: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	session.Values["uuid"] = user.UUID
	if err := session.Save(c.Request(), c.Response()); err != nil {
		logs.Error("Failed to save session: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"success": "login successful"})
}

// LogoutHandler is a handler for the logout route
func LogoutHandler(c echo.Context) error {
	session, err := session.Get("session", c)
	if err != nil {
		logs.Error("Failed to get session: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	session.Options.MaxAge = -1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		logs.Error("Failed to save session: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"success": "logout successful"})
}

// GetCurrentUser returns the current user
func GetCurrentUser(c echo.Context) (*User, error) {
	session, err := session.Get("session", c)
	if err != nil {
		logs.Error("Failed to get session: ", err)
		return nil, errors.New("internal server error")
	}

	uuid, ok := session.Values["uuid"].(string)
	if !ok {
		return nil, errors.New("unauthenticated")
	}

	var user User
	if err := db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		logs.Error("Failed to query user: ", err)
		return nil, errors.New("internal server error")
	}

	return &user, nil
}

func encryptPassword(password string) (string, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, 1<<14, 8, 1, keyLen)
	if err != nil {
		return "", err
	}

	encpass := base64.StdEncoding.EncodeToString(append(salt, hash...))
	return encpass, nil
}

func comparePasswords(encpass, password string) bool {
	decoded, err := base64.StdEncoding.DecodeString(encpass)
	if err != nil {
		logs.Error("Failed to decode password: ", err)
		return false
	}

	salt := decoded[:saltSize]
	hash, err := scrypt.Key([]byte(password), salt, 1<<14, 8, 1, keyLen)
	if err != nil {
		logs.Error("Failed to hash password: ", err)
		return false
	}

	return strings.Compare(string(hash), string(decoded[saltSize:])) == 0
}

func genuuid() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logs.Error("Failed to generate UUID: ", err)
		return ""
	}
	return uuid.String()
}

// AdminCheck checks if the current user is an admin
func AdminCheck(c echo.Context) (bool, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return false, err
	}

	if !user.Group.Admin {
		return false, nil
	}

	return true, nil
}

// ModeratorCheck checks if the current user is a moderator
func ModeratorCheck(c echo.Context) (bool, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return false, err
	}

	if !user.Group.Moderator {
		return false, nil
	}

	return true, nil
}

// UserCheck checks if the current user is a user
func UserCheck(c echo.Context) (bool, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return false, err
	}

	if !user.Group.User {
		return false, nil
	}

	return true, nil
}

func GetUserByID(uuid string) (*User, error) {
	var user User
	if err := db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func TotalUserCount() (int64, error) {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
