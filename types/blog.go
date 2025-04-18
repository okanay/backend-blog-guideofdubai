package types

import (
	"time"

	"github.com/google/uuid"
)

// BlogStatus - blog status enum type
type BlogStatus string

const (
	BlogStatusDraft     BlogStatus = "draft"
	BlogStatusPublished BlogStatus = "published"
	BlogStatusArchived  BlogStatus = "archived"
	BlogStatusDeleted   BlogStatus = "deleted"
)

// ----- DATABASE TABLE STRUCTURES -----

// BlogPost - main blog post structure
type BlogPost struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"userId" db:"user_id"`
	GroupID     string     `json:"groupId" db:"group_id"`
	Slug        string     `json:"slug" db:"slug"`
	Language    string     `json:"language" db:"language"`
	Featured    bool       `json:"featured" db:"featured"`
	Status      BlogStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	PublishedAt time.Time  `json:"publishedAt" db:"published_at"`
}

// BlogMetadata - blog metadata
type BlogMetadata struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Image       string    `json:"image" db:"image"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// BlogContent - blog content
type BlogContent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	ReadTime    int       `json:"readTime" db:"read_time"`
	HTML        string    `json:"html" db:"html"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// BlogStats - blog statistics
type BlogStats struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Views        int        `json:"views" db:"views"`
	Likes        int        `json:"likes" db:"likes"`
	Shares       int        `json:"shares" db:"shares"`
	Comments     int        `json:"comments" db:"comments"`
	LastViewedAt *time.Time `json:"lastViewedAt,omitempty" db:"last_viewed_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

// ----- VIEW STRUCTURES -----

// BlogPostView - blog post view structure
type BlogPostView struct {
	ID          string       `json:"id"`
	GroupID     string       `json:"groupId"`
	Slug        string       `json:"slug"`
	Language    string       `json:"language"`
	Featured    bool         `json:"featured"`
	Status      BlogStatus   `json:"status"`
	Metadata    MetadataView `json:"metadata"`
	Content     ContentView  `json:"content"`
	Stats       StatsView    `json:"stats"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	PublishedAt time.Time    `json:"publishedAt"`
}

// MetadataView - metadata view structure
type MetadataView struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// ContentView - content view structure
type ContentView struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ReadTime    int            `json:"readTime"`
	Categories  []CategoryView `json:"categories"`
	Tags        []TagView      `json:"tags"`
	HTML        string         `json:"html"`
}

// StatsView - statistics view structure
type StatsView struct {
	Views        int        `json:"views"`
	Likes        int        `json:"likes"`
	Shares       int        `json:"shares"`
	Comments     int        `json:"comments"`
	LastViewedAt *time.Time `json:"lastViewedAt,omitempty"`
}

// CategoryView - category view structure
type CategoryView struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TagView - tag view structure
type TagView struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ----- INPUT STRUCTURES -----

// BlogPostCreateInput - blog post creation input
type BlogPostCreateInput struct {
	GroupID    string        `json:"groupId" binding:"required"`
	Slug       string        `json:"slug" binding:"required"`
	Language   string        `json:"language" binding:"required"`
	Featured   bool          `json:"featured"`
	Metadata   MetadataInput `json:"metadata" binding:"required"`
	Content    ContentInput  `json:"content" binding:"required"`
	Categories []string      `json:"categories"`
	Tags       []string      `json:"tags"`
}

// MetadataInput - metadata input structure
type MetadataInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// ContentInput - content input structure
type ContentInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	ReadTime    int    `json:"readTime"`
	HTML        string `json:"html" binding:"required"`
}

// CategoryInput - category creation input
type CategoryInput struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// TagInput - tag creation input
type TagInput struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type BlogSelectByGroupIDInput struct {
	GroupID  string `json:"groupId" binding:"required"`
	Language string `json:"language" binding:"required"`
}
