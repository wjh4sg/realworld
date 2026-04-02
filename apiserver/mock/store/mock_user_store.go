package mockstore

import (
	"context"
	"errors"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

type MockUserStore struct{}

var _ store.UserStore = (*MockUserStore)(nil)

func (m *MockUserStore) Create(ctx context.Context, obj *model.UserM) error {
	obj.ID = 1
	return nil
}

func (m *MockUserStore) Update(ctx context.Context, obj *model.UserM) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, condition interface{}) error {
	return nil
}

func (m *MockUserStore) Get(ctx context.Context, condition interface{}) (*model.UserM, error) {
	return &model.UserM{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}, nil
}

func (m *MockUserStore) List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.UserM, error) {
	return 0, []*model.UserM{}, nil
}

func (m *MockUserStore) GetByUsername(ctx context.Context, username string) (*model.UserM, error) {
	return nil, errors.New("user not found")
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*model.UserM, error) {
	return &model.UserM{
		ID:       1,
		Username: "testuser",
		Email:    email,
		Password: "password123",
		Bio:      nil,
		Image:    nil,
	}, nil
}

func (m *MockUserStore) FindOneUser(ctx context.Context, condition interface{}) (*model.UserM, error) {
	return &model.UserM{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Bio:      nil,
		Image:    nil,
	}, nil
}

func (m *MockUserStore) SetPassword(ctx context.Context, userID int64, password string) error {
	return nil
}

func (m *MockUserStore) CheckPassword(ctx context.Context, userID int64, password string) error {
	if password == "password123" {
		return nil
	}
	return errors.New("invalid password")
}

func (m *MockUserStore) Follow(ctx context.Context, followingID, followedByID int64) error {
	return nil
}

func (m *MockUserStore) Unfollow(ctx context.Context, followingID, followedByID int64) error {
	return nil
}

func (m *MockUserStore) IsFollowing(ctx context.Context, followingID, followedByID int64) (bool, error) {
	return false, nil
}

func (m *MockUserStore) GetFollowings(ctx context.Context, userID int64) ([]*model.UserM, error) {
	return []*model.UserM{}, nil
}

func (m *MockUserStore) GetFollowers(ctx context.Context, userID int64) ([]*model.UserM, error) {
	return []*model.UserM{}, nil
}
