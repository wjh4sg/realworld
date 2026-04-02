package biz

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrAlreadyFollowing      = errors.New("already following this user")
	ErrNotFollowing          = errors.New("not following this user")
)
