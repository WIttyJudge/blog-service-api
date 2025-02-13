package services

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"go.uber.org/zap"
)

type ArticleService struct {
	ctx    context.Context
	logger *zap.Logger

	repo        domains.ArticleRepository
	redisClient *redis.Client
}

func NewArticleService(ctx context.Context, logger *zap.Logger, repo domains.ArticleRepository, redisClient *redis.Client) *ArticleService {
	return &ArticleService{
		ctx:    ctx,
		logger: logger,

		repo:        repo,
		redisClient: redisClient,
	}
}

func (p *ArticleService) GetAll(cursor int, pageSize int) ([]*domains.Article, int, error) {
	articles, err := p.repo.GetAll(cursor, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if len(articles) < pageSize {
		return articles, -1, nil
	}

	nextCursor := articles[len(articles)-1].ID
	articles = articles[:len(articles)-1]

	return articles, nextCursor, nil
}

func (s *ArticleService) GetBySlug(slug string) (*domains.Article, error) {
	return s.repo.GetBySlug(slug)
}

func (s *ArticleService) Create(article *domains.Article) error {
	return s.repo.Create(article)
}

func (s *ArticleService) Update(article *domains.Article) error {
	return s.repo.Update(article)
}

func (s *ArticleService) DeleteBySlugAndUserID(slug string, userID int) error {
	return s.repo.DeleteBySlugAndUserID(slug, userID)
}
