package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/jwt"
	"github.com/onexstack/realworld/apiserver/model"
	"gorm.io/gorm"
)

type articleHandler struct {
	biz        biz.IBiz
	jwtManager *jwt.Manager
}

func newArticleHandler(biz biz.IBiz, jwtManager *jwt.Manager) *articleHandler {
	return &articleHandler{
		biz:        biz,
		jwtManager: jwtManager,
	}
}

type CreateArticleRequest struct {
	Article struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description" binding:"required"`
		Body        string   `json:"body" binding:"required"`
		TagList     []string `json:"tagList"`
	} `json:"article" binding:"required"`
}

type UpdateArticleRequest struct {
	Article struct {
		Title       *string   `json:"title"`
		Description *string   `json:"description"`
		Body        *string   `json:"body"`
		TagList     *[]string `json:"tagList"`
	} `json:"article" binding:"required"`
}

type ArticleResponse struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Body           string   `json:"body"`
	TagList        []string `json:"tagList"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	Author         struct {
		Username  string  `json:"username"`
		Bio       *string `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"author"`
}

type ArticlesResponse struct {
	Articles      []ArticleResponse `json:"articles"`
	ArticlesCount int               `json:"articlesCount"`
}

type SingleArticleResponse struct {
	Article ArticleResponse `json:"article"`
}

type CursorPaginationResponse struct {
	Articles   []ArticleResponse `json:"articles"`
	HasMore    bool              `json:"hasMore"`
	NextCursor int64             `json:"nextCursor,omitempty"`
}

type articleRelationData struct {
	authors        map[int64]*model.UserM
	tags           map[int64][]string
	favoriteCounts map[int64]int64
	favorited      map[int64]bool
	following      map[int64]bool
}

type articleTagRow struct {
	ArticleModelID int64  `gorm:"column:article_model_id"`
	Tag            string `gorm:"column:tag"`
}

type favoriteCountRow struct {
	FavoriteID int64 `gorm:"column:favorite_id"`
	Count      int64 `gorm:"column:count"`
}

func (h *articleHandler) CreateArticle(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	description := req.Article.Description
	body := req.Article.Body
	article := &model.ArticleM{
		Slug:        h.generateUniqueSlug(c.Request.Context(), req.Article.Title),
		Title:       req.Article.Title,
		Description: &description,
		Body:        &body,
		AuthorID:    userID,
	}

	createdArticle, err := h.biz.Article().CreateArticle(c.Request.Context(), article)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.replaceArticleTags(c.Request.Context(), createdArticle.ID, req.Article.TagList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.buildArticleResponse(c.Request.Context(), createdArticle, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SingleArticleResponse{Article: resp})
}

func (h *articleHandler) UpdateArticle(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slug := c.Param("slug")
	if slug == "" || slug == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article slug"})
		return
	}

	existingArticle, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || existingArticle == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if existingArticle.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this article"})
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := make(map[string]interface{})
	if req.Article.Title != nil {
		update["title"] = *req.Article.Title
	}
	if req.Article.Description != nil {
		update["description"] = req.Article.Description
	}
	if req.Article.Body != nil {
		update["body"] = req.Article.Body
	}

	if len(update) > 0 {
		if _, err := h.biz.Article().UpdateArticle(c.Request.Context(), existingArticle.ID, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Article.TagList != nil {
		if err := h.replaceArticleTags(c.Request.Context(), existingArticle.ID, *req.Article.TagList); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updatedArticle, err := h.biz.Article().GetArticleByID(c.Request.Context(), existingArticle.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.buildArticleResponse(c.Request.Context(), updatedArticle, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SingleArticleResponse{Article: resp})
}

func (h *articleHandler) GetArticlesWithCursor(c *gin.Context) {
	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", "0"), 10, 64)
	limit := parseLimit(c.DefaultQuery("limit", "20"))

	hasMore, articles, nextCursor, err := h.biz.Article().GetArticlesWithCursor(c.Request.Context(), cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, _ := currentUserID(c)
	responses, err := h.buildArticleResponses(c.Request.Context(), articles, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := CursorPaginationResponse{
		Articles: responses,
		HasMore:  hasMore,
	}
	if hasMore && len(articles) > 0 {
		resp.NextCursor = nextCursor
	}

	c.JSON(http.StatusOK, resp)
}

func (h *articleHandler) GetArticlesWithDeferredJoin(c *gin.Context) {
	offset, limit := parseOffsetAndLimit(c)

	total, articles, err := h.biz.Article().GetArticlesWithDeferredJoin(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, _ := currentUserID(c)
	responses, err := h.buildArticleResponses(c.Request.Context(), articles, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ArticlesResponse{
		Articles:      responses,
		ArticlesCount: int(total),
	})
}

func (h *articleHandler) DeleteArticle(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slug := c.Param("slug")
	if slug == "" || slug == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article slug"})
		return
	}

	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if article.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this article"})
		return
	}

	if err := h.biz.Article().DeleteArticle(c.Request.Context(), article.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *articleHandler) GetArticle(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" || slug == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article slug"})
		return
	}

	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	userID, _ := currentUserID(c)
	resp, err := h.buildArticleResponse(c.Request.Context(), article, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SingleArticleResponse{Article: resp})
}

func (h *articleHandler) GetArticles(c *gin.Context) {
	offset, limit := parseOffsetAndLimit(c)
	author := strings.TrimSpace(c.Query("author"))
	tag := strings.TrimSpace(c.Query("tag"))
	favorited := strings.TrimSpace(c.Query("favorited"))

	var (
		total    int64
		articles []*model.ArticleM
		err      error
	)

	if author == "" && tag == "" && favorited == "" {
		total, articles, err = h.biz.Article().GetArticles(c.Request.Context(), offset, limit)
	} else {
		conditions := make(map[string]interface{})
		if author != "" {
			authorUser, authorErr := h.biz.User().GetUserByUsername(c.Request.Context(), author)
			if authorErr != nil {
				c.JSON(http.StatusOK, ArticlesResponse{Articles: []ArticleResponse{}, ArticlesCount: 0})
				return
			}
			conditions["author_id"] = authorUser.ID
		}
		if tag != "" {
			conditions["tag"] = tag
		}
		if favorited != "" {
			conditions["favorited"] = favorited
		}
		total, articles, err = h.biz.Article().GetArticlesByCondition(c.Request.Context(), conditions, offset, limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, _ := currentUserID(c)
	responses, err := h.buildArticleResponses(c.Request.Context(), articles, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ArticlesResponse{
		Articles:      responses,
		ArticlesCount: int(total),
	})
}

func (h *articleHandler) GetFeed(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	offset, limit := parseOffsetAndLimit(c)

	total, articles, err := h.biz.Article().GetFeed(c.Request.Context(), userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses, err := h.buildArticleResponses(c.Request.Context(), articles, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ArticlesResponse{
		Articles:      responses,
		ArticlesCount: int(total),
	})
}

func (h *articleHandler) FavoriteArticle(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slug := c.Param("slug")
	if slug == "" || slug == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article slug"})
		return
	}

	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if err := h.biz.Article().FavoriteArticle(c.Request.Context(), article.ID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.buildArticleResponse(c.Request.Context(), article, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SingleArticleResponse{Article: resp})
}

func (h *articleHandler) UnfavoriteArticle(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slug := c.Param("slug")
	if slug == "" || slug == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article slug"})
		return
	}

	article, err := h.biz.Article().GetArticleBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if err := h.biz.Article().UnfavoriteArticle(c.Request.Context(), article.ID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.buildArticleResponse(c.Request.Context(), article, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SingleArticleResponse{Article: resp})
}

func (h *articleHandler) buildArticleResponse(ctx context.Context, article *model.ArticleM, currentUserID int64) (ArticleResponse, error) {
	responses, err := h.buildArticleResponses(ctx, []*model.ArticleM{article}, currentUserID)
	if err != nil {
		return ArticleResponse{}, err
	}
	if len(responses) == 0 {
		return ArticleResponse{}, fmt.Errorf("article response not found")
	}

	return responses[0], nil
}

func (h *articleHandler) buildArticleResponses(ctx context.Context, articles []*model.ArticleM, currentUserID int64) ([]ArticleResponse, error) {
	relations, err := h.loadArticleRelations(ctx, articles, currentUserID)
	if err != nil {
		return nil, err
	}

	responses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		author := relations.authors[article.AuthorID]
		resp := ArticleResponse{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    derefString(article.Description),
			Body:           derefString(article.Body),
			TagList:        ensureStringSlice(relations.tags[article.ID]),
			CreatedAt:      formatAPITime(article.CreatedAt),
			UpdatedAt:      formatAPITime(article.UpdatedAt),
			Favorited:      relations.favorited[article.ID],
			FavoritesCount: int(relations.favoriteCounts[article.ID]),
		}
		if author != nil {
			resp.Author.Username = author.Username
			resp.Author.Bio = author.Bio
			resp.Author.Image = author.Image
			resp.Author.Following = relations.following[author.ID]
		}
		responses[i] = resp
	}

	return responses, nil
}

func (h *articleHandler) loadArticleRelations(ctx context.Context, articles []*model.ArticleM, currentUserID int64) (articleRelationData, error) {
	relations := articleRelationData{
		authors:        map[int64]*model.UserM{},
		tags:           map[int64][]string{},
		favoriteCounts: map[int64]int64{},
		favorited:      map[int64]bool{},
		following:      map[int64]bool{},
	}
	if len(articles) == 0 {
		return relations, nil
	}

	articleIDs := make([]int64, 0, len(articles))
	authorIDs := make([]int64, 0, len(articles))
	for _, article := range articles {
		articleIDs = append(articleIDs, article.ID)
		authorIDs = append(authorIDs, article.AuthorID)
	}
	articleIDs = uniqueInt64s(articleIDs)
	authorIDs = uniqueInt64s(authorIDs)

	db := h.biz.Store().DB(ctx)
	if db == nil {
		for _, article := range articles {
			if _, exists := relations.tags[article.ID]; !exists {
				relations.tags[article.ID] = []string{}
			}

			count, _ := h.biz.Article().GetFavoritesCount(ctx, article.ID)
			relations.favoriteCounts[article.ID] = count

			if currentUserID > 0 {
				favorited, _ := h.biz.Article().IsFavorited(ctx, article.ID, currentUserID)
				relations.favorited[article.ID] = favorited
			}

			if _, exists := relations.authors[article.AuthorID]; !exists {
				author, _ := h.biz.User().GetUser(ctx, article.AuthorID)
				if author != nil {
					relations.authors[author.ID] = author
					if currentUserID > 0 {
						following, _ := h.biz.User().IsFollowing(ctx, currentUserID, author.ID)
						relations.following[author.ID] = following
					}
				}
			}
		}

		return relations, nil
	}

	authors, err := loadUsersByIDs(ctx, db, authorIDs)
	if err != nil {
		return relations, err
	}
	relations.authors = authors

	tags, err := loadArticleTags(ctx, db, articleIDs)
	if err != nil {
		return relations, err
	}
	relations.tags = tags

	counts, err := loadFavoriteCounts(ctx, db, articleIDs)
	if err != nil {
		return relations, err
	}
	relations.favoriteCounts = counts

	if currentUserID > 0 {
		favorited, err := loadFavoritedSet(ctx, db, currentUserID, articleIDs)
		if err != nil {
			return relations, err
		}
		relations.favorited = favorited

		following, err := loadFollowingSet(ctx, db, currentUserID, authorIDs)
		if err != nil {
			return relations, err
		}
		relations.following = following
	}

	for _, articleID := range articleIDs {
		if _, exists := relations.tags[articleID]; !exists {
			relations.tags[articleID] = []string{}
		}
	}

	return relations, nil
}

func (h *articleHandler) replaceArticleTags(ctx context.Context, articleID int64, tags []string) error {
	db := h.biz.Store().DB(ctx)
	if db == nil {
		return nil
	}

	normalizedTags := normalizeTags(tags)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_model_id = ?", articleID).Delete(&model.ArticleTagM{}).Error; err != nil {
			return err
		}

		for _, tagName := range normalizedTags {
			tag := model.TagM{}
			if err := tx.Where("tag = ?", tagName).FirstOrCreate(&tag, model.TagM{Tag: tagName}).Error; err != nil {
				return err
			}

			relation := &model.ArticleTagM{
				ArticleModelID: articleID,
				TagModelID:     tag.ID,
			}
			if err := tx.Create(relation).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (h *articleHandler) generateUniqueSlug(ctx context.Context, title string) string {
	base := slugify(title)
	if base == "" {
		base = "article"
	}

	slug := base
	for i := 2; ; i++ {
		article, err := h.biz.Article().GetArticleBySlug(ctx, slug)
		if err != nil || article == nil {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

func loadUsersByIDs(ctx context.Context, db *gorm.DB, userIDs []int64) (map[int64]*model.UserM, error) {
	result := map[int64]*model.UserM{}
	if len(userIDs) == 0 {
		return result, nil
	}

	var users []*model.UserM
	if err := db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	for _, user := range users {
		result[user.ID] = user
	}

	return result, nil
}

func loadFollowingSet(ctx context.Context, db *gorm.DB, currentUserID int64, authorIDs []int64) (map[int64]bool, error) {
	result := map[int64]bool{}
	if currentUserID == 0 || len(authorIDs) == 0 {
		return result, nil
	}

	var follows []model.FollowM
	if err := db.WithContext(ctx).Where("followed_by_id = ? AND following_id IN ?", currentUserID, authorIDs).Find(&follows).Error; err != nil {
		return nil, err
	}

	for _, follow := range follows {
		result[follow.FollowingID] = true
	}

	return result, nil
}

func loadFavoritedSet(ctx context.Context, db *gorm.DB, currentUserID int64, articleIDs []int64) (map[int64]bool, error) {
	result := map[int64]bool{}
	if currentUserID == 0 || len(articleIDs) == 0 {
		return result, nil
	}

	var favorites []model.FavoriteM
	if err := db.WithContext(ctx).Where("favorite_by_id = ? AND favorite_id IN ?", currentUserID, articleIDs).Find(&favorites).Error; err != nil {
		return nil, err
	}

	for _, favorite := range favorites {
		result[favorite.FavoriteID] = true
	}

	return result, nil
}

func loadFavoriteCounts(ctx context.Context, db *gorm.DB, articleIDs []int64) (map[int64]int64, error) {
	result := map[int64]int64{}
	if len(articleIDs) == 0 {
		return result, nil
	}

	var rows []favoriteCountRow
	if err := db.WithContext(ctx).
		Model(&model.FavoriteM{}).
		Select("favorite_id, COUNT(*) AS count").
		Where("favorite_id IN ?", articleIDs).
		Group("favorite_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.FavoriteID] = row.Count
	}

	return result, nil
}

func loadArticleTags(ctx context.Context, db *gorm.DB, articleIDs []int64) (map[int64][]string, error) {
	result := map[int64][]string{}
	if len(articleIDs) == 0 {
		return result, nil
	}

	var rows []articleTagRow
	if err := db.WithContext(ctx).
		Table("article_tags AS at").
		Select("at.article_model_id, t.tag").
		Joins("JOIN tag_models AS t ON t.id = at.tag_model_id").
		Where("at.article_model_id IN ?", articleIDs).
		Order("t.tag ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.ArticleModelID] = append(result[row.ArticleModelID], row.Tag)
	}

	return result, nil
}

func parseOffsetAndLimit(c *gin.Context) (int, int) {
	limit := parseLimit(c.DefaultQuery("limit", "20"))

	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		page, _ := strconv.Atoi(rawPage)
		if page <= 0 {
			page = 1
		}
		return (page - 1) * limit, limit
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	return offset, limit
}

func parseLimit(raw string) int {
	limit, _ := strconv.Atoi(raw)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ensureStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.TrimSpace(tag)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
