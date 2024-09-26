package blog

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"blogr.moe/auth"
	"blogr.moe/database"
	"blogr.moe/logs"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/image/draw"
)

type BlogPost struct {
	ID       uint     `json:"id" gorm:"primary_key"`
	BlogID   uint     `json:"blog_id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     string   `json:"tags"`
	Image    string   `json:"image"`
	Date     string   `json:"date"`
	Author   string   `json:"author"`
	Comments Comments `json:"comments" gorm:"foreignKey:BlogID"`
	CSS      string   `json:"css"`
	Views    int      `json:"views"`
}

type Comments struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	Comment  string `json:"comment"`
	BlogID   uint   `json:"blog_id"`
	Username string `json:"username"`
	Date     string `json:"date"`
}

type BlogTrack struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	BlogID uint   `json:"blog_id"`
	Thumb  string `json:"thumb"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Date   string `json:"date"`
	Views  int    `json:"views"`
	Dir    string `json:"dir"`
}

type TotalPosts struct {
	Count int `json:"count"`
}

var db = database.DB

func init() {
	gob.Register(BlogPost{})
	db.AutoMigrate(&BlogTrack{})
	db.AutoMigrate(&TotalPosts{})

}
func GetTotalPosts() int {
	var count struct {
		Count int
	}
	db.Table("blogs").Select("count(*) as count").Scan(&count)
	return count.Count
}

func NewBlogHandler(c echo.Context) error {
	content := c.FormValue("postcontent")
	title := c.FormValue("title")
	tags := strings.Split(c.FormValue("tags"), ",")
	image, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid image file")
	}
	return NewBlog(c, content, title, tags, image)
}

func NewBlog(c echo.Context, content string, title string, tags []string, image *multipart.FileHeader) error {
	blogDir := "blogs/"
	// Get the session
	session, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve session")
	}

	// Get the user id
	userid, ok := session.Values["uuid"].(string)
	if !ok {
		logs.Error("Failed to get user id")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session")
	}

	// Check if the user has a blog dir
	usrBlogDir := blogDir + userid
	if _, err := os.Stat(usrBlogDir); os.IsNotExist(err) {
		if err := os.Mkdir(usrBlogDir, 0755); err != nil {
			logs.Error("Failed to create user blog dir", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user blog dir")
		}
	}
	blogid := genblogid()
	// make blog post dir
	if err := os.Mkdir(usrBlogDir+"/"+strconv.FormatUint(uint64(blogid), 10), 0755); err != nil {
		logs.Error("Failed to create blog post dir", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create blog post dir")
	}
	// Create the blog post
	blog := BlogPost{
		BlogID:  blogid,
		Title:   title,
		Content: content,
		Tags:    strings.Join(tags, ","),
		Image:   usrBlogDir + "/" + strconv.FormatUint(uint64(blogid), 10) + "/" + image.Filename,
		Date:    time.Now().Format("2006-01-02"),
	}

	// Save the blog post to gob
	imageFile, err := image.Open()
	if err != nil {
		logs.Error("Failed to open image file", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open image file")
	}
	defer imageFile.Close()

	dst, err := os.Create(usrBlogDir + "/" + strconv.FormatUint(uint64(blogid), 10) + "/" + image.Filename)
	if err != nil {
		logs.Error("Failed to create image file", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create image file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, imageFile); err != nil {
		logs.Error("Failed to copy image file", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to copy image file")
	}

	// Save the blog post to the .gob file
	blogFilePath := usrBlogDir + "/" + strconv.FormatUint(uint64(blogid), 10) + "/blogpost.gob"
	blogFile, err := os.Create(blogFilePath)
	if err != nil {
		logs.Error("Failed to create blog file", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create blog file")
	}
	defer blogFile.Close()

	if err := gob.NewEncoder(blogFile).Encode(blog); err != nil {
		logs.Error("Failed to encode blog post", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to encode blog post")
	}

	// create the thumbnail
	GenerateThumbnail(usrBlogDir+"/"+strconv.FormatUint(uint64(blogid), 10)+"/"+image.Filename, usrBlogDir+"/thumb_"+image.Filename, 200, 200)
	username := auth.GetUsername(userid)
	// Save the blog post to the database
	blogtrack := BlogTrack{
		Thumb:  usrBlogDir + "/thumb_" + image.Filename,
		Title:  title,
		Author: username,
		Date:   time.Now().Format("2006-01-02 15:04:05"),
		Views:  0,
		BlogID: blogid,
		Dir:    usrBlogDir,
	}

	if err := db.Create(&blogtrack).Error; err != nil {
		logs.Error("Failed to save blog post to database", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save blog post to database")
	}

	// Save the total posts
	totalPosts := TotalPosts{}
	db.First(&totalPosts)
	totalPosts.Count++
	db.Save(&totalPosts)

	return nil
}
func genblogid() uint {
	return uint(time.Now().Unix())
}

func GenerateThumbnail(inputPath, outputPath string, maxWidth, maxHeight int) error {
	// Get file extension
	ext := filepath.Ext(inputPath)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".tiff", ".bmp", ".webp", ".avif":
		return imageThumbnail(inputPath, outputPath, maxWidth, maxHeight)
	case ".webm", ".mp4", ".mov":
		return videoThumbnail(inputPath, outputPath, maxWidth, maxHeight)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

func imageThumbnail(inputPath, outputPath string, maxWidth, maxHeight int) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open the input file: %w", err)
	}
	defer inputFile.Close()

	inputImage, _, err := image.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode the input image: %w", err)
	}

	bounds := inputImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width > maxWidth || height > maxHeight {
		width, height = resize(width, height, maxWidth, maxHeight)
	}

	thumbnail := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), inputImage, bounds, draw.Over, nil)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create the output file: %w", err)
	}
	defer outputFile.Close()

	if err := jpeg.Encode(outputFile, thumbnail, nil); err != nil {
		return fmt.Errorf("failed to encode the thumbnail: %w", err)
	}

	return nil
}

func resize(width, height, maxWidth, maxHeight int) (int, int) {
	if maxWidth <= 0 || maxHeight <= 0 {
		return width, height // No resizing if max dimensions are invalid
	}
	if width > height {
		height = height * maxWidth / width
		width = maxWidth
	} else {
		width = width * maxHeight / height
		height = maxHeight
	}
	return width, height
}

func videoThumbnail(inputPath, outputPath string, maxWidth, maxHeight int) error {
	ext := filepath.Ext(inputPath)
	if ext != ".webm" && ext != ".mp4" && ext != ".mov" {
		return fmt.Errorf("unsupported video format: %s", ext)
	}
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vf", fmt.Sprintf("thumbnail,scale=%d:%d", maxWidth, maxHeight), "-frames:v", "1", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate the video thumbnail: %w", err)
	}
	return nil
}

