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

func init() {
	database.Connect()
	database.DB.AutoMigrate(&Blog{})
	database.DB.AutoMigrate(&Comments{})
	database.Close()
}

func GetBlogs(num int) []Blog {
	database.Connect()
	var blogs []Blog
	if num == 0 {
		database.DB.Find(&blogs)
	} else {
		database.DB.Limit(num).Find(&blogs)
	}
	database.Close()

	return blogs
}

func GetBlog(id string) Blog {
	database.Connect()
	var blog Blog
	database.DB.Where("id = ?", id).First(&blog)
	database.Close()
	return blog
}

func CreateBlog(blog Blog) {
	database.Connect()
	database.DB.Create(&blog)
	database.Close()
}

func UpdateBlog(blog Blog) {
	database.Connect()
	database.DB.Save(&blog)
	database.Close()
}

func DeleteBlog(blog Blog) {
	database.Connect()
	database.DB.Delete(&blog)
	database.Close()
}

func GetComments(blogID string) []Comments {
	database.Connect()
	var comments []Comments
	database.DB.Where("blog_id = ?", blogID).Find(&comments)
	database.Close()
	return comments
}

func CreateComment(comment Comments) {
	database.Connect()
	database.DB.Create(&comment)
	database.Close()
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
