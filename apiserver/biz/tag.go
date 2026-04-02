package biz

import (
	"context"
	"errors"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// TagBiz 定义了标签业务需要实现的方法
type TagBiz interface {
	// 创建标签
	CreateTag(ctx context.Context, tag *model.TagM) (*model.TagM, error)
	// 根据ID获取标签
	GetTagByID(ctx context.Context, tagID int64) (*model.TagM, error)
	// 根据标签名获取标签
	GetTagByTag(ctx context.Context, tag string) (*model.TagM, error)
	// 获取所有标签
	GetAllTags(ctx context.Context) ([]*model.TagM, error)
	// 获取标签列表
	GetTags(ctx context.Context, offset, limit int) (int64, []*model.TagM, error)
}

// tagBiz 是 TagBiz 接口的实现
type tagBiz struct {
	store store.IStore
}

// newTagBiz 创建一个 tagBiz 实例
func newTagBiz(store store.IStore) *tagBiz {
	return &tagBiz{
		store: store,
	}
}

// CreateTag 创建标签
func (b *tagBiz) CreateTag(ctx context.Context, tag *model.TagM) (*model.TagM, error) {
	// 检查标签是否已存在
	existingTag, err := b.store.Tag().GetByTag(ctx, tag.Tag)
	if err == nil && existingTag != nil {
		return existingTag, nil
	}

	// 创建标签
	err = b.store.Tag().Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetTagByID 根据ID获取标签
func (b *tagBiz) GetTagByID(ctx context.Context, tagID int64) (*model.TagM, error) {
	tag, err := b.store.Tag().Get(ctx, map[string]interface{}{"id": tagID})
	if err != nil {
		return nil, errors.New("tag not found")
	}

	return tag, nil
}

// GetTagByTag 根据标签名获取标签
func (b *tagBiz) GetTagByTag(ctx context.Context, tag string) (*model.TagM, error) {
	return b.store.Tag().GetByTag(ctx, tag)
}

// GetAllTags 获取所有标签
func (b *tagBiz) GetAllTags(ctx context.Context) ([]*model.TagM, error) {
	return b.store.Tag().GetAll(ctx)
}

// GetTags 获取标签列表
func (b *tagBiz) GetTags(ctx context.Context, offset, limit int) (int64, []*model.TagM, error) {
	return b.store.Tag().List(ctx, nil, offset, limit)
}
