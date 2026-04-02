package biz

import (
	"context"
	"errors"

	"github.com/onexstack/realworld/apiserver/model"
	"github.com/onexstack/realworld/apiserver/store"
)

// ArticleBiz 定义了文章业务需要实现的方法
type ArticleBiz interface {
	// 创建文章
	CreateArticle(ctx context.Context, article *model.ArticleM) (*model.ArticleM, error)
	// 更新文章
	UpdateArticle(ctx context.Context, articleID int64, update map[string]interface{}) (*model.ArticleM, error)
	// 删除文章
	DeleteArticle(ctx context.Context, articleID int64) error
	// 根据ID获取文章
	GetArticleByID(ctx context.Context, articleID int64) (*model.ArticleM, error)
	// 根据Slug获取文章
	GetArticleBySlug(ctx context.Context, slug string) (*model.ArticleM, error)
	// 获取文章列表
	GetArticles(ctx context.Context, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据作者获取文章列表
	GetArticlesByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据标签获取文章列表
	GetArticlesByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据收藏用户获取文章列表
	GetArticlesByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error)
	// 获取关注用户的文章feed
	GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error)
	// 收藏文章
	FavoriteArticle(ctx context.Context, articleID, userID int64) error
	// 取消收藏文章
	UnfavoriteArticle(ctx context.Context, articleID, userID int64) error
	// 检查是否收藏了文章
	IsFavorited(ctx context.Context, articleID, userID int64) (bool, error)
	// 获取文章的点赞数
	GetFavoritesCount(ctx context.Context, articleID int64) (int64, error)
	// 根据复合条件获取文章列表
	GetArticlesByCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (int64, []*model.ArticleM, error)
	// 使用游标获取文章列表
	GetArticlesWithCursor(ctx context.Context, cursor int64, limit int) (hasMore bool, articles []*model.ArticleM, nextCursor int64, err error)
	// 使用游标和复合条件获取文章列表
	GetArticlesWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (hasMore bool, articles []*model.ArticleM, nextCursor int64, err error)
	// 使用延迟关联优化深度分页（保留跳页能力）
	GetArticlesWithDeferredJoin(ctx context.Context, offset, limit int) (int64, []*model.ArticleM, error)
}

// articleBiz 是 ArticleBiz 接口的实现
type articleBiz struct {
	store store.IStore
}

// newArticleBiz 创建一个 articleBiz 实例
func newArticleBiz(store store.IStore) *articleBiz {
	return &articleBiz{
		store: store,
	}
}

// CreateArticle 创建文章
func (b *articleBiz) CreateArticle(ctx context.Context, article *model.ArticleM) (*model.ArticleM, error) {
	// 检查文章标题是否已存在
	// 这里可以根据业务需求添加更多验证

	// 创建文章
	err := b.store.Article().Create(ctx, article)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// UpdateArticle 更新文章
func (b *articleBiz) UpdateArticle(ctx context.Context, articleID int64, update map[string]interface{}) (*model.ArticleM, error) {
	// 获取文章
	article, err := b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
	if err != nil {
		return nil, errors.New("article not found")
	}

	// 更新文章信息
	if title, ok := update["title"].(string); ok {
		article.Title = title
	}

	if slug, ok := update["slug"].(string); ok {
		article.Slug = slug
	}

	if body, ok := update["body"].(*string); ok {
		article.Body = body
	}

	if description, ok := update["description"].(*string); ok {
		article.Description = description
	}

	// 保存更新
	err = b.store.Article().Update(ctx, article)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// DeleteArticle 删除文章
func (b *articleBiz) DeleteArticle(ctx context.Context, articleID int64) error {
	// 检查文章是否存在
	_, err := b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
	if err != nil {
		return errors.New("article not found")
	}

	// 删除文章
	return b.store.Article().Delete(ctx, map[string]interface{}{"id": articleID})
}

// GetArticleByID 根据ID获取文章
func (b *articleBiz) GetArticleByID(ctx context.Context, articleID int64) (*model.ArticleM, error) {
	return b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
}

// GetArticleBySlug 根据Slug获取文章
func (b *articleBiz) GetArticleBySlug(ctx context.Context, slug string) (*model.ArticleM, error) {
	return b.store.Article().GetBySlug(ctx, slug)
}

// GetArticles 获取文章列表
func (b *articleBiz) GetArticles(ctx context.Context, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().List(ctx, nil, offset, limit)
}

// GetArticlesByAuthor 根据作者获取文章列表
func (b *articleBiz) GetArticlesByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().GetByAuthor(ctx, authorID, offset, limit)
}

// GetArticlesByTag 根据标签获取文章列表
func (b *articleBiz) GetArticlesByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().GetByTag(ctx, tag, offset, limit)
}

