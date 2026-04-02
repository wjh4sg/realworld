package biz

import (
	"context"
	"errors"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// UserBiz defines user business behavior.
type UserBiz interface {
	Register(ctx context.Context, username, email, password string) (*model.UserM, error)
	Login(ctx context.Context, email, password string) (*model.UserM, error)
	UpdateUser(ctx context.Context, userID int64, update map[string]interface{}) (*model.UserM, error)
	GetUser(ctx context.Context, userID int64) (*model.UserM, error)
	GetUserByUsername(ctx context.Context, username string) (*model.UserM, error)
	FollowUser(ctx context.Context, userID, followID int64) error
	UnfollowUser(ctx context.Context, userID, followID int64) error
	IsFollowing(ctx context.Context, userID, followID int64) (bool, error)
	GetFollowings(ctx context.Context, userID int64) ([]*model.UserM, error)
	GetFollowers(ctx context.Context, userID int64) ([]*model.UserM, error)
}

type userBiz struct {
	store store.IStore
}

func newUserBiz(store store.IStore) *userBiz {
	return &userBiz{store: store}
}

func (b *userBiz) Register(ctx context.Context, username, email, password string) (*model.UserM, error) {
	existingUser, err := b.store.User().GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	} else if err != nil && !isUserNotFound(err) {
		return nil, err
	}

	existingUser, err = b.store.User().GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	} else if err != nil && !isUserNotFound(err) {
		return nil, err
	}

	user := &model.UserM{
		Username: username,
		Email:    email,
		Password: password,
	}

	if err := b.store.User().Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *userBiz) Login(ctx context.Context, email, password string) (*model.UserM, error) {
	user, err := b.store.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := b.store.User().CheckPassword(ctx, user.ID, password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (b *userBiz) UpdateUser(ctx context.Context, userID int64, update map[string]interface{}) (*model.UserM, error) {
	user, err := b.store.User().Get(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return nil, ErrUserNotFound
	}

	if username, ok := update["username"].(string); ok {
		existingUser, existingErr := b.store.User().GetByUsername(ctx, username)
		if existingErr == nil && existingUser != nil && existingUser.ID != userID {
			return nil, ErrUsernameAlreadyExists
		} else if existingErr != nil && !isUserNotFound(existingErr) {
			return nil, existingErr
		}
		user.Username = username
	}

	if email, ok := update["email"].(string); ok {
		existingUser, existingErr := b.store.User().GetByEmail(ctx, email)
		if existingErr == nil && existingUser != nil && existingUser.ID != userID {
			return nil, ErrEmailAlreadyExists
		} else if existingErr != nil && !isUserNotFound(existingErr) {
			return nil, existingErr
		}
		user.Email = email
	}

	if bio, ok := update["bio"].(*string); ok {
		user.Bio = bio
	}

	if image, ok := update["image"].(*string); ok {
		user.Image = image
	}

	if password, ok := update["password"].(string); ok {
		if err := b.store.User().SetPassword(ctx, userID, password); err != nil {
			return nil, err
		}
		delete(update, "password")
	}

	if err := b.store.User().Update(ctx, user); err != nil {
		return nil, err
	}

	return b.store.User().Get(ctx, map[string]interface{}{"id": userID})
}

func (b *userBiz) GetUser(ctx context.Context, userID int64) (*model.UserM, error) {
	return b.store.User().Get(ctx, map[string]interface{}{"id": userID})
}

func (b *userBiz) GetUserByUsername(ctx context.Context, username string) (*model.UserM, error) {
	return b.store.User().GetByUsername(ctx, username)
}

func (b *userBiz) FollowUser(ctx context.Context, userID, followID int64) error {
	isFollowing, err := b.store.User().IsFollowing(ctx, followID, userID)
	if err != nil {
		return err
	}

	if isFollowing {
		return ErrAlreadyFollowing
	}

	return b.store.User().Follow(ctx, followID, userID)
}

func (b *userBiz) UnfollowUser(ctx context.Context, userID, followID int64) error {
	isFollowing, err := b.store.User().IsFollowing(ctx, followID, userID)
	if err != nil {
		return err
	}

	if !isFollowing {
		return ErrNotFollowing
	}

	return b.store.User().Unfollow(ctx, followID, userID)
}

func (b *userBiz) IsFollowing(ctx context.Context, userID, followID int64) (bool, error) {
	return b.store.User().IsFollowing(ctx, followID, userID)
}

func (b *userBiz) GetFollowings(ctx context.Context, userID int64) ([]*model.UserM, error) {
	return b.store.User().GetFollowings(ctx, userID)
}

func (b *userBiz) GetFollowers(ctx context.Context, userID int64) ([]*model.UserM, error) {
	return b.store.User().GetFollowers(ctx, userID)
}

func isUserNotFound(err error) bool {
	return err != nil && err.Error() == "user not found"
}
