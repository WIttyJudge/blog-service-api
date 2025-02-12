package domains

import (
	"strconv"
	"strings"
	"time"
)

type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Slug      string    `json:"slug"`
	AuthorID  int       `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AuthorEmail string `json:"author_email"`
}

type ArticleRepository interface {
	GetAll(cursor int, pageSize int) ([]*Article, error)
	GetBySlug(slug string) (*Article, error)
	Create(article *Article) error
	Update(article *Article) error
	DeleteBySlugAndUserID(slug string, userID int) error
}

type ArticleService interface {
	GetAll(cursor int, pageSize int) ([]*Article, int, error)
	GetBySlug(slug string) (*Article, error)
	Create(article *Article) error
	Update(article *Article) error
	DeleteBySlugAndUserID(slug string, userID int) error
}

func (a *Article) TitleToSlug() string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	slug := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(a.Title)), " ", "-")
	slug = slug + "-" + timestamp

	return slug
}
