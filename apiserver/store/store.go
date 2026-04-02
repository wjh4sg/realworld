package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/onexstack/realworld/apiserver/cache"
)

// IStore 定义了 Store 层需要实现的方法（简化版，移除了事务相关方法）
type IStore interface {
	// 返回 Store 层的 *gorm.DB 实例
	DB(ctx context.Context) *gorm.DB
	// 返回缓存实例
	Cache() cache.ICache

	User() UserStore
	Article() ArticleStore
	Comment() CommentStore
	Tag() TagStore
}

// datastore 是 IStore 的具体实现(简化版)
type datastore struct {
	core  *gorm.DB
	cache cache.ICache
}

// 确保 datastore 实现了 IStore 接口
var _ IStore = (*datastore)(nil)

// NewStore 创建一个 IStore 类型的实例(简化版,移除了单例逻辑)
func NewStore(db *gorm.DB, cacheClient cache.ICache) *datastore {
	return &datastore{
		core:  db,
		cache: cacheClient,
	}
}

// DB 根据传入的条件（wheres）对数据库实例进行筛选（简化版，移除了事务上下文处理）
func (store *datastore) DB(ctx context.Context) *gorm.DB {
	return store.core // 直接返回基础数据库连接
}

// Cache 返回缓存实例
func (store *datastore) Cache() cache.ICache {
	return store.cache
}

// User 返回一个实现了 UserStore 接口的实例
func (store *datastore) User() UserStore {
	return newUserStore(store)
}

// Article 返回一个实现了 ArticleStore 接口的实例
func (store *datastore) Article() ArticleStore {
	return newArticleStore(store)
}

// Comment 返回一个实现了 CommentStore 接口的实例
func (store *datastore) Comment() CommentStore {
	return newCommentStore(store)
}

// Tag 返回一个实现了 TagStore 接口的实例
func (store *datastore) Tag() TagStore {
	return newTagStore(store)
}
