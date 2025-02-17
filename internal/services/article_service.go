package services

import (
	"context"
	"time"

	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/pkg/cache"
	"go.uber.org/zap"
)

type ArticleService struct {
	ctx                context.Context
	logger             *zap.Logger
	repo               domains.ArticleRepository
	articleBySlugCache *cache.Cache[string, *domains.Article]
}

func NewArticleService(ctx context.Context, logger *zap.Logger, repo domains.ArticleRepository) *ArticleService {
	return &ArticleService{
		ctx:                ctx,
		logger:             logger,
		repo:               repo,
		articleBySlugCache: cache.New[string, *domains.Article](1000, "article_by_slug_cache"),
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
	if cached, ok := s.articleBySlugCache.Get(slug); ok {
		return cached, nil
	}

	article, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	s.articleBySlugCache.Set(slug, article, 10*time.Minute)

	return article, nil
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
