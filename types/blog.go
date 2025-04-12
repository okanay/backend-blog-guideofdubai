package types

import "time"

type BlogPost struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"groupId"`
	Slug        string    `json:"slug"`
	Metadata    Metadata  `json:"metadata"`
	Content     Content   `json:"content"`
	Stats       Stats     `json:"stats"`
	Language    Language  `json:"language"`
	Featured    bool      `json:"featured"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	PublishedAt time.Time `json:"publishedAt"`
	Version     int       `json:"version"`
}

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type Content struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ReadTime    int      `json:"readTime"`
	Tags        []string `json:"tags"`
	Categories  []string `json:"categories"`
	HTML        string   `json:"html"`
}

type Stats struct {
	Views int `json:"views"`
}
