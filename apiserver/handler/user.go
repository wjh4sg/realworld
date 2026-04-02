package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/jwt"
	"github.com/onexstack/realworld/apiserver/model"
)

type userHandler struct {
	biz        biz.IBiz
	jwtManager *jwt.Manager
}

func newUserHandler(biz biz.IBiz, jwtManager *jwt.Manager) *userHandler {
	return &userHandler{
		biz:        biz,
		jwtManager: jwtManager,
	}
}

type RegisterRequest struct {
	User struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	} `json:"user" binding:"required"`
}

type LoginRequest struct {
	User struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	} `json:"user" binding:"required"`
}

type UpdateUserRequest struct {
	User struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Bio      *string `json:"bio"`
		Image    *string `json:"image"`
	} `json:"user" binding:"required"`
}

type UserResponse struct {
	User struct {
		Username string  `json:"username"`
		Email    string  `json:"email"`
		Token    string  `json:"token"`
		Bio      *string `json:"bio"`
		Image    *string `json:"image"`
	} `json:"user"`
}

type ProfileResponse struct {
	Profile struct {
		Username  string  `json:"username"`
		Bio       *string `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"profile"`
}

func (h *userHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": map[string]string{"body": "Invalid request format: " + err.Error()}})
		return
	}

	user, err := h.biz.User().Register(c.Request.Context(), req.User.Username, req.User.Email, req.User.Password)
	if err != nil {
		if errors.Is(err, biz.ErrUsernameAlreadyExists) || errors.Is(err, biz.ErrEmailAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"errors": map[string]string{"body": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	resp, refreshToken, err := h.buildAuthenticatedUserResponse(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.Header("X-Refresh-Token", refreshToken)
	c.JSON(http.StatusCreated, resp)
}

func (h *userHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.biz.User().Login(c.Request.Context(), req.User.Email, req.User.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	resp, refreshToken, err := h.buildAuthenticatedUserResponse(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.Header("X-Refresh-Token", refreshToken)
	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.biz.User().GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	resp, refreshToken, err := h.buildAuthenticatedUserResponse(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.Header("X-Refresh-Token", refreshToken)
	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) UpdateUser(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := make(map[string]interface{})
	if req.User.Username != nil {
		update["username"] = *req.User.Username
	}
	if req.User.Email != nil {
		update["email"] = *req.User.Email
	}
	if req.User.Password != nil {
		update["password"] = *req.User.Password
	}
	if req.User.Bio != nil {
		update["bio"] = req.User.Bio
	}
	if req.User.Image != nil {
		update["image"] = req.User.Image
	}

	user, err := h.biz.User().UpdateUser(c.Request.Context(), userID, update)
	if err != nil {
		if errors.Is(err, biz.ErrUsernameAlreadyExists) || errors.Is(err, biz.ErrEmailAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, refreshToken, err := h.buildAuthenticatedUserResponse(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.Header("X-Refresh-Token", refreshToken)
	c.JSON(http.StatusOK, resp)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *userHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userID, err := h.jwtManager.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	newToken, err := h.jwtManager.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
		return
	}

	newRefreshToken, err := h.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	})
}

func (h *userHandler) GetProfile(c *gin.Context) {
	username := c.Param("username")

	user, err := h.biz.User().GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	following := false
	if userID, exists := currentUserID(c); exists {
		following, _ = h.biz.User().IsFollowing(c.Request.Context(), userID, user.ID)
	}

	c.JSON(http.StatusOK, buildProfileResponse(user, following))
}

func (h *userHandler) FollowUser(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	username := c.Param("username")
	targetUser, err := h.biz.User().GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := h.biz.User().FollowUser(c.Request.Context(), userID, targetUser.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, buildProfileResponse(targetUser, true))
}

func (h *userHandler) UnfollowUser(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	username := c.Param("username")
	targetUser, err := h.biz.User().GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := h.biz.User().UnfollowUser(c.Request.Context(), userID, targetUser.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, buildProfileResponse(targetUser, false))
}

func (h *userHandler) buildAuthenticatedUserResponse(user *model.UserM) (UserResponse, string, error) {
	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return UserResponse{}, "", err
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return UserResponse{}, "", err
	}

	return buildUserResponse(user, token), refreshToken, nil
}
