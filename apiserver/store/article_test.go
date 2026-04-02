package store

import (
	"context"
	"testing"

	"github.com/onexstack/realworld/apiserver/model"
)

type testArticleStore struct{}

func (t *testArticleStore) Create(ctx context.Context, obj *model.ArticleM) error {
	return nil
}

func (t *testArticleStore) Update(ctx context.Context, obj *model.ArticleM) error {
	return nil
}

func (t *testArticleStore) Delete(ctx context.Context, condition interface{}) error {
	return nil
}

func (t *testArticleStore) Get(ctx context.Context, condition interface{}) (*model.ArticleM, error) {
	return nil, nil
}

func (t *testArticleStore) List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	return 0, nil, nil
}

func (t *testArticleStore) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	return 0, nil, nil
}

func (t *testArticleStore) GetBySlug(ctx context.Context, slug string) (*model.ArticleM, error) {
	return nil, nil
}

func (t *testArticleStore) Favorite(ctx context.Context, articleID, userID int64) error {
	return nil
}

func (t *testArticleStore) Unfavorite(ctx context.Context, articleID, userID int64) error {
	return nil
}

func (t *testArticleStore) IsFavorited(ctx context.Context, articleID, userID int64) (bool, error) {
	return false, nil
}

func (t *testArticleStore) GetFavoritesCount(ctx context.Context, articleID int64) (int64, error) {
	return 0, nil
}

func (t *testArticleStore) GetByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error) {
	return 0, nil, nil
}

func (t *testArticleStore) GetByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error) {
	return 0, nil, nil
}

func (t *testArticleStore) GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	return 0, nil, nil
}

func (t *testArticleStore) GetByComplexCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "test"
	description := "test"
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test",
			Slug:        "test",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
		},
	}, nil
}

func (t *testArticleStore) ListWithCursor(ctx context.Context, condition interface{}, cursor int64, limit int) (bool, []*model.ArticleM, int64, error) {
	body := "test"
	description := "test"
	return false, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test",
			Slug:        "test",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
		},
	}, 0, nil
}

func (t *testArticleStore) ListWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (bool, []*model.ArticleM, int64, error) {
	body := "test"
	description := "test"
	return false, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test",
			Slug:        "test",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
		},
	}, 0, nil
}

func (t *testArticleStore) ListWithDeferredJoin(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	body := "test"
	description := "test"
	return 1, []*model.ArticleM{
		{
			ID:          1,
			Title:       "Test",
			Slug:        "test",
			Body:        &body,
			Description: &description,
			AuthorID:    1,
		},
	}, nil
}

func TestArticleStoreInterface(t *testing.T) {
	t.Parallel()

	var store ArticleStore = &testArticleStore{}
	ctx := context.Background()

	t.Run("GetByComplexCondition", func(t *testing.T) {
		conditions := map[string]interface{}{}
		count, articles, err := store.GetByComplexCondition(ctx, conditions, 0, 20)
		if err != nil {
			t.Errorf("GetByComplexCondition() error = %v", err)
		}
		if count != 1 {
			t.Errorf("GetByComplexCondition() count = %v, want 1", count)
		}
		if len(articles) != 1 {
			t.Errorf("GetByComplexCondition() articles length = %v, want 1", len(articles))
		}
	})

	t.Run("ListWithCursor", func(t *testing.T) {
		hasMore, articles, nextCursor, err := store.ListWithCursor(ctx, nil, 0, 20)
		if err != nil {
			t.Errorf("ListWithCursor() error = %v", err)
		}
		if hasMore != false {
			t.Errorf("ListWithCursor() hasMore = %v, want false", hasMore)
		}
		if len(articles) != 1 {
			t.Errorf("ListWithCursor() articles length = %v, want 1", len(articles))
		}
		if nextCursor < 0 {
			t.Errorf("ListWithCursor() nextCursor = %v, want >= 0", nextCursor)
		}
	})

	t.Run("ListWithCursorAndCondition", func(t *testing.T) {
		conditions := map[string]interface{}{"tag": "test"}
		hasMore, articles, nextCursor, err := store.ListWithCursorAndCondition(ctx, conditions, 0, 20)
		if err != nil {
			t.Errorf("ListWithCursorAndCondition() error = %v", err)
		}
		if hasMore != false {
			t.Errorf("ListWithCursorAndCondition() hasMore = %v, want false", hasMore)
		}
		if len(articles) != 1 {
			t.Errorf("ListWithCursorAndCondition() articles length = %v, want 1", len(articles))
		}
		if nextCursor < 0 {
			t.Errorf("ListWithCursorAndCondition() nextCursor = %v, want >= 0", nextCursor)
		}
	})
}
