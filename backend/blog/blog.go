package blog

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"blogr.moe/backend/auth"
	"blogr.moe/backend/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogPost struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BlogID   string             `bson:"blog_id" json:"blog_id"`
	Title    string             `bson:"title" json:"title"`
	Content  string             `bson:"content" json:"content"`
	Tags     string             `bson:"tags" json:"tags"`
	Image    primitive.ObjectID `bson:"image" json:"image"`
	Date     string             `bson:"date" json:"date"`
	Author   string             `bson:"author" json:"author"`
	Comments []Comment          `bson:"comments" json:"comments"`
	CSS      string             `bson:"css" json:"css"`
	Views    int                `bson:"views" json:"views"`
}

type Comment struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Comment  string             `bson:"comment" json:"comment"`
	BlogID   uuid.UUID          `bson:"blog_id" json:"blog_id"`
	Username string             `bson:"username" json:"username"`
	Date     string             `bson:"date" json:"date"`
}

type TotalPosts struct {
	Count int `json:"count"`
}

func generateRandomString(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:n]
}

func NewBlogHandler(c echo.Context) error {
	blog := new(BlogPost)

	user := auth.GetUserFromContext(c)
	if user.Email == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	blog.Title = c.FormValue("title")
	blog.Content = c.FormValue("content")
	tags := c.FormValue("tags")
	tagsSlice := strings.Split(tags, ",")
	sort.Strings(tagsSlice)
	sortedTags := strings.Join(tagsSlice, ",")
	blog.Tags = strconv.Quote(sortedTags)
	blog.Author = user.Username
	blog.Date = time.Now().Format(time.RFC3339)
	blog.Views = 0
	blog.BlogID = generateRandomString(6)
	uuid := user.UUID

	// Check if the image file is provided
	image, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Image file is required"})
	}

	// gridfs image
	file, err := image.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error opening image file"})
	}
	defer file.Close()

	bucket, err := gridfs.NewBucket(database.DB_Users)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error creating GridFS bucket"})
	}

	data, err := bucket.UploadFromStream(image.Filename, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading image file"})
	}
	blog.Image = data

	res, err := database.DB_Users.Collection(uuid).InsertOne(c.Request().Context(), blog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error creating blog"})
	}
	_, err = database.DB_Main.Collection("posts").InsertOne(c.Request().Context(), blog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error creating blog"})
	}

	return c.JSON(http.StatusCreated, res)
}
func GetLatestPostsUser(c echo.Context) error {
	user := auth.GetUserFromContext(c)
	if user.Email == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	uuid := user.UUID

	// Extract pagination parameters
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Define the filter to exclude the document with ID 0 and empty string
	filter := bson.M{
		"blog_id": bson.M{
			"$ne": "",
		},
	}

	// Define options to skip and limit the documents
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Find the documents
	cursor, err := database.DB_Users.Collection(uuid).Find(c.Request().Context(), filter, findOptions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error fetching posts"})
	}

	// Iterate through the cursor and decode the documents
	var posts []BlogPost
	for cursor.Next(c.Request().Context()) {
		var post BlogPost
		cursor.Decode(&post)
		posts = append(posts, post)
	}

	return c.JSON(http.StatusOK, posts)
}

func GetLatestPosts(c echo.Context) error {
	// Extract pagination parameters
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Define the filter to exclude the document with ID 0
	filter := bson.M{"blog_id": bson.M{"$ne": ""}}
	// Define options to skip and limit the documents
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Find the documents
	cursor, err := database.DB_Main.Collection("posts").Find(c.Request().Context(), filter, findOptions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error fetching posts"})
	}

	// Iterate through the cursor and decode the documents
	var posts []BlogPost
	for cursor.Next(c.Request().Context()) {
		var post BlogPost
		cursor.Decode(&post)
		posts = append(posts, post)
	}

	return c.JSON(http.StatusOK, posts)

}

func DeleteUserPost(c echo.Context) error {
	id := c.Param("id")
	user := auth.GetUserFromContext(c)
	if user.Email == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	uuid := user.UUID

	// Define the filter to exclude the document with ID 0
	filter := bson.M{"blog_id": id}

	_, err := database.DB_Users.Collection(uuid).DeleteOne(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error deleting post"})
	}

	_, err = database.DB_Main.Collection("posts").DeleteOne(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error deleting post"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Post deleted"})
}

func GetPost(c echo.Context, user string, id string) (*BlogPost, error) {

	userDoc, err := auth.GetUserByUsername(user)
	if err != nil {
		return nil, fmt.Errorf("Error fetching user")
	}

	filter := bson.M{"blog_id": id}

	var post BlogPost
	err = database.DB_Users.Collection(userDoc.UUID).FindOne(c.Request().Context(), filter).Decode(&post)
	if err != nil {
		return nil, fmt.Errorf("Error fetching post")
	}

	return &post, nil
}

func decode(image primitive.ObjectID) primitive.ObjectID {
	bucket, _ := gridfs.NewBucket(database.DB_Users)
	var buf bytes.Buffer
	_, err := bucket.DownloadToStream(image, &buf)
	if err != nil {
		return primitive.NilObjectID
	}
	img := buf.Bytes()
	_ = img
	return image
}

func GetPostImage(c echo.Context) error {
	postid := c.Param("postid")
	user, err := auth.GetUserByUsername(c.Param("user"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error fetching user"})
	}
	uuid := user.UUID

	filter := bson.M{"blog_id": postid}

	var post BlogPost
	err = database.DB_Users.Collection(uuid).FindOne(c.Request().Context(), filter).Decode(&post)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error fetching post"})
	}

	bucket, _ := gridfs.NewBucket(database.DB_Users)
	var buf bytes.Buffer

	_, err = bucket.DownloadToStream(post.Image, &buf)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error fetching image"})
	}

	return c.Blob(http.StatusOK, "image/jpeg", buf.Bytes())
}
