package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/cache"
	"github.com/onexstack/realworld/apiserver/jwt"
	"github.com/onexstack/realworld/apiserver/middleware"
	mockstore "github.com/onexstack/realworld/apiserver/mock/store"
	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
	"gorm.io/gorm"
)

type contractTestStore struct {
	userStore    store.UserStore
	articleStore store.ArticleStore
	commentStore store.CommentStore
	tagStore     store.TagStore
}

func (s *contractTestStore) DB(ctx context.Context) *gorm.DB {
	return nil
}

func (s *contractTestStore) Cache() cache.ICache {
	return nil
}

func (s *contractTestStore) User() store.UserStore {
	return s.userStore
}

func (s *contractTestStore) Article() store.ArticleStore {
	return s.articleStore
}

func (s *contractTestStore) Comment() store.CommentStore {
	return s.commentStore
}

func (s *contractTestStore) Tag() store.TagStore {
	return s.tagStore
}

type duplicateEmailUserStore struct {
	*mockstore.MockUserStore
}

func (s *duplicateEmailUserStore) Get(ctx context.Context, condition interface{}) (*model.UserM, error) {
	return &model.UserM{
		ID:       1,
		Username: "current-user",
		Email:    "current@example.com",
	}, nil
}

func (s *duplicateEmailUserStore) GetByEmail(ctx context.Context, email string) (*model.UserM, error) {
	if email == "taken@example.com" {
		return &model.UserM{
			ID:       2,
			Username: "taken-user",
			Email:    email,
		}, nil
	}

	return nil, errors.New("user not found")
}

func (s *duplicateEmailUserStore) GetByUsername(ctx context.Context, username string) (*model.UserM, error) {
	return nil, errors.New("user not found")
}

func TestLoginResponseMatchesContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeInst := mockstore.NewMockStore()
	jwtManager := jwt.NewManager("test-secret", nil)
	h := NewHandler(biz.NewBiz(storeInst), jwtManager)

	router := gin.New()
	router.POST("/api/users/login", h.User().Login)

	body := []byte(`{"user":{"email":"test@example.com","password":"password123"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if refreshToken := rec.Header().Get("X-Refresh-Token"); refreshToken == "" {
		t.Fatalf("expected X-Refresh-Token header to be set")
	}

	var payload map[string]map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	user := payload["user"]
	required := []string{"email", "username", "bio", "image", "token"}
	for _, key := range required {
		if _, exists := user[key]; !exists {
			t.Fatalf("expected user.%s to exist", key)
		}
	}

	if _, exists := user["refresh_token"]; exists {
		t.Fatalf("did not expect user.refresh_token in response body")
	}
	if _, exists := user["following"]; exists {
		t.Fatalf("did not expect user.following in response body")
	}
}

func TestUpdateUserReturnsBadRequestOnEmailConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeInst := &contractTestStore{
		userStore:    &duplicateEmailUserStore{MockUserStore: &mockstore.MockUserStore{}},
		articleStore: &mockstore.MockArticleStore{},
		commentStore: &mockstore.MockCommentStore{},
		tagStore:     &mockstore.MockTagStore{},
	}
	jwtManager := jwt.NewManager("test-secret", nil)
	h := NewHandler(biz.NewBiz(storeInst), jwtManager)

	router := gin.New()
	auth := router.Group("/api")
	auth.Use(middleware.Auth(jwtManager))
	auth.PUT("/user", h.User().UpdateUser)

	token, err := jwtManager.GenerateToken(1)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body := []byte(`{"user":{"email":"taken@example.com"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteEndpointsReturnOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeInst := mockstore.NewMockStore()
	jwtManager := jwt.NewManager("test-secret", nil)
	h := NewHandler(biz.NewBiz(storeInst), jwtManager)

	router := gin.New()
	auth := router.Group("/api")
	auth.Use(middleware.Auth(jwtManager))
	auth.DELETE("/articles/:slug", h.Article().DeleteArticle)
	auth.DELETE("/articles/:slug/comments/:id", h.Comment().DeleteComment)

	token, err := jwtManager.GenerateToken(1)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	testCases := []string{
		"/api/articles/test-article",
		"/api/articles/test-article/comments/1",
	}

	for _, path := range testCases {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.Header.Set("Authorization", "Token "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d: %s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestParseOffsetAndLimitPrefersPage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/articles?page=2&limit=5&offset=99", nil)

	offset, limit := parseOffsetAndLimit(c)
	if offset != 5 {
		t.Fatalf("expected offset 5 from page=2&limit=5, got %d", offset)
	}
	if limit != 5 {
		t.Fatalf("expected limit 5, got %d", limit)
	}
}
