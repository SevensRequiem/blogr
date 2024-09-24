package blog

import (
	"blogr.moe/database"
)

type Blog struct {
	ID       uint       `json:"id" gorm:"primary_key"`
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Tags     string     `json:"tags"`
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

var db = database.DB
