package blog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"moe-blogger/database"

	"github.com/labstack/echo/v4"
)

type Blog struct {
	ID      uint   `json:"id" gorm:"primary_key"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image"`
	Date    string `json:"date"`
}

type Comments struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	Comment  string `json:"comment"`
	BlogID   uint   `json:"blog_id"`
	Username string `json:"username"`
	Date     string `json:"date"`
}

var db = database.DB

func init() {
	db.AutoMigrate(&Blog{})
	db.AutoMigrate(&Comments{})
}

func GetBlogs(num int) []Blog {
	var blogs []Blog
	if num == 0 {
		db.Find(&blogs)
	} else {
		db.Limit(num).Find(&blogs)
	}

	return blogs
}

func GetBlog(id string) Blog {
	var blog Blog
	db.Where("id = ?", id).First(&blog)
	return blog
}

func CreateBlog(blog Blog) {
	db.Create(&blog)
}

func UpdateBlog(blog Blog) {
	db.Save(&blog)
}

func DeleteBlog(blog Blog) {
	db.Delete(&blog)
}

func GetComments(blogID string) []Comments {
	var comments []Comments
	db.Where("blog_id = ?", blogID).Find(&comments)
	return comments
}

func CreateComment(comment Comments) {
	db.Create(&comment)
}

func NewPostHandler(c echo.Context) error {
	title := c.FormValue("title")
	content := c.FormValue("content")

	// Extract the image file from the request
	imageFile, err := c.FormFile("image")
	if err != nil {
		return fmt.Errorf("failed to get image file from form: %w", err)
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Generate a unique filename for the image
	uniqueFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), imageFile.Filename)
	imagePath := filepath.Join(baseDir, "assets", "blog", uniqueFileName)

	// Save the image to local storage
	src, err := imageFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy image to destination: %w", err)
	}

	CreateBlog(Blog{Title: title, Content: content, Image: uniqueFileName, Date: time.Now().Format("2006-01-02")})

	return err
}

func dummydata() {
	CreateBlog(Blog{Title: "Hello World", Content: "This is a test blog post", Image: "test.jpg", Date: time.Now().Format("2006-01-02")})
	CreateBlog(Blog{Title: "Goodbye World", Content: "This is another test blog post", Image: "test.jpg", Date: time.Now().Format("2006-01-02")})
}