func GetBlog(c echo.Context) error {
	blogID := c.Param("id")
	var blogtrack BlogTrack
	if err := db.Where("blog_id = ?", blogID).First(&blogtrack).Error; err != nil {
		logs.Error("Failed to get blog post", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get blog post")
	}

	blogpost := BlogPost{}
	blogFilePath := blogtrack.Dir + "/" + blogID + "/blogpost.gob"

	blogFile, err := os.Open(blogFilePath)
	if err != nil {
		logs.Error("Failed to open blog file", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open blog file")
	}
	defer blogFile.Close()

	if err := gob.NewDecoder(blogFile).Decode(&blogpost); err != nil {
		logs.Error("Failed to decode blog post", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decode blog post")
	}

	// Increment the views
	blogtrack.Views++
	if err := db.Save(&blogtrack).Error; err != nil {
		logs.Error("Failed to increment views", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to increment views")
	}

	return c.JSON(http.StatusOK, blogpost)
}

func GetUserBlogPosts(c echo.Context) error {
	// Get the session
	session, err := session.Get("session", c)
	if err != nil {
		logs.Error("Could not retrieve session", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve session")
	}

	// Get the user id
	userid, ok := session.Values["uuid"].(string)
	if !ok {
		logs.Error("Failed to get user id")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session")
	}

	// Log the user ID
	logs.Info("User ID:", userid)

	var blogtracks []BlogTrack
	author := auth.GetUsername(userid)
	logs.Info("Querying blog posts for author:", author)
	if err := db.Where("author = ?", author).Find(&blogtracks).Error; err != nil {
		logs.Error("Failed to get user blog posts", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user blog posts")
	}

	var blogposts []BlogPost
	for _, blogtrack := range blogtracks {
		blogpost := BlogPost{}
		blogFilePath := blogtrack.Dir + "/" + strconv.FormatUint(uint64(blogtrack.BlogID), 10) + "/blogpost.gob"

		// Log the blog file path
		logs.Info("Opening blog file at path:", blogFilePath)

		blogFile, err := os.Open(blogFilePath)
		if err != nil {
			// Log the specific error
			logs.Error("Failed to open blog file at path:", blogFilePath, "Error:", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open blog file")
		}
		defer blogFile.Close()

		if err := gob.NewDecoder(blogFile).Decode(&blogpost); err != nil {
			logs.Error("Failed to decode blog post", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decode blog post")
		}

		blogposts = append(blogposts, blogpost)
	}

	// Sort blogposts by Date ascending
	sort.Slice(blogposts, func(i, j int) bool {
		return blogposts[i].Date < blogposts[j].Date
	})

	return c.JSON(http.StatusOK, blogposts)
}

func GetBlogFile(c echo.Context) error {
	uuid := c.Param("uuid")
	file := c.Param("file")
	return c.File("blogs/" + uuid + "/" + file)
}
