package store

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/onexstack/realworld/apiserver/model"
)

// CommentStore 定义了 comment 模块在 store 层所实现的方法.
type CommentStore interface {
	Create(ctx context.Context, obj *model.CommentM) error
	Update(ctx context.Context, obj *model.CommentM) error
	Delete(ctx context.Context, condition interface{}) error
	Get(ctx context.Context, condition interface{}) (*model.CommentM, error)
	List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.CommentM, error)

	CommentExpansion
}

// CommentExpansion 定义了评论操作的附加方法.
type CommentExpansion interface {
	// 根据文章ID获取评论列表
	GetByArticle(ctx context.Context, articleID int64, offset, limit int) (int64, []*model.CommentM, error)
	// 根据作者ID获取评论列表
	GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.CommentM, error)
}

// commentStore 是 CommentStore 接口的实现.
type commentStore struct {
	store *datastore
}

// 确保 commentStore 实现了 CommentStore 接口.
var _ CommentStore = (*commentStore)(nil)

// newCommentStore 创建 commentStore 的实例.
func newCommentStore(store *datastore) *commentStore {
	return &commentStore{store}
}

// Create 插入一条评论记录.
func (s *commentStore) Create(ctx context.Context, obj *model.CommentM) error {
	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		slog.Error("Failed to insert comment into database", "err", err, "comment", obj)
		return errors.New("failed to insert comment: " + err.Error())
	}

	return nil
}

// Update 更新评论数据库记录.
func (s *commentStore) Update(ctx context.Context, obj *model.CommentM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		slog.Error("Failed to update comment in database", "err", err, "comment", obj)
		return errors.New("failed to update comment: " + err.Error())
	}

	return nil
}

// Delete 根据条件删除评论记录.
func (s *commentStore) Delete(ctx context.Context, condition interface{}) error {
	err := s.store.DB(ctx).Where(condition).Delete(new(model.CommentM)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to delete comment from database", "err", err, "condition", condition)
		return errors.New("failed to delete comment: " + err.Error())
	}

	return nil
}

// Get 根据条件查询评论记录.
func (s *commentStore) Get(ctx context.Context, condition interface{}) (*model.CommentM, error) {
	var obj model.CommentM
	if err := s.store.DB(ctx).Where(condition).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		slog.Error("Failed to retrieve comment from database", "err", err, "condition", condition)
		return nil, errors.New("failed to get comment: " + err.Error())
	}

	return &obj, nil
}

// List 返回评论列表和总数.
func (s *commentStore) List(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.CommentM, err error) {
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	// 先获取总数
	if err = db.Model(&model.CommentM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count comments from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to count comments: " + err.Error())
	}

	// 再获取列表
	if err = db.Offset(offset).Limit(limit).Order("created_at desc").Find(&ret).Error; err != nil {
		slog.Error("Failed to list comments from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to list comments: " + err.Error())
	}

	return
}

// GetByArticle 根据文章ID获取评论列表
func (s *commentStore) GetByArticle(ctx context.Context, articleID int64, offset, limit int) (int64, []*model.CommentM, error) {
	condition := map[string]interface{}{"article_id": articleID}
	return s.List(ctx, condition, offset, limit)
}

// GetByAuthor 根据作者ID获取评论列表
func (s *commentStore) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.CommentM, error) {
	condition := map[string]interface{}{"author_id": authorID}
	return s.List(ctx, condition, offset, limit)
}
