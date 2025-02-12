package domains

import "time"

type Comment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	ArticleID int       `json:"article_id"`
	AuthorID  int       `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
