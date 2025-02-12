package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type ArticleResp struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AuthorEmail string `json:"author_email"`
}

type ArticlesResp struct {
	Articles   []*ArticleResp `json:"articles"`
	NextCursor int            `json:"next_cursor"`
}

type CreateArticlePayload struct {
	Title string `json:"title" validate:"required,max=128"`
	Body  string `json:"body" validate:"required"`
}

type CreateArticleResp struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type UpdateArticlePayload struct {
	Title string `json:"title" validate:"required,max=128"`
	Body  string `json:"body" validate:"required"`
}

func (a *API) getArticlesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paginationOptions := contextPaginationOptionsCtx(r.Context())

		articles, nextCursor, err := a.articleService.GetAll(
			paginationOptions.Cursor,
			paginationOptions.PageSize,
		)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		articleResps := make([]*ArticleResp, 0, len(articles))
		for _, a := range articles {
			resp := &ArticleResp{
				ID:          a.ID,
				Title:       a.Title,
				Body:        a.Body,
				Slug:        a.Slug,
				CreatedAt:   a.CreatedAt,
				UpdatedAt:   a.UpdatedAt,
				AuthorEmail: a.AuthorEmail,
			}

			articleResps = append(articleResps, resp)
		}

		resp := &ArticlesResp{
			Articles:   articleResps,
			NextCursor: nextCursor,
		}

		a.successResponse(w, resp, http.StatusOK)
	}
}

func (a *API) getArticleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		article, err := a.articleService.GetBySlug(slug)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := &ArticleResp{
			ID:          article.ID,
			Title:       article.Title,
			Body:        article.Body,
			Slug:        article.Slug,
			CreatedAt:   article.CreatedAt,
			UpdatedAt:   article.UpdatedAt,
			AuthorEmail: article.AuthorEmail,
		}

		a.successResponse(w, resp, http.StatusOK)
	}
}

func (a *API) createArticleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := &CreateArticlePayload{}
		claims := JWTUserClaimsCtx(r.Context())

		if err := a.fromJSON(r.Body, payload); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		if err := a.validator.Struct(payload); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		article := &domains.Article{
			Title:    payload.Title,
			Body:     payload.Body,
			AuthorID: claims.UserID,
		}
		if err := a.articleService.Create(article); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := &CreateArticleResp{
			ID:   article.ID,
			Slug: article.Slug,
		}
		a.successResponse(w, resp, http.StatusCreated)
	}
}

func (a *API) updateArticleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		claims := JWTUserClaimsCtx(r.Context())
		payload := &UpdateArticlePayload{}

		if err := a.fromJSON(r.Body, payload); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		if err := a.validator.Struct(payload); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		article := &domains.Article{
			Title:    payload.Title,
			Body:     payload.Body,
			Slug:     slug,
			AuthorID: claims.UserID,
		}
		if err := a.articleService.Update(article); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (a *API) deleteArticleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := JWTUserClaimsCtx(r.Context())

		slug := chi.URLParam(r, "slug")
		if err := a.articleService.DeleteBySlugAndUserID(slug, claims.UserID); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
