package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"blogr.moe/backend/database"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/scrypt"
)

var (
	saltSize = 32
	keyLen   = 64
)

type UserCollection struct {
	Users []User `json:"users"`
}

type User struct {
	ID            uint   `json:"id" gorm:"primary_key"`
	UUID          string `json:"uuid"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Avatar        string `json:"avatar"`
	LastLogin     string `json:"last_login"`
	DateCreated   string `json:"date_created"`
	Reputation    int    `json:"reputation"`
	TotalViews    int    `json:"total_views"`
	GroupID       uint   `json:"group_id"`
	Premium       bool   `json:"premium"`
	PremiumExpiry string `json:"premium_expiry"`
	TransactionID string `json:"transaction_id"`
	VerifiedEmail bool   `json:"verified_email"`
	Webhook       string `json:"webhook"`
	Theme         string `json:"theme"`
}

type UserList struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
	Email    string `json:"email"`
}

type Stats struct {
	TotalUsers int `json:"total_users"`
}

func init() {
	gob.Register(&User{}) // Register the pointer type as well
}

type UserRegister struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(c echo.Context) error {
	var user UserRegister
	if err := c.Bind(&user); err != nil {
		log.Println("Error binding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad Request"})
	}

	if user.Email == "" || user.Username == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad Request"})
	}

	uuid := uuid.New().String()
	hash, err := hashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	userDoc := User{
		UUID:        uuid,
		Email:       user.Email,
		Username:    user.Username,
		Password:    hash,
		DateCreated: time.Now().Format("2006-01-02 15:04:05"),

		Reputation:    0,
		TotalViews:    0,
		GroupID:       1,
		Premium:       false,
		VerifiedEmail: false,
		Theme:         "light",

		Avatar:        "",
		LastLogin:     "",
		PremiumExpiry: "",
		TransactionID: "",
		Webhook:       "",
	}

	_, err = database.DB_Users.Collection(uuid).InsertOne(context.Background(), userDoc)
	if err != nil {
		log.Println("Error inserting user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	_, err = database.DB_UserList.Collection("users").InsertOne(context.Background(), map[string]string{"username": user.Username, "uuid": uuid, "email": user.Email})
	if err != nil {
		log.Println("Error inserting user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User created"})
}

func Login(c echo.Context) error {
	var user UserRegister
	if err := c.Bind(&user); err != nil {
		log.Println("Error binding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad Request"})
	}

	if user.Email == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad Request"})
	}

	usrList := &UserList{}
	userDoc := &User{}

	// find user by email, the collections are named by the user's UUID
	err := database.DB_UserList.Collection("users").FindOne(context.Background(), map[string]string{"email": user.Email}).Decode(&usrList)
	if err != nil {
		log.Println("Error finding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cant find user"})
	}

	err = database.DB_Users.Collection(usrList.UUID).FindOne(context.Background(), map[string]string{"email": user.Email}).Decode(&userDoc)
	if err != nil {
		log.Println("Error finding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cant find user"})
	}

	if !checkPassword(user.Password, userDoc.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	session, err := session.Get("session", c)
	if err != nil {
		log.Println("Error getting session:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400,
		Domain: os.Getenv("BASE_URL"),

		Secure:   true,
		HttpOnly: false,

		SameSite: http.SameSiteStrictMode,
	}

	session.Values["user"] = User{
		ID:            userDoc.ID,
		UUID:          userDoc.UUID,
		Email:         userDoc.Email,
		Username:      userDoc.Username,
		Avatar:        userDoc.Avatar,
		LastLogin:     userDoc.LastLogin,
		DateCreated:   userDoc.DateCreated,
		Reputation:    userDoc.Reputation,
		TotalViews:    userDoc.TotalViews,
		GroupID:       userDoc.GroupID,
		Premium:       userDoc.Premium,
		PremiumExpiry: userDoc.PremiumExpiry,
		TransactionID: userDoc.TransactionID,
		VerifiedEmail: userDoc.VerifiedEmail,
		Webhook:       userDoc.Webhook,
		Theme:         userDoc.Theme,
	}
	if err := session.Save(c.Request(), c.Response()); err != nil {
		log.Println("Error saving session:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged in"})
}

func Logout(c echo.Context) error {
	session, err := session.Get("session", c)
	if err != nil {
		log.Println("Error getting session:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	// Invalidate the session
	session.Options.MaxAge = -1

	if err := session.Save(c.Request(), c.Response()); err != nil {
		log.Println("Error saving session:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.Redirect(http.StatusFound, "/")
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func hashPassword(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, 1<<14, 8, 1, keyLen)
	if err != nil {
		return "", err
	}

	saltHash := base64.StdEncoding.EncodeToString(salt) + "." + base64.StdEncoding.EncodeToString(hash)
	return saltHash, nil
}

func checkPassword(password, hash string) bool {
	saltHash := strings.Split(hash, ".")
	if len(saltHash) != 2 {
		return false // Invalid hash format
	}

	saltBytes, err := base64.StdEncoding.DecodeString(saltHash[0])
	if err != nil {
		return false // Failed to decode salt
	}

	hashBytes, err := base64.StdEncoding.DecodeString(saltHash[1])
	if err != nil {
		return false // Failed to decode hash
	}

	newHash, err := scrypt.Key([]byte(password), saltBytes, 1<<14, 8, 1, keyLen)
	if err != nil {
		return false
	}

	return bytes.Equal(hashBytes, newHash)
}

func GetUserByUsername(username string) (User, error) {
	userDoc := User{}
	userList := UserList{}

	err := database.DB_UserList.Collection("users").FindOne(context.Background(), map[string]string{"username": username}).Decode(&userList)
	if err != nil {
		return User{}, err
	}

	err = database.DB_Users.Collection(userList.UUID).FindOne(context.Background(), map[string]string{"username": username}).Decode(&userDoc)
	if err != nil {
		return User{}, err
	}

	return userDoc, nil
}

func UsernameToUUID(c echo.Context) (string, error) {
	username := c.Param("username")
	userDoc := UserList{}
	err := database.DB_UserList.Collection("users").FindOne(context.Background(), map[string]string{"username": username}).Decode(&userDoc)
	if err != nil {
		log.Println("Error finding user:", err)
		return "", err
	}
	return userDoc.UUID, nil
}

func GetUserByID(c echo.Context) error {
	id := c.Param("id")
	userDoc := User{}
	err := database.DB_Users.Collection(id).FindOne(context.Background(), map[string]string{"uuid": id}).Decode(&userDoc)
	if err != nil {
		log.Println("Error finding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cant find user"})
	}
	return c.JSON(http.StatusOK, userDoc)
}

func UpdateUser(c echo.Context) error {
	user := c.Get("user").(User)
	var userUpdate User
	if err := c.Bind(&userUpdate); err != nil {
		log.Println("Error binding user:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad Request"})
	}

	if userUpdate.Email != "" {
		user.Email = userUpdate.Email
	}
	if userUpdate.Username != "" {
		user.Username = userUpdate.Username
	}
	if userUpdate.Avatar != "" {
		user.Avatar = userUpdate.Avatar
	}
	if userUpdate.Theme != "" {
		user.Theme = userUpdate.Theme
	}

	_, err := database.DB_Users.Collection(user.UUID).UpdateOne(context.Background(), map[string]string{"uuid": user.UUID}, user)
	if err != nil {
		log.Println("Error updating user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "User updated"})
}

func DeleteUser(c echo.Context) error {
	user := c.Get("user").(User)
	_, err := database.DB_Users.Collection("users").DeleteOne(context.Background(), map[string]string{"uuid": user.UUID})
	if err != nil {
		log.Println("Error deleting user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted"})
}

func GetUserFromContext(c echo.Context) User {
	session, err := session.Get("session", c)
	if err != nil {
		log.Println("Error getting session:", err)
		return User{}
	}
	user := session.Values["user"]
	if user == nil {
		return User{}
	}
	if u, ok := user.(*User); ok {
		return *u
	}
	return User{}
}

func IsLoggedIn(c echo.Context) bool {
	session, err := session.Get("session", c)
	if err != nil {
		log.Println("Error getting session:", err)
		return false
	}
	user := session.Values["user"]
	if user == nil {
		return false
	}
	return true
}

func IsPremium(c echo.Context) bool {
	user := c.Get("user").(User)
	return user.Premium
}

func GetUserByEmail(email string) (User, error) {
	userDoc := User{}
	err := database.DB_Users.Collection("users").FindOne(context.Background(), map[string]string{"email": email}).Decode(&userDoc)
	if err != nil {
		return User{}, err
	}
	return userDoc, nil
}

func GetUserByUUID(uuid string) (User, error) {
	userDoc := User{}
	err := database.DB_Users.Collection(uuid).FindOne(context.Background(), map[string]string{"uuid": uuid}).Decode(&userDoc)
	if err != nil {
		return User{}, err
	}
	return userDoc, nil
}

func IsPaymentActive(c echo.Context) bool {
	user := c.Get("user").(User)
	if !user.Premium {
		return false
	}
	expiry, err := time.Parse(time.RFC3339, user.PremiumExpiry)
	if err != nil {
		return false
	}
	if time.Now().After(expiry) {
		return false
	}
	return true
}

func VerifyEmail(c echo.Context) error {
	user := c.Get("user").(User)
	user.VerifiedEmail = true
	_, err := database.DB_Users.Collection(user.UUID).UpdateOne(c.Request().Context(), map[string]string{"email": user.Email}, user)
	if err != nil {
		log.Println("Error updating user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Email verified"})
}
