package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type ArticlePostgres struct {
	ctx    context.Context
	pgPool *pgxpool.Pool
}

func NewArticlePostgres(ctx context.Context, pgPool *pgxpool.Pool) *ArticlePostgres {
	return &ArticlePostgres{
		ctx:    ctx,
		pgPool: pgPool,
	}
}

func (p *ArticlePostgres) GetAll(cursor int, pageSize int) ([]*domains.Article, error) {
	sql := `
	  SELECT 
			a.id,
			a.title,
			a.body,
			a.slug,
			a.created_at,
			a.updated_at,
			u.email
		FROM articles a
		LEFT JOIN users u
			ON u.id = a.author_id
		WHERE a.id >= @id
		LIMIT @limit
	`

	args := pgx.NamedArgs{
		"id":    cursor,
		"limit": pageSize + 1,
	}

	articles, err := p.fetch(sql, args)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (p *ArticlePostgres) GetBySlug(slug string) (*domains.Article, error) {
	sql := `
	  SELECT 
			a.id,
			a.title,
			a.body,
			a.slug,
			a.created_at,
			a.updated_at,
			u.email
		FROM articles a
		LEFT JOIN users u
			ON u.id = a.author_id
		WHERE slug = @slug
	`

	args := pgx.NamedArgs{"slug": slug}

	articles, err := p.fetch(sql, args)
	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		return nil, domains.ErrArticleNotFound
	}

	return articles[0], nil
}

func (p *ArticlePostgres) Update(article *domains.Article) error {
	sql := `
	  UPDATE articles
		SET title = @title,
			body = @body,
			slug = @newSlug,
			updated_at = @updatedAt
		WHERE slug = @currentSlug
		AND author_id = @authorID
		RETURNING id, slug
	`

	args := pgx.NamedArgs{
		"title":       article.Title,
		"body":        article.Body,
		"newSlug":     article.TitleToSlug(),
		"currentSlug": article.Slug,
		"authorID":    article.AuthorID,
		"updatedAt":   time.Now(),
	}

	return p.pgPool.QueryRow(p.ctx, sql, args).Scan(
		&article.ID,
		&article.Slug,
	)
}

func (p *ArticlePostgres) Create(article *domains.Article) error {
	sql := `
		INSERT INTO articles (title, body, slug, author_id) 
		VALUES (@title, @body, @slug, @authorID)
		RETURNING id, slug

	`
	args := pgx.NamedArgs{
		"title":    article.Title,
		"body":     article.Body,
		"slug":     article.TitleToSlug(),
		"authorID": article.AuthorID,
	}

	return p.pgPool.QueryRow(p.ctx, sql, args).Scan(
		&article.ID,
		&article.Slug,
	)
}

func (p *ArticlePostgres) DeleteBySlugAndUserID(slug string, userID int) error {
	sql := `
		DELETE FROM articles
		WHERE slug = @slug
		AND author_id = @authorID
	`

	args := pgx.NamedArgs{
		"slug":     slug,
		"authorID": userID,
	}

	_, err := p.pgPool.Exec(p.ctx, sql, args)
	return err
}

func (a *ArticlePostgres) fetch(sql string, args ...any) ([]*domains.Article, error) {
	rows, err := a.pgPool.Query(a.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := []*domains.Article{}

	for rows.Next() {
		article := &domains.Article{}
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Body,
			&article.Slug,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.AuthorEmail,
		); err != nil {
			return nil, err
		}

		articles = append(articles, article)
	}

	return articles, nil
}
