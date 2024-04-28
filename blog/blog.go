package blog

import (
	"time"
)

type BlogPost struct {
	ID          int64     `json:"id"`
	Image       string    `json:"image"`
	Post        string    `json:"post"`
	Title       string    `json:"title"`
	Tags        string    `json:"tags"`
	Author      string    `json:"author"`
	DateCreated time.Time `json:"date_created"`
}
