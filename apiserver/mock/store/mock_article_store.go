package mockstore

import (
	"context"
	"time"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// MockArticleStore 是ArticleStore的mock实现
type MockArticleStore struct{}

// 确保MockArticleStore实现了ArticleStore接口
var _ store.ArticleStore = (*MockArticleStore)(nil)

func (m *MockArticleStore) Create(ctx context.Context, obj *model.ArticleM) error {
	// 设置必要的字段
	obj.ID = 1
	if obj.Slug == "" {
		obj.Slug = "test-article"
	}
	now := time.Now()
	obj.CreatedAt = &now
	obj.UpdatedAt = &now
	return nil
}

func (m *MockArticleStore) Update(ctx context.Context, obj *model.ArticleM) error {
	return nil
}

func (m *MockArticleStore) Delete(ctx context.Context, condition interface{}) error {
	return nil
}

func (m *MockArticleStore) Get(ctx context.Context, condition interface{}) (*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return &model.ArticleM{
		ID:          1,
		Title:       "Test Article",
		Slug:        "test-article",
		Body:        &body,
		Description: &description,
		AuthorID:    1,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}, nil
}

func (m *MockArticleStore) List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

// ArticleExpansion接口的实现
func (m *MockArticleStore) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    authorID,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

func (m *MockArticleStore) GetBySlug(ctx context.Context, slug string) (*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return &model.ArticleM{
		ID:          1,
		Title:       "Test Article",
		Slug:        slug,
		Body:        &body,
		Description: &description,
		AuthorID:    1,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}, nil
}

func (m *MockArticleStore) Favorite(ctx context.Context, articleID, userID int64) error {
	return nil
}

func (m *MockArticleStore) Unfavorite(ctx context.Context, articleID, userID int64) error {
	return nil
}

func (m *MockArticleStore) IsFavorited(ctx context.Context, articleID, userID int64) (bool, error) {
	return false, nil
}

func (m *MockArticleStore) GetFavoritesCount(ctx context.Context, articleID int64) (int64, error) {
	return 0, nil
}

func (m *MockArticleStore) GetByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

func (m *MockArticleStore) GetByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

func (m *MockArticleStore) GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    2, // Different author than the current user
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

func (m *MockArticleStore) GetByComplexCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}

func (m *MockArticleStore) ListWithCursor(ctx context.Context, condition interface{}, cursor int64, limit int) (bool, []*model.ArticleM, int64, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return false, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, 0, nil
}

func (m *MockArticleStore) ListWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (bool, []*model.ArticleM, int64, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return false, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, 0, nil
}

func (m *MockArticleStore) ListWithDeferredJoin(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "This is a test article"
	description := "Test article description"
	now := time.Now()
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test Article",
			Slug:        "test-article",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
	}, nil
}
