package biz

import (
	"context"
	"errors"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// CommentBiz 定义了评论业务需要实现的方法
type CommentBiz interface {
	// 创建评论
	CreateComment(ctx context.Context, comment *model.CommentM) (*model.CommentM, error)
	// 删除评论
	DeleteComment(ctx context.Context, commentID int64) error
	// 根据ID获取评论
	GetCommentByID(ctx context.Context, commentID int64) (*model.CommentM, error)
	// 根据文章ID获取评论列表
	GetCommentsByArticleID(ctx context.Context, articleID int64, offset, limit int) (int64, []*model.CommentM, error)
	// 根据作者ID获取评论列表
	GetCommentsByAuthorID(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.CommentM, error)
}

// commentBiz 是 CommentBiz 接口的实现
type commentBiz struct {
	store store.IStore
}

// newCommentBiz 创建一个 commentBiz 实例
func newCommentBiz(store store.IStore) *commentBiz {
	return &commentBiz{
		store: store,
	}
}

// CreateComment 创建评论
func (b *commentBiz) CreateComment(ctx context.Context, comment *model.CommentM) (*model.CommentM, error) {
	// 检查文章是否存在
	_, err := b.store.Article().Get(ctx, map[string]interface{}{"id": comment.ArticleID})
	if err != nil {
		return nil, errors.New("article not found")
	}

	// 检查用户是否存在
	_, err = b.store.User().Get(ctx, map[string]interface{}{"id": comment.AuthorID})
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 创建评论
	err = b.store.Comment().Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment 删除评论
func (b *commentBiz) DeleteComment(ctx context.Context, commentID int64) error {
	// 检查评论是否存在
	_, err := b.store.Comment().Get(ctx, map[string]interface{}{"id": commentID})
	if err != nil {
		return errors.New("comment not found")
	}

	// 删除评论
	return b.store.Comment().Delete(ctx, map[string]interface{}{"id": commentID})
}

// GetCommentByID 根据ID获取评论
func (b *commentBiz) GetCommentByID(ctx context.Context, commentID int64) (*model.CommentM, error) {
	return b.store.Comment().Get(ctx, map[string]interface{}{"id": commentID})
}

// GetCommentsByArticleID 根据文章ID获取评论列表
func (b *commentBiz) GetCommentsByArticleID(ctx context.Context, articleID int64, offset, limit int) (int64, []*model.CommentM, error) {
	// 检查文章是否存在
	_, err := b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
	if err != nil {
		return 0, nil, errors.New("article not found")
	}

	return b.store.Comment().GetByArticle(ctx, articleID, offset, limit)
}

// GetCommentsByAuthorID 根据作者ID获取评论列表
func (b *commentBiz) GetCommentsByAuthorID(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.CommentM, error) {
	// 检查用户是否存在
	_, err := b.store.User().Get(ctx, map[string]interface{}{"id": authorID})
	if err != nil {
		return 0, nil, errors.New("user not found")
	}

	return b.store.Comment().GetByAuthor(ctx, authorID, offset, limit)
}
