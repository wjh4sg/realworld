package mockstore

import (
	"context"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// MockTagStore 是TagStore的mock实现
type MockTagStore struct{}

// 确保MockTagStore实现了TagStore接口
var _ store.TagStore = (*MockTagStore)(nil)

func (m *MockTagStore) Create(ctx context.Context, obj *model.TagM) error {
	return nil
}

func (m *MockTagStore) Update(ctx context.Context, obj *model.TagM) error {
	return nil
}

func (m *MockTagStore) Delete(ctx context.Context, condition interface{}) error {
	return nil
}

func (m *MockTagStore) Get(ctx context.Context, condition interface{}) (*model.TagM, error) {
	return &model.TagM{
		ID:  1,
		Tag: "test",
	}, nil
}

func (m *MockTagStore) List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.TagM, error) {
	return 0, []*model.TagM{}, nil
}

// TagExpansion接口的实现
func (m *MockTagStore) GetByTag(ctx context.Context, tag string) (*model.TagM, error) {
	return &model.TagM{
		ID:  1,
		Tag: tag,
	}, nil
}

func (m *MockTagStore) GetAll(ctx context.Context) ([]*model.TagM, error) {
	return []*model.TagM{
		{ID: 1, Tag: "test"},
		{ID: 2, Tag: "example"},
	}, nil
}
