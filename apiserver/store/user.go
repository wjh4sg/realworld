package store

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/onexstack/realworld/apiserver/model"
)

type UserStore interface {
	Create(ctx context.Context, obj *model.UserM) error
	Update(ctx context.Context, obj *model.UserM) error
	Delete(ctx context.Context, condition interface{}) error
	Get(ctx context.Context, condition interface{}) (*model.UserM, error)
	List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.UserM, error)
	GetByUsername(ctx context.Context, username string) (*model.UserM, error)
	GetByEmail(ctx context.Context, email string) (*model.UserM, error)
	FindOneUser(ctx context.Context, condition interface{}) (*model.UserM, error)
	SetPassword(ctx context.Context, userID int64, password string) error
	CheckPassword(ctx context.Context, userID int64, password string) error
	Follow(ctx context.Context, followingID, followedByID int64) error
	Unfollow(ctx context.Context, followingID, followedByID int64) error
	IsFollowing(ctx context.Context, followingID, followedByID int64) (bool, error)
	GetFollowings(ctx context.Context, userID int64) ([]*model.UserM, error)
	GetFollowers(ctx context.Context, userID int64) ([]*model.UserM, error)
}

type userStore struct {
	store *datastore
}

var _ UserStore = (*userStore)(nil)

func newUserStore(store *datastore) *userStore {
	return &userStore{store: store}
}

func (s *userStore) Create(ctx context.Context, obj *model.UserM) error {
	if obj.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(obj.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Failed to generate password hash", "err", err, "user", obj)
			return errors.New("failed to generate password hash: " + err.Error())
		}
		obj.Password = string(passwordHash)
	}

	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		slog.Error("Failed to insert user into database", "err", err, "user", obj)
		return errors.New("failed to insert user: " + err.Error())
	}

	return nil
}

func (s *userStore) Update(ctx context.Context, obj *model.UserM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		slog.Error("Failed to update user in database", "err", err, "user", obj)
		return errors.New("failed to update user: " + err.Error())
	}

	return nil
}

func (s *userStore) Delete(ctx context.Context, condition interface{}) error {
	err := s.store.DB(ctx).Where(condition).Delete(new(model.UserM)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to delete user from database", "err", err, "condition", condition)
		return errors.New("failed to delete user: " + err.Error())
	}

	return nil
}

func (s *userStore) Get(ctx context.Context, condition interface{}) (*model.UserM, error) {
	var obj model.UserM
	if err := s.store.DB(ctx).Where(condition).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		slog.Error("Failed to retrieve user from database", "err", err, "condition", condition)
		return nil, errors.New("failed to get user: " + err.Error())
	}

	return &obj, nil
}

func (s *userStore) List(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.UserM, err error) {
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	if err = db.Model(&model.UserM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count users from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to count users: " + err.Error())
	}

	if err = db.Offset(offset).Limit(limit).Order("id desc").Find(&ret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return count, []*model.UserM{}, nil
		}
		slog.Error("Failed to list users from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to list users: " + err.Error())
	}

	return count, ret, nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*model.UserM, error) {
	var user model.UserM
	if err := s.store.DB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		slog.Error("Failed to get user by username", "err", err, "username", username)
		return nil, errors.New("failed to get user by username: " + err.Error())
	}

	return &user, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*model.UserM, error) {
	var user model.UserM
	if err := s.store.DB(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		slog.Error("Failed to get user by email", "err", err, "email", email)
		return nil, errors.New("failed to get user by email: " + err.Error())
	}

	return &user, nil
}

func (s *userStore) FindOneUser(ctx context.Context, condition interface{}) (*model.UserM, error) {
	var user model.UserM
	if err := s.store.DB(ctx).Where(condition).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		slog.Error("Failed to find user by condition", "err", err, "condition", condition)
		return nil, errors.New("failed to find user: " + err.Error())
	}

	return &user, nil
}

