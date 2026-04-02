package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/model"
)

type commentHandler struct {
	biz biz.IBiz
}

func newCommentHandler(biz biz.IBiz) *commentHandler {
	return &commentHandler{biz: biz}
}

type CommentRequest struct {
	Comment struct {
		Body string `json:"body" binding:"required"`
	} `json:"comment" binding:"required"`
}

type Comment struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Body      string `json:"body"`
	Author    struct {
		Username  string  `json:"username"`
		Bio       *string `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"author"`
}

type CommentResponse struct {
	Comment Comment `json:"comment"`
}

type CommentsResponse struct {
	Comments []Comment `json:"comments"`
}

func (h *commentHandler) CreateComment(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slug := c.Param("slug")
	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	var req CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := &model.CommentM{
		Body:      req.Comment.Body,
		ArticleID: article.ID,
		AuthorID:  userID,
	}

	createdComment, err := h.biz.Comment().CreateComment(c.Request.Context(), comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.buildCommentResponse(c.Request.Context(), createdComment, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CommentResponse{Comment: resp})
}

func (h *commentHandler) GetCommentsByArticle(c *gin.Context) {
	slug := c.Param("slug")
	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	offset, limit := parseOffsetAndLimit(c)
	_, comments, err := h.biz.Comment().GetCommentsByArticleID(c.Request.Context(), article.ID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, _ := currentUserID(c)
	responses, err := h.buildCommentResponses(c.Request.Context(), comments, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CommentsResponse{Comments: responses})
}

func (h *commentHandler) DeleteComment(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := h.biz.Comment().GetCommentByID(c.Request.Context(), commentID)
	if err != nil || comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	if slug := c.Param("slug"); slug != "" {
		article, articleErr := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
		if articleErr != nil || article == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		if comment.ArticleID != article.ID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
	}

	if comment.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this comment"})
		return
	}

	if err := h.biz.Comment().DeleteComment(c.Request.Context(), commentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *commentHandler) buildCommentResponse(ctx context.Context, comment *model.CommentM, currentUserID int64) (Comment, error) {
	responses, err := h.buildCommentResponses(ctx, []*model.CommentM{comment}, currentUserID)
	if err != nil {
		return Comment{}, err
	}
	if len(responses) == 0 {
		return Comment{}, nil
	}

	return responses[0], nil
}

func (h *commentHandler) buildCommentResponses(ctx context.Context, comments []*model.CommentM, currentUserID int64) ([]Comment, error) {
	if len(comments) == 0 {
		return []Comment{}, nil
	}

	authorIDs := make([]int64, 0, len(comments))
	for _, comment := range comments {
		authorIDs = append(authorIDs, comment.AuthorID)
	}
	authorIDs = uniqueInt64s(authorIDs)

	authors := map[int64]*model.UserM{}
	following := map[int64]bool{}
	db := h.biz.Store().DB(ctx)
	if db != nil {
		var err error
		authors, err = loadUsersByIDs(ctx, db, authorIDs)
		if err != nil {
			return nil, err
		}
		if currentUserID > 0 {
			following, err = loadFollowingSet(ctx, db, currentUserID, authorIDs)
			if err != nil {
				return nil, err
			}
		}
	} else {
		for _, authorID := range authorIDs {
			author, _ := h.biz.User().GetUser(ctx, authorID)
			if author != nil {
				authors[authorID] = author
				if currentUserID > 0 {
					followingValue, _ := h.biz.User().IsFollowing(ctx, currentUserID, authorID)
					following[authorID] = followingValue
				}
			}
		}
	}

	responses := make([]Comment, len(comments))
	for i, comment := range comments {
		resp := Comment{
			ID:        comment.ID,
			CreatedAt: formatAPITime(comment.CreatedAt),
			UpdatedAt: formatAPITime(comment.UpdatedAt),
			Body:      comment.Body,
		}

		if author := authors[comment.AuthorID]; author != nil {
			resp.Author.Username = author.Username
			resp.Author.Bio = author.Bio
			resp.Author.Image = author.Image
			resp.Author.Following = following[author.ID]
		}

		responses[i] = resp
	}

	return responses, nil
}
