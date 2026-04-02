package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
)

// tagHandler 定义了标签相关的HTTP请求处理器
type tagHandler struct {
	biz biz.IBiz
}

// newTagHandler 创建一个 tagHandler 实例
func newTagHandler(biz biz.IBiz) *tagHandler {
	return &tagHandler{
		biz: biz,
	}
}

// TagsResponse 标签列表响应结构
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// GetTags 获取所有标签
func (h *tagHandler) GetTags(c *gin.Context) {
	// 获取所有标签
	tags, err := h.biz.Tag().GetAllTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应
	resp := TagsResponse{}
	resp.Tags = make([]string, len(tags))
	for i, tag := range tags {
		resp.Tags[i] = tag.Tag
	}

	c.JSON(http.StatusOK, resp)
}