// GetArticlesByFavorited 根据收藏用户获取文章列表
func (b *articleBiz) GetArticlesByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().GetByFavorited(ctx, username, offset, limit)
}

// GetFeed 获取关注用户的文章feed
func (b *articleBiz) GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().GetFeed(ctx, userID, offset, limit)
}

// FavoriteArticle 收藏文章
func (b *articleBiz) FavoriteArticle(ctx context.Context, articleID, userID int64) error {
	// 检查文章是否存在
	_, err := b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
	if err != nil {
		return errors.New("article not found")
	}

	// 检查是否已经收藏
	isFavorited, err := b.store.Article().IsFavorited(ctx, articleID, userID)
	if err != nil {
		return err
	}

	if isFavorited {
		return errors.New("already favorited")
	}

	// 收藏文章
	return b.store.Article().Favorite(ctx, articleID, userID)
}

// UnfavoriteArticle 取消收藏文章
func (b *articleBiz) UnfavoriteArticle(ctx context.Context, articleID, userID int64) error {
	// 检查文章是否存在
	_, err := b.store.Article().Get(ctx, map[string]interface{}{"id": articleID})
	if err != nil {
		return errors.New("article not found")
	}

	// 检查是否已经收藏
	isFavorited, err := b.store.Article().IsFavorited(ctx, articleID, userID)
	if err != nil {
		return err
	}

	if !isFavorited {
		return errors.New("not favorited")
	}

	// 取消收藏文章
	return b.store.Article().Unfavorite(ctx, articleID, userID)
}

// IsFavorited 检查是否收藏了文章
func (b *articleBiz) IsFavorited(ctx context.Context, articleID, userID int64) (bool, error) {
	return b.store.Article().IsFavorited(ctx, articleID, userID)
}

// GetFavoritesCount 获取文章的点赞数
func (b *articleBiz) GetFavoritesCount(ctx context.Context, articleID int64) (int64, error) {
	return b.store.Article().GetFavoritesCount(ctx, articleID)
}

// GetArticlesByCondition 根据复合条件获取文章列表
func (b *articleBiz) GetArticlesByCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().GetByComplexCondition(ctx, conditions, offset, limit)
}

// GetArticlesWithCursor 使用游标获取文章列表
func (b *articleBiz) GetArticlesWithCursor(ctx context.Context, cursor int64, limit int) (hasMore bool, articles []*model.ArticleM, nextCursor int64, err error) {
	return b.store.Article().ListWithCursor(ctx, nil, cursor, limit)
}

// GetArticlesWithCursorAndCondition 使用游标和复合条件获取文章列表
func (b *articleBiz) GetArticlesWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (hasMore bool, articles []*model.ArticleM, nextCursor int64, err error) {
	return b.store.Article().ListWithCursorAndCondition(ctx, conditions, cursor, limit)
}

// GetArticlesWithDeferredJoin 使用延迟关联优化深度分页
func (b *articleBiz) GetArticlesWithDeferredJoin(ctx context.Context, offset, limit int) (int64, []*model.ArticleM, error) {
	return b.store.Article().ListWithDeferredJoin(ctx, nil, offset, limit)
}
