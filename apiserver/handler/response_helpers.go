package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/model"
)

const apiTimeLayout = "2006-01-02T15:04:05.000Z"

func formatAPITime(ts *time.Time) string {
	if ts == nil {
		return ""
	}

	return ts.UTC().Format(apiTimeLayout)
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	if !ok {
		return 0, false
	}

	return userID, true
}

func buildUserResponse(user *model.UserM, token string) UserResponse {
	resp := UserResponse{}
	resp.User.Username = user.Username
	resp.User.Email = user.Email
	resp.User.Token = token
	resp.User.Bio = user.Bio
	resp.User.Image = user.Image
	return resp
}

func buildProfileResponse(user *model.UserM, following bool) ProfileResponse {
	resp := ProfileResponse{}
	resp.Profile.Username = user.Username
	resp.Profile.Bio = user.Bio
	resp.Profile.Image = user.Image
	resp.Profile.Following = following
	return resp
}
