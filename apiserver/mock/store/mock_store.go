package mockstore

import (
	"context"

	"github.com/onexstack/realworld/apiserver/cache"
	"github.com/onexstack/realworld/apiserver/store"
	"gorm.io/gorm"
)

// MockStore 是IStore的简单实现，用于测试和mock模式
type MockStore struct {
	userStore    *MockUserStore
	articleStore *MockArticleStore
	commentStore *MockCommentStore
	tagStore     *MockTagStore
}

// 确保MockStore实现了IStore接口
var _ store.IStore = (*MockStore)(nil)

// NewMockStore 初始化MockStore
func NewMockStore() *MockStore {
	return &MockStore{
		userStore:    &MockUserStore{},
		articleStore: &MockArticleStore{},
		commentStore: &MockCommentStore{},
		tagStore:     &MockTagStore{},
	}
}

// DB 返回nil，因为mock模式不需要真实的数据库连接
func (m *MockStore) DB(ctx context.Context) *gorm.DB {
	return nil
}

// User 返回mock的userStore
func (m *MockStore) User() store.UserStore {
	return m.userStore
}

// Article 返回mock的articleStore
func (m *MockStore) Article() store.ArticleStore {
	return m.articleStore
}

// Comment 返回mock的commentStore
func (m *MockStore) Comment() store.CommentStore {
	return m.commentStore
}

// Tag 返回mock的tagStore
func (m *MockStore) Tag() store.TagStore {
	return m.tagStore
}

// Cache 返回nil，因为mock模式不需要真实的缓存连接
func (m *MockStore) Cache() cache.ICache {
	return nil
}
