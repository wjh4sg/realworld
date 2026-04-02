package store

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/onexstack/realworld/apiserver/model"
)

// TagStore 定义了 tag 模块在 store 层所实现的方法.
type TagStore interface {
	Create(ctx context.Context, obj *model.TagM) error
	Update(ctx context.Context, obj *model.TagM) error
	Delete(ctx context.Context, condition interface{}) error
	Get(ctx context.Context, condition interface{}) (*model.TagM, error)
	List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.TagM, error)

	TagExpansion
}

// TagExpansion 定义了标签操作的附加方法.
type TagExpansion interface {
	// 根据标签名获取标签
	GetByTag(ctx context.Context, tag string) (*model.TagM, error)
	// 获取所有标签
	GetAll(ctx context.Context) ([]*model.TagM, error)
}

// tagStore 是 TagStore 接口的实现.
type tagStore struct {
	store *datastore
}

// 确保 tagStore 实现了 TagStore 接口.
var _ TagStore = (*tagStore)(nil)

// newTagStore 创建 tagStore 的实例.
func newTagStore(store *datastore) *tagStore {
	return &tagStore{store}
}

// Create 插入一条标签记录.
func (s *tagStore) Create(ctx context.Context, obj *model.TagM) error {
	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		slog.Error("Failed to insert tag into database", "err", err, "tag", obj)
		return errors.New("failed to insert tag: " + err.Error())
	}

	return nil
}

// Update 更新标签数据库记录.
func (s *tagStore) Update(ctx context.Context, obj *model.TagM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		slog.Error("Failed to update tag in database", "err", err, "tag", obj)
		return errors.New("failed to update tag: " + err.Error())
	}

	return nil
}

// Delete 根据条件删除标签记录.
func (s *tagStore) Delete(ctx context.Context, condition interface{}) error {
	err := s.store.DB(ctx).Where(condition).Delete(new(model.TagM)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to delete tag from database", "err", err, "condition", condition)
		return errors.New("failed to delete tag: " + err.Error())
	}

	return nil
}

// Get 根据条件查询标签记录.
func (s *tagStore) Get(ctx context.Context, condition interface{}) (*model.TagM, error) {
	var obj model.TagM
	if err := s.store.DB(ctx).Where(condition).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		slog.Error("Failed to retrieve tag from database", "err", err, "condition", condition)
		return nil, errors.New("failed to get tag: " + err.Error())
	}

	return &obj, nil
}

// List 返回标签列表和总数.
func (s *tagStore) List(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.TagM, err error) {
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	// 先获取总数
	if err = db.Model(&model.TagM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count tags from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to count tags: " + err.Error())
	}

	// 再获取列表
	if err = db.Offset(offset).Limit(limit).Order("tag asc").Find(&ret).Error; err != nil {
		slog.Error("Failed to list tags from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to list tags: " + err.Error())
	}

	return
}

// GetByTag 根据标签名获取标签
func (s *tagStore) GetByTag(ctx context.Context, tag string) (*model.TagM, error) {
	var obj model.TagM
	if err := s.store.DB(ctx).Where("tag = ?", tag).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		slog.Error("Failed to get tag by tag name", "err", err, "tag", tag)
		return nil, errors.New("failed to get tag by name: " + err.Error())
	}
	return &obj, nil
}

// GetAll 获取所有标签
func (s *tagStore) GetAll(ctx context.Context) ([]*model.TagM, error) {
	var tags []*model.TagM
	if err := s.store.DB(ctx).Order("tag asc").Find(&tags).Error; err != nil {
		slog.Error("Failed to get all tags", "err", err)
		return nil, errors.New("failed to get all tags: " + err.Error())
	}
	return tags, nil
}