func (s *userStore) Follow(ctx context.Context, followingID, followedByID int64) error {
	follow := &model.FollowM{
		FollowingID:  followingID,
		FollowedByID: followedByID,
	}

	var existingFollow model.FollowM
	if err := s.store.DB(ctx).Where("following_id = ? AND followed_by_id = ?", followingID, followedByID).First(&existingFollow).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to check existing follow", "err", err, "following_id", followingID, "followed_by_id", followedByID)
		return errors.New("failed to check existing follow: " + err.Error())
	}

	if err := s.store.DB(ctx).Create(follow).Error; err != nil {
		slog.Error("Failed to create follow", "err", err, "follow", follow)
		return errors.New("failed to create follow: " + err.Error())
	}

	return nil
}

func (s *userStore) Unfollow(ctx context.Context, followingID, followedByID int64) error {
	result := s.store.DB(ctx).Where("following_id = ? AND followed_by_id = ?", followingID, followedByID).Delete(&model.FollowM{})
	if result.Error != nil {
		slog.Error("Failed to delete follow", "err", result.Error, "following_id", followingID, "followed_by_id", followedByID)
		return errors.New("failed to delete follow: " + result.Error.Error())
	}

	return nil
}

func (s *userStore) IsFollowing(ctx context.Context, followingID, followedByID int64) (bool, error) {
	var follow model.FollowM
	err := s.store.DB(ctx).Where("following_id = ? AND followed_by_id = ?", followingID, followedByID).First(&follow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		slog.Error("Failed to check follow", "err", err, "following_id", followingID, "followed_by_id", followedByID)
		return false, errors.New("failed to check follow: " + err.Error())
	}

	return true, nil
}

func (s *userStore) GetFollowings(ctx context.Context, userID int64) ([]*model.UserM, error) {
	var follows []model.FollowM
	if err := s.store.DB(ctx).Where("followed_by_id = ?", userID).Find(&follows).Error; err != nil {
		slog.Error("Failed to get follows", "err", err, "user_id", userID)
		return nil, errors.New("failed to get follows: " + err.Error())
	}

	if len(follows) == 0 {
		return []*model.UserM{}, nil
	}

	followingIDs := make([]int64, len(follows))
	for i, follow := range follows {
		followingIDs[i] = follow.FollowingID
	}

	var users []*model.UserM
	if err := s.store.DB(ctx).Where("id IN ?", followingIDs).Find(&users).Error; err != nil {
		slog.Error("Failed to get following users", "err", err, "following_ids", followingIDs)
		return nil, errors.New("failed to get following users: " + err.Error())
	}

	return users, nil
}

func (s *userStore) GetFollowers(ctx context.Context, userID int64) ([]*model.UserM, error) {
	var follows []model.FollowM
	if err := s.store.DB(ctx).Where("following_id = ?", userID).Find(&follows).Error; err != nil {
		slog.Error("Failed to get followers", "err", err, "user_id", userID)
		return nil, errors.New("failed to get followers: " + err.Error())
	}

	if len(follows) == 0 {
		return []*model.UserM{}, nil
	}

	followerIDs := make([]int64, len(follows))
	for i, follow := range follows {
		followerIDs[i] = follow.FollowedByID
	}

	var users []*model.UserM
	if err := s.store.DB(ctx).Where("id IN ?", followerIDs).Find(&users).Error; err != nil {
		slog.Error("Failed to get follower users", "err", err, "follower_ids", followerIDs)
		return nil, errors.New("failed to get follower users: " + err.Error())
	}

	return users, nil
}

func (s *userStore) SetPassword(ctx context.Context, userID int64, password string) error {
	if password == "" {
		return errors.New("password should not be empty")
	}

	user, err := s.Get(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		slog.Error("Failed to get user for setting password", "err", err, "user_id", userID)
		return errors.New("failed to get user: " + err.Error())
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to generate password hash", "err", err, "user_id", userID)
		return errors.New("failed to generate password hash: " + err.Error())
	}

	user.Password = string(passwordHash)
	if err := s.Update(ctx, user); err != nil {
		slog.Error("Failed to update user password", "err", err, "user_id", userID)
		return errors.New("failed to update password: " + err.Error())
	}

	return nil
}

func (s *userStore) CheckPassword(ctx context.Context, userID int64, password string) error {
	user, err := s.Get(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		slog.Error("Failed to get user for checking password", "err", err, "user_id", userID)
		return errors.New("failed to get user: " + err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return errors.New("invalid password")
	}

	return nil
}
