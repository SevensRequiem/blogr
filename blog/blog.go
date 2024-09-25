package blog

import (
	"log"
	"strings"
	"time"

	"blogr.moe/auth"
	"blogr.moe/database"
	"blogr.moe/utils/queue"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	UUID      string `json:"uuid"`
	Blog      []Blog `json:"blog_list"`
	SubDomain string `json:"sub_domain"`
}

type Blog struct {
	ID       uint       `json:"id" gorm:"primary_key"`
	BlogID   uint       `json:"blog_id"`
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Tags     []Tag      `json:"tags" gorm:"many2many:blog_tags;"`
	Image    string     `json:"image"`
	Date     string     `json:"date"`
	Author   string     `json:"author"`
	Comments []Comments `json:"comments"`
	CSS      string     `json:"css"`
	Views    int        `json:"views"`
}

type Comments struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	Comment  string `json:"comment"`
	BlogID   uint   `json:"blog_id"`
	Username string `json:"username"`
	Date     string `json:"date"`
}

type CSS struct {
	ID    uint   `json:"id" gorm:"primary_key"`
	Title string `json:"title"`
	CSS   string `json:"css"`
}

type Tag struct {
	ID   uint   `json:"id" gorm:"primary_key"`
	Name string `json:"name"`
}

type PostCount struct {
	ID    uint `json:"id" gorm:"primary_key"`
	Count int  `json:"count"`
}

var db = database.DB
var manager = queue.NewQueueManager()
var q = manager.GetQueue("blog", 1000)

func GetTotalPosts() int {
	var count PostCount
	db.Table("blogs").Select("count(*) as count").Scan(&count)
	return count.Count
}

func NewBlogHandler(c echo.Context) error {
	q.Enqueue(func() {
		NewBlog(c)
	})
	return nil
}

func NewBlog(c echo.Context) error {
	usrcheck, err := auth.UserCheck(c)
	if err != nil {
		return echo.NewHTTPError(500, "Failed to check user")
	}
	if !usrcheck {
		return echo.NewHTTPError(401, "User not authenticated")
	}
	var blog Blog
	if err := c.Bind(&blog); err != nil {
		return echo.NewHTTPError(400, "Invalid request payload")
	}

	var blogpost Blog
	session, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(500, "Could not retrieve session")
	}

	uuid, ok := session.Values["user"].(string)
	if !ok {
		return echo.NewHTTPError(401, "User not authenticated")
	}

	var bloguser User
	if err := db.Where("uuid = ?", uuid).First(&bloguser).Error; err != nil {
		return echo.NewHTTPError(404, "User not found")
	}

	user, err := auth.GetUserByID(uuid)
	if err != nil {
		log.Println("Error fetching user:", err)
		return echo.NewHTTPError(500, "Failed to fetch user")
	}
	tagsStr := c.FormValue("tags")
	tagsSlice := strings.Split(tagsStr, ",")
	var tags []Tag
	for _, tag := range tagsSlice {
		tags = append(tags, Tag{Name: strings.TrimSpace(tag)})
	}
	blogpost = Blog{
		Title:   c.FormValue("title"),
		BlogID:  genblogid(),
		Content: c.FormValue("content"),
		Tags:    tags,
		Image:   c.FormValue("image"),
		Date:    time.Time.String(time.Now()),
		Author:  user.Username,
	}

	if err := db.Create(&blogpost).Error; err != nil {
		return echo.NewHTTPError(500, "Failed to create blogpost")
	}

	bloguser.Blog = append(bloguser.Blog, blogpost)
	if err := db.Save(&bloguser).Error; err != nil {
		return echo.NewHTTPError(500, "Failed to save blogpost")
	}

	return nil
}

func genblogid() uint {
	return uint(time.Now().Unix())
}

func GetBlog(c echo.Context) error {
	var blog Blog
	id := c.Param("id")
	if err := db.Where("id = ?", id).First(&blog).Error; err != nil {
		return echo.NewHTTPError(404, "Blog not found")
	}

	blog.Views++
	db.Save(&blog)

	return c.JSON(200, blog)
}

func GetBlogs(c echo.Context) error {
	var blogs []Blog
	if err := db.Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(500, "Failed to fetch blogs")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByTag(c echo.Context) error {
	var blogs []Blog
	tag := c.Param("tag")
	if err := db.Where("name = ?", tag).Preload("Tags").Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "Tag not found")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByUser(c echo.Context) error {
	var blogs []Blog
	user := c.Param("user")
	if err := db.Where("author = ?", user).Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "User not found")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByDate(c echo.Context) error {
	var blogs []Blog
	date := c.Param("date")
	if err := db.Where("date = ?", date).Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "Date not found")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByTitle(c echo.Context) error {
	var blogs []Blog
	title := c.Param("title")
	if err := db.Where("title = ?", title).Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "Title not found")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByContent(c echo.Context) error {
	var blogs []Blog
	content := c.Param("content")
	if err := db.Where("content = ?", content).Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "Content not found")
	}

	return c.JSON(200, blogs)
}

func GetBlogsByViews(c echo.Context) error {
	var blogs []Blog
	if err := db.Order("views desc").Find(&blogs).Error; err != nil {
		return echo.NewHTTPError(404, "Views not found")
	}

	return c.JSON(200, blogs)
}

func GetComments(c echo.Context) error {
	var comments []Comments
	id := c.Param("id")
	if err := db.Where("blog_id = ?", id).Find(&comments).Error; err != nil {
		return echo.NewHTTPError(404, "Comments not found")
	}

	return c.JSON(200, comments)
}
