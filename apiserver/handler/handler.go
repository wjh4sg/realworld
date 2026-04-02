package handler

import (
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/jwt"
)

// IHandler 定义了 Handler 层的接口
type IHandler interface {
	// 获取 Biz 实例，与 Biz 层的 Store() 方法设计保持对应
	Biz() biz.IBiz
	// 用户相关
	User() *userHandler
	// 文章相关
	Article() *articleHandler
	// 评论相关
	Comment() *commentHandler
	// 标签相关
	Tag() *tagHandler
}

// Handler 是 IHandler 的实现
type Handler struct {
	biz        biz.IBiz
	jwtManager *jwt.Manager

	userHandler    *userHandler
	articleHandler *articleHandler
	commentHandler *commentHandler
	tagHandler     *tagHandler
}

// NewHandler 创建一个 Handler 实例
func NewHandler(biz biz.IBiz, jwtManager *jwt.Manager) *Handler {
	userHandler := newUserHandler(biz, jwtManager)
	articleHandler := newArticleHandler(biz, jwtManager)
	commentHandler := newCommentHandler(biz)
	tagHandler := newTagHandler(biz)

	return &Handler{
		biz:        biz,
		jwtManager: jwtManager,

		userHandler:    userHandler,
		articleHandler: articleHandler,
		commentHandler: commentHandler,
		tagHandler:     tagHandler,
	}
}

// User 返回 userHandler 实例
func (h *Handler) User() *userHandler {
	return h.userHandler
}

// Article 返回 articleHandler 实例
func (h *Handler) Article() *articleHandler {
	return h.articleHandler
}

// Comment 返回 commentHandler 实例
func (h *Handler) Comment() *commentHandler {
	return h.commentHandler
}

// Tag 返回 tagHandler 实例
func (h *Handler) Tag() *tagHandler {
	return h.tagHandler
}

// Biz 返回 Biz 实例
func (h *Handler) Biz() biz.IBiz {
	return h.biz
}
