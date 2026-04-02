package mockstore

import (
	"context"
	"time"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// MockCommentStore 是CommentStore的mock实现
type MockCommentStore struct{}

// 确保MockCommentStore实现了CommentStore接口
var _ store.CommentStore = (*MockCommentStore)(nil)

func (m *MockCommentStore) Create(ctx context.Context, obj *model.CommentM) error {
	return nil
}

func (m *MockCommentStore) Update(ctx context.Context, obj *model.CommentM) error {
	return nil
}

func (m *MockCommentStore) Delete(ctx context.Context, condition interface{}) error {
	return nil
}

func (m *MockCommentStore) Get(ctx context.Context, condition interface{}) (*model.CommentM, error) {
	now := time.Now()
	return &model.CommentM{
		ID:        1,
		Body:      "Test comment",
		ArticleID: 1,
		AuthorID:  1,
		CreatedAt: &now,
		UpdatedAt: &now,
	}, nil
}

func (m *MockCommentStore) List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.CommentM, error) {
	return 0, []*model.CommentM{}, nil
}

// CommentExpansion接口的实现
func (m *MockCommentStore) GetByArticle(ctx context.Context, articleID int64, offset, limit int) (int64, []*model.CommentM, error) {
	return 0, []*model.CommentM{}, nil
}

func (m *MockCommentStore) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.CommentM, error) {
	return 0, []*model.CommentM{}, nil
}
