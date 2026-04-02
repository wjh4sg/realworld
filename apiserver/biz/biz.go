package biz

import (
	"github.com/onexstack/realworld/apiserver/store"
)

// IBiz 定义了业务层需要实现的方法
type IBiz interface {
	// 获取 Store 实例
	Store() store.IStore

	// 用户业务
	User() UserBiz
	// 文章业务
	Article() ArticleBiz
	// 评论业务
	Comment() CommentBiz
	// 标签业务
	Tag() TagBiz
}

// Biz 是 IBiz 接口的实现
type Biz struct {
	store store.IStore

	userBiz    UserBiz
	articleBiz ArticleBiz
	commentBiz CommentBiz
	tagBiz     TagBiz
}

// NewBiz 创建一个 Biz 实例
func NewBiz(store store.IStore) *Biz {
	userBiz := newUserBiz(store)
	articleBiz := newArticleBiz(store)
	commentBiz := newCommentBiz(store)
	tagBiz := newTagBiz(store)

	return &Biz{
		store: store,

		userBiz:    userBiz,
		articleBiz: articleBiz,
		commentBiz: commentBiz,
		tagBiz:     tagBiz,
	}
}

// Store 返回 Store 实例
func (b *Biz) Store() store.IStore {
	return b.store
}

// User 返回 UserBiz 实例
func (b *Biz) User() UserBiz {
	return b.userBiz
}

// Article 返回 ArticleBiz 实例
func (b *Biz) Article() ArticleBiz {
	return b.articleBiz
}

// Comment 返回 CommentBiz 实例
func (b *Biz) Comment() CommentBiz {
	return b.commentBiz
}

// Tag 返回 TagBiz 实例
func (b *Biz) Tag() TagBiz {
	return b.tagBiz
}
