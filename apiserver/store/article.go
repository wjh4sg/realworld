package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/onexstack/realworld/apiserver/cache"
	"github.com/onexstack/realworld/apiserver/model"
)

// ArticleStore 定义了 article 模块在 store 层所实现的方法.
type ArticleStore interface {
	Create(ctx context.Context, obj *model.ArticleM) error
	Update(ctx context.Context, obj *model.ArticleM) error
	Delete(ctx context.Context, condition interface{}) error
	Get(ctx context.Context, condition interface{}) (*model.ArticleM, error)
	List(ctx context.Context, condition interface{}, offset, limit int) (int64, []*model.ArticleM, error)

	ArticleExpansion
}

// ArticleExpansion 定义了文章操作的附加方法.
type ArticleExpansion interface {
	// 根据作者ID获取文章列表
	GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据Slug获取文章
	GetBySlug(ctx context.Context, slug string) (*model.ArticleM, error)
	// 收藏文章
	Favorite(ctx context.Context, articleID, userID int64) error
	// 取消收藏文章
	Unfavorite(ctx context.Context, articleID, userID int64) error
	// 检查是否收藏了文章
	IsFavorited(ctx context.Context, articleID, userID int64) (bool, error)
	// 获取文章的点赞数
	GetFavoritesCount(ctx context.Context, articleID int64) (int64, error)
	// 根据标签获取文章列表
	GetByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据收藏用户获取文章列表
	GetByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error)
	// 获取关注用户的文章feed
	GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error)
	// 根据复合条件获取文章列表
	GetByComplexCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (int64, []*model.ArticleM, error)
	// 使用游标获取文章列表（支持大数据量）
	ListWithCursor(ctx context.Context, condition interface{}, cursor int64, limit int) (hasMore bool, ret []*model.ArticleM, nextCursor int64, err error)
	// 使用游标和复合条件获取文章列表
	ListWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (hasMore bool, ret []*model.ArticleM, nextCursor int64, err error)
	// 使用延迟关联优化深度分页（保留跳页能力）
	ListWithDeferredJoin(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.ArticleM, err error)
}

// articleStore 是 ArticleStore 接口的实现.
type articleStore struct {
	store *datastore
}

// 确保 articleStore 实现了 ArticleStore 接口.
var _ ArticleStore = (*articleStore)(nil)

// newArticleStore 创建 articleStore 的实例.
func newArticleStore(store *datastore) *articleStore {
	return &articleStore{store}
}

// Create 插入一条文章记录.
func (s *articleStore) Create(ctx context.Context, obj *model.ArticleM) error {
	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		slog.Error("Failed to insert article into database", "err", err, "article", obj)
		return errors.New("failed to insert article: " + err.Error())
	}

	// 文章创建后,清除所有列表缓存(不缓存新文章详情)
	if s.store.cache != nil {
		if err := s.store.cache.Article().DeleteArticle(ctx, obj.Slug); err != nil {
			slog.Warn("Failed to clear article detail cache after create", "err", err, "slug", obj.Slug)
		}
		if err := s.store.cache.Article().SetArticle(ctx, obj.Slug, obj); err != nil {
			slog.Warn("Failed to set article cache after create", "err", err, "slug", obj.Slug)
		}
		if err := s.store.cache.Article().InvalidateArticleListCache(ctx); err != nil {
			slog.Warn("Failed to invalidate article list cache after create", "err", err)
		}
	}

	return nil
}

// Update 更新文章数据库记录.
func (s *articleStore) Update(ctx context.Context, obj *model.ArticleM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		slog.Error("Failed to update article in database", "err", err, "article", obj)
		return errors.New("failed to update article: " + err.Error())
	}

	// 文章更新后,删除详情缓存并清除所有列表缓存
	if s.store.cache != nil {
		if err := s.store.cache.Article().DeleteArticle(ctx, obj.Slug); err != nil {
			slog.Warn("Failed to delete article cache after update", "err", err, "slug", obj.Slug)
		}
		if err := s.store.cache.Article().InvalidateArticleListCache(ctx); err != nil {
			slog.Warn("Failed to invalidate article list cache after update", "err", err)
		}
	}

	return nil
}

// Delete 根据条件删除文章记录.
func (s *articleStore) Delete(ctx context.Context, condition interface{}) error {
	// 先获取文章信息用于清除缓存
	var article model.ArticleM
	if err := s.store.DB(ctx).Where(condition).First(&article).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to get article before delete", "err", err, "condition", condition)
	}

	err := s.store.DB(ctx).Where(condition).Delete(new(model.ArticleM)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("Failed to delete article from database", "err", err, "condition", condition)
		return errors.New("failed to delete article: " + err.Error())
	}

	// 文章删除后,删除详情缓存并清除所有列表缓存
	if s.store.cache != nil && article.Slug != "" {
		if err := s.store.cache.Article().DeleteArticle(ctx, article.Slug); err != nil {
			slog.Warn("Failed to delete article cache after delete", "err", err, "slug", article.Slug)
		}
		if err := s.store.cache.Article().InvalidateArticleListCache(ctx); err != nil {
			slog.Warn("Failed to invalidate article list cache after delete", "err", err)
		}
	}

	return nil
}

// Get 根据条件查询文章记录.
func (s *articleStore) Get(ctx context.Context, condition interface{}) (*model.ArticleM, error) {
	var obj model.ArticleM
	if err := s.store.DB(ctx).Where(condition).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("article not found")
		}
		slog.Error("Failed to retrieve article from database", "err", err, "condition", condition)
		return nil, errors.New("failed to get article: " + err.Error())
	}

	return &obj, nil
}

// List 返回文章列表和总数.
func (s *articleStore) List(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.ArticleM, err error) {
	// 1. 生成查询键
	queryKey := ""
	if s.store.cache != nil {
		params := map[string]interface{}{
			"condition": condition,
			"offset":    offset,
			"limit":     limit,
		}
		queryKey = cache.GenerateQueryKey("list", params)

		// 2. 尝试从缓存获取
		total, articles, cacheErr := s.store.cache.Article().GetArticleList(ctx, queryKey)
		if cacheErr == nil && articles != nil {
			slog.Debug("Article list cache hit", "queryKey", queryKey)
			return total, articles, nil
		}
		slog.Debug("Article list cache miss", "queryKey", queryKey)
	}

	// 3. 缓存未命中,查询数据库
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	// 先获取总数
	if err = db.Model(&model.ArticleM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count articles from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to count articles: " + err.Error())
	}

	// 再获取列表
	if err = db.Offset(offset).Limit(limit).Order("created_at desc").Find(&ret).Error; err != nil {
		slog.Error("Failed to list articles from database", "err", err, "condition", condition)
		return count, nil, errors.New("failed to list articles: " + err.Error())
	}

	// 4. 回写缓存
	if s.store.cache != nil && queryKey != "" {
		if cacheErr := s.store.cache.Article().SetArticleList(ctx, queryKey, count, ret); cacheErr != nil {
			slog.Warn("Failed to set article list to cache", "err", cacheErr, "queryKey", queryKey)
		}
	}

	return
}

// GetByAuthor 根据作者ID获取文章列表
func (s *articleStore) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	condition := map[string]interface{}{"author_id": authorID}
	return s.List(ctx, condition, offset, limit)
}

// GetBySlug 根据Slug获取文章
func (s *articleStore) GetBySlug(ctx context.Context, slug string) (*model.ArticleM, error) {
	// 1. 尝试从缓存获取
	if s.store.cache != nil {
		article, err := s.store.cache.Article().GetArticle(ctx, slug)
		if err == nil && article != nil {
			slog.Debug("Article cache hit", "slug", slug)
			return article, nil
		}
		if err != nil && err.Error() == "article not found" {
			// 空标记,直接返回
			return nil, errors.New("article not found")
		}
		slog.Debug("Article cache miss", "slug", slug)
	}

	// 2. 缓存未命中,查询数据库
	var obj model.ArticleM
	if err := s.store.DB(ctx).Where("slug = ?", slug).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 3. 数据不存在,设置空标记防止缓存穿透
			if s.store.cache != nil {
				s.store.cache.Article().SetArticleNotFound(ctx, slug)
			}
			return nil, errors.New("article not found")
		}
		slog.Error("Failed to get article by slug", "err", err, "slug", slug)
		return nil, errors.New("failed to get article by slug: " + err.Error())
	}

	// 4. 回写缓存
	if s.store.cache != nil {
		if err := s.store.cache.Article().SetArticle(ctx, slug, &obj); err != nil {
			slog.Warn("Failed to set article to cache", "err", err, "slug", slug)
		}
	}

	return &obj, nil
}

// Favorite 收藏文章
func (s *articleStore) Favorite(ctx context.Context, articleID, userID int64) error {
	favorite := &model.FavoriteM{
		FavoriteID:   articleID,
		FavoriteByID: userID,
	}

	// 检查是否已经收藏
	var existingFavorite model.FavoriteM
	if err := s.store.DB(ctx).Where("favorite_id = ? AND favorite_by_id = ?", articleID, userID).First(&existingFavorite).Error; err == nil {
		// 已经收藏，不需要重复操作
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他错误
		slog.Error("Failed to check existing favorite", "err", err, "article_id", articleID, "user_id", userID)
		return errors.New("failed to check existing favorite: " + err.Error())
	}

	// 创建收藏关系
	if err := s.store.DB(ctx).Create(favorite).Error; err != nil {
		slog.Error("Failed to create favorite", "err", err, "favorite", favorite)
		return errors.New("failed to create favorite: " + err.Error())
	}

	// 收藏后，清除所有列表缓存(因为favorited状态会影响列表展示)
	if s.store.cache != nil {
		if err := s.store.cache.Article().InvalidateArticleListCache(ctx); err != nil {
			slog.Warn("Failed to invalidate article list cache after favorite", "err", err)
		}
	}

	return nil
}

// Unfavorite 取消收藏文章
func (s *articleStore) Unfavorite(ctx context.Context, articleID, userID int64) error {
	result := s.store.DB(ctx).Where("favorite_id = ? AND favorite_by_id = ?", articleID, userID).Delete(&model.FavoriteM{})
	if result.Error != nil {
		slog.Error("Failed to delete favorite", "err", result.Error, "article_id", articleID, "user_id", userID)
		return errors.New("failed to delete favorite: " + result.Error.Error())
	}

	// 取消收藏后，清除所有列表缓存(因为favorited状态会影响列表展示)
	if s.store.cache != nil {
		if err := s.store.cache.Article().InvalidateArticleListCache(ctx); err != nil {
			slog.Warn("Failed to invalidate article list cache after unfavorite", "err", err)
		}
	}

	return nil
}

// IsFavorited 检查是否收藏了文章
func (s *articleStore) IsFavorited(ctx context.Context, articleID, userID int64) (bool, error) {
	var favorite model.FavoriteM
	err := s.store.DB(ctx).Where("favorite_id = ? AND favorite_by_id = ?", articleID, userID).First(&favorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		slog.Error("Failed to check favorite", "err", err, "article_id", articleID, "user_id", userID)
		return false, errors.New("failed to check favorite: " + err.Error())
	}

	return true, nil
}

// GetFavoritesCount 获取文章的点赞数
func (s *articleStore) GetFavoritesCount(ctx context.Context, articleID int64) (int64, error) {
	var count int64
	err := s.store.DB(ctx).Model(&model.FavoriteM{}).Where("favorite_id = ?", articleID).Count(&count).Error
	if err != nil {
		slog.Error("Failed to get favorites count", "err", err, "article_id", articleID)
		return 0, errors.New("failed to get favorites count: " + err.Error())
	}

	return count, nil
}

// GetByTag 根据标签获取文章列表
func (s *articleStore) GetByTag(ctx context.Context, tag string, offset, limit int) (int64, []*model.ArticleM, error) {
	// 首先根据标签名获取标签ID
	var tagModel model.TagM
	if err := s.store.DB(ctx).Where("tag = ?", tag).First(&tagModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, []*model.ArticleM{}, nil
		}
		slog.Error("Failed to get tag by name", "err", err, "tag", tag)
		return 0, nil, errors.New("failed to get tag by name: " + err.Error())
	}

	// 然后通过文章标签关联表获取文章ID列表
	var articleTagModels []model.ArticleTagM
	if err := s.store.DB(ctx).Where("tag_model_id = ?", tagModel.ID).Find(&articleTagModels).Error; err != nil {
		slog.Error("Failed to get article tags", "err", err, "tag_id", tagModel.ID)
		return 0, nil, errors.New("failed to get article tags: " + err.Error())
	}

	if len(articleTagModels) == 0 {
		return 0, []*model.ArticleM{}, nil
	}

	// 提取文章ID
	articleIDs := make([]int64, len(articleTagModels))
	for i, at := range articleTagModels {
		articleIDs[i] = at.ArticleModelID
	}

	// 查询文章列表
	condition := map[string]interface{}{"id": articleIDs}
	return s.List(ctx, condition, offset, limit)
}

// GetByFavorited 根据收藏用户获取文章列表
func (s *articleStore) GetByFavorited(ctx context.Context, username string, offset, limit int) (int64, []*model.ArticleM, error) {
	// 首先根据用户名获取用户ID
	var user model.UserM
	if err := s.store.DB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, []*model.ArticleM{}, nil
		}
		slog.Error("Failed to get user by username", "err", err, "username", username)
		return 0, nil, errors.New("failed to get user by username: " + err.Error())
	}

	// 然后获取该用户收藏的文章ID列表
	var favoriteModels []model.FavoriteM
	if err := s.store.DB(ctx).Where("favorite_by_id = ?", user.ID).Find(&favoriteModels).Error; err != nil {
		slog.Error("Failed to get favorites", "err", err, "user_id", user.ID)
		return 0, nil, errors.New("failed to get favorites: " + err.Error())
	}

	if len(favoriteModels) == 0 {
		return 0, []*model.ArticleM{}, nil
	}

	// 提取文章ID
	articleIDs := make([]int64, len(favoriteModels))
	for i, f := range favoriteModels {
		articleIDs[i] = f.FavoriteID
	}

	// 查询文章列表
	condition := map[string]interface{}{"id": articleIDs}
	return s.List(ctx, condition, offset, limit)
}

// GetFeed 获取关注用户的文章feed
func (s *articleStore) GetFeed(ctx context.Context, userID int64, offset, limit int) (int64, []*model.ArticleM, error) {
	// 首先获取用户关注的用户ID列表
	var followModels []model.FollowM
	if err := s.store.DB(ctx).Where("followed_by_id = ?", userID).Find(&followModels).Error; err != nil {
		slog.Error("Failed to get follows", "err", err, "user_id", userID)
		return 0, nil, errors.New("failed to get follows: " + err.Error())
	}

	// 提取关注的用户ID
	followingIDs := make([]int64, len(followModels))
	for i, f := range followModels {
		followingIDs[i] = f.FollowingID
	}

	// 如果没有关注任何用户，返回空列表
	if len(followingIDs) == 0 {
		return 0, []*model.ArticleM{}, nil
	}

	// 查询关注用户的文章列表
	condition := map[string]interface{}{"author_id": followingIDs}
	return s.List(ctx, condition, offset, limit)
}

// GetByComplexCondition 根据复合条件获取文章列表
func (s *articleStore) GetByComplexCondition(ctx context.Context, conditions map[string]interface{}, offset, limit int) (count int64, ret []*model.ArticleM, err error) {
	// 1. 生成查询键
	queryKey := ""
	if s.store.cache != nil {
		params := map[string]interface{}{
			"conditions": conditions,
			"offset":     offset,
			"limit":      limit,
		}
		queryKey = cache.GenerateQueryKey("complex", params)

		// 2. 尝试从缓存获取
		total, articles, cacheErr := s.store.cache.Article().GetArticleList(ctx, queryKey)
		if cacheErr == nil && articles != nil {
			slog.Debug("Complex article list cache hit", "query_key", queryKey)
			return total, articles, nil
		}
		slog.Debug("Complex article list cache miss", "query_key", queryKey)
	}

	// 3. 缓存未命中，查询数据库
	db := s.store.DB(ctx)

	// 分离特殊条件和普通条件
	var specialConditions = make(map[string]interface{})
	var normalConditions = make(map[string]interface{})

	for key, value := range conditions {
		if key == "tag" || key == "favorited" || key == "author_id" {
			specialConditions[key] = value
		} else {
			normalConditions[key] = value
		}
	}

	// 处理标签条件
	if tagVal, hasTag := specialConditions["tag"]; hasTag {
		tag, ok := tagVal.(string)
		if ok && tag != "" {
			// 首先根据标签名获取标签ID
			var tagModel model.TagM
			if err = db.Where("tag = ?", tag).First(&tagModel).Error; err == nil {
				// 通过JOIN查询关联表
				db = db.Joins("INNER JOIN article_tags ON article_models.id = article_tags.article_model_id").
					Where("article_tags.tag_model_id = ?", tagModel.ID)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("Failed to get tag by name", "err", err, "tag", tag)
				return 0, nil, errors.New("failed to get tag: " + err.Error())
			} else {
				// 标签不存在，返回空结果
				return 0, []*model.ArticleM{}, nil
			}
		}
	}

	// 处理收藏用户条件
	if favoritedVal, hasFavorited := specialConditions["favorited"]; hasFavorited {
		username, ok := favoritedVal.(string)
		if ok && username != "" {
			// 首先根据用户名获取用户ID
			var userModel model.UserM
			if err = db.Where("username = ?", username).First(&userModel).Error; err == nil {
				// 通过JOIN查询收藏表
				db = db.Joins("INNER JOIN favorite_models ON article_models.id = favorite_models.favorite_id").
					Where("favorite_models.favorite_by_id = ?", userModel.ID)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("Failed to get user by username", "err", err, "username", username)
				return 0, nil, errors.New("failed to get user: " + err.Error())
			} else {
				// 用户不存在，返回空结果
				return 0, []*model.ArticleM{}, nil
			}
		}
	}

	// 处理普通条件
	for key, value := range normalConditions {
		db = db.Where(key, value)
	}

	// 处理作者ID条件（单独处理，避免与JOIN冲突）
	if authorIDVal, hasAuthor := specialConditions["author_id"]; hasAuthor {
		db = db.Where("article_models.author_id = ?", authorIDVal)
	}

	// 先获取总数
	if err = db.Model(&model.ArticleM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count complex articles from database", "err", err, "conditions", conditions)
		return count, nil, errors.New("failed to count complex articles: " + err.Error())
	}

	// 再获取列表
	if err = db.Offset(offset).Limit(limit).Order("article_models.created_at desc").Find(&ret).Error; err != nil {
		slog.Error("Failed to list complex articles from database", "err", err, "conditions", conditions)
		return count, nil, errors.New("failed to list complex articles: " + err.Error())
	}

	// 4. 回写缓存
	if s.store.cache != nil && queryKey != "" {
		if cacheErr := s.store.cache.Article().SetArticleList(ctx, queryKey, count, ret); cacheErr != nil {
			slog.Warn("Failed to set complex article list to cache", "err", cacheErr, "query_key", queryKey)
		}
	}

	return
}

// ListWithCursor 使用游标获取文章列表（支持大数据量）
func (s *articleStore) ListWithCursor(ctx context.Context, condition interface{}, cursor int64, limit int) (hasMore bool, ret []*model.ArticleM, nextCursor int64, err error) {
	// 1. 生成查询键
	queryKey := ""
	if s.store.cache != nil {
		params := map[string]interface{}{
			"condition": condition,
			"cursor":    cursor,
			"limit":     limit,
		}
		queryKey = cache.GenerateQueryKey("cursor", params)

		// 2. 尝试从缓存获取
		articles, cachedHasMore, cacheErr := s.store.cache.Article().GetArticleCursorList(ctx, queryKey)
		if cacheErr == nil && articles != nil {
			slog.Debug("Article cursor list cache hit", "query_key", queryKey)
			hasMore = cachedHasMore
			if len(articles) > 0 {
				nextCursor = articles[len(articles)-1].ID
			}
			return hasMore, articles, nextCursor, nil
		}
		slog.Debug("Article cursor list cache miss", "query_key", queryKey)
	}

	// 3. 缓存未命中，查询数据库
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	// 使用游标替代Offset
	if cursor > 0 {
		db = db.Where("id < ?", cursor)
	}

	// 多查询一条记录来判断是否还有更多数据
	queryLimit := limit + 1

	// 获取列表
	if err = db.Order("id desc").Limit(queryLimit).Find(&ret).Error; err != nil {
		slog.Error("Failed to list articles with cursor from database", "err", err, "condition", condition, "cursor", cursor)
		return false, nil, 0, errors.New("failed to list articles with cursor: " + err.Error())
	}

	// 处理hasMore和nextCursor
	hasMore = len(ret) > limit
	if hasMore {
		// 移除多余的那一条记录
		ret = ret[:limit]
	}
	if len(ret) > 0 {
		nextCursor = ret[len(ret)-1].ID
	}

	// 4. 回写缓存
	if s.store.cache != nil && queryKey != "" {
		if cacheErr := s.store.cache.Article().SetArticleCursorList(ctx, queryKey, ret, hasMore); cacheErr != nil {
			slog.Warn("Failed to set article cursor list to cache", "err", cacheErr, "query_key", queryKey)
		}
	}

	return
}

// ListWithCursorAndCondition 使用游标和复合条件获取文章列表
func (s *articleStore) ListWithCursorAndCondition(ctx context.Context, conditions map[string]interface{}, cursor int64, limit int) (hasMore bool, ret []*model.ArticleM, nextCursor int64, err error) {
	// 1. 生成查询键
	queryKey := ""
	if s.store.cache != nil {
		params := map[string]interface{}{
			"conditions": conditions,
			"cursor":     cursor,
			"limit":      limit,
		}
		queryKey = cache.GenerateQueryKey("cursor_complex", params)

		// 2. 尝试从缓存获取
		articles, cachedHasMore, cacheErr := s.store.cache.Article().GetArticleCursorList(ctx, queryKey)
		if cacheErr == nil && articles != nil {
			slog.Debug("Complex article cursor list cache hit", "query_key", queryKey)
			hasMore = cachedHasMore
			if len(articles) > 0 {
				nextCursor = articles[len(articles)-1].ID
			}
			return hasMore, articles, nextCursor, nil
		}
		slog.Debug("Complex article cursor list cache miss", "query_key", queryKey)
	}

	// 3. 缓存未命中，查询数据库
	db := s.store.DB(ctx)

	// 分离特殊条件和普通条件
	var specialConditions = make(map[string]interface{})
	var normalConditions = make(map[string]interface{})

	for key, value := range conditions {
		if key == "tag" || key == "favorited" || key == "author_id" {
			specialConditions[key] = value
		} else {
			normalConditions[key] = value
		}
	}

	// 处理标签条件
	if tagVal, hasTag := specialConditions["tag"]; hasTag {
		tag, ok := tagVal.(string)
		if ok && tag != "" {
			var tagModel model.TagM
			if err = db.Where("tag = ?", tag).First(&tagModel).Error; err == nil {
				db = db.Joins("INNER JOIN article_tags ON article_models.id = article_tags.article_model_id").
					Where("article_tags.tag_model_id = ?", tagModel.ID)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("Failed to get tag by name", "err", err, "tag", tag)
				return false, nil, 0, errors.New("failed to get tag: " + err.Error())
			} else {
				return false, []*model.ArticleM{}, 0, nil
			}
		}
	}

	// 处理收藏用户条件
	if favoritedVal, hasFavorited := specialConditions["favorited"]; hasFavorited {
		username, ok := favoritedVal.(string)
		if ok && username != "" {
			var userModel model.UserM
			if err = db.Where("username = ?", username).First(&userModel).Error; err == nil {
				db = db.Joins("INNER JOIN favorite_models ON article_models.id = favorite_models.favorite_id").
					Where("favorite_models.favorite_by_id = ?", userModel.ID)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("Failed to get user by username", "err", err, "username", username)
				return false, nil, 0, errors.New("failed to get user: " + err.Error())
			} else {
				return false, []*model.ArticleM{}, 0, nil
			}
		}
	}

	// 处理普通条件
	for key, value := range normalConditions {
		db = db.Where(key, value)
	}

	// 处理作者ID条件
	if authorIDVal, hasAuthor := specialConditions["author_id"]; hasAuthor {
		db = db.Where("article_models.author_id = ?", authorIDVal)
	}

	// 使用游标替代Offset
	if cursor > 0 {
		db = db.Where("article_models.id < ?", cursor)
	}

	// 多查询一条记录来判断是否还有更多数据
	queryLimit := limit + 1

	// 获取列表
	if err = db.Order("article_models.id desc").Limit(queryLimit).Find(&ret).Error; err != nil {
		slog.Error("Failed to list complex articles with cursor from database", "err", err, "conditions", conditions, "cursor", cursor)
		return false, nil, 0, errors.New("failed to list complex articles with cursor: " + err.Error())
	}

	// 处理hasMore和nextCursor
	hasMore = len(ret) > limit
	if hasMore {
		ret = ret[:limit]
	}
	if len(ret) > 0 {
		nextCursor = ret[len(ret)-1].ID
	}

	// 4. 回写缓存
	if s.store.cache != nil && queryKey != "" {
		if cacheErr := s.store.cache.Article().SetArticleCursorList(ctx, queryKey, ret, hasMore); cacheErr != nil {
			slog.Warn("Failed to set complex article cursor list to cache", "err", cacheErr, "query_key", queryKey)
		}
	}

	return
}

// ListWithDeferredJoin 使用延迟关联优化深度分页
// 原理：先通过子查询获取目标页的主键ID（只走索引，不回表），再关联查询完整数据
// 性能：回表次数从 O(offset + limit) 降到 O(limit)，深度分页性能提升 10-100 倍
func (s *articleStore) ListWithDeferredJoin(ctx context.Context, condition interface{}, offset, limit int) (count int64, ret []*model.ArticleM, err error) {
	// 1. 生成查询键
	queryKey := ""
	if s.store.cache != nil {
		params := map[string]interface{}{
			"condition": condition,
			"offset":    offset,
			"limit":     limit,
			"method":    "deferred_join",
		}
		queryKey = cache.GenerateQueryKey("deferred", params)

		// 2. 尝试从缓存获取
		total, articles, cacheErr := s.store.cache.Article().GetArticleList(ctx, queryKey)
		if cacheErr == nil && articles != nil {
			slog.Debug("Article deferred join list cache hit", "query_key", queryKey)
			return total, articles, nil
		}
		slog.Debug("Article deferred join list cache miss", "query_key", queryKey)
	}

	// 3. 缓存未命中，查询数据库
	db := s.store.DB(ctx)
	if condition != nil {
		db = db.Where(condition)
	}

	// 3.1 先获取总数
	if err = db.Model(&model.ArticleM{}).Count(&count).Error; err != nil {
		slog.Error("Failed to count articles for deferred join", "err", err, "condition", condition)
		return count, nil, errors.New("failed to count articles: " + err.Error())
	}

	// 3.2 使用延迟关联查询列表
	// 第一步：子查询只获取 ID（走覆盖索引，不回表）
	var ids []int64
	subQuery := s.store.DB(ctx).Model(&model.ArticleM{})
	if condition != nil {
		subQuery = subQuery.Where(condition)
	}
	if err = subQuery.Select("id").Order("created_at DESC").Offset(offset).Limit(limit).Pluck("id", &ids).Error; err != nil {
		slog.Error("Failed to get article IDs with deferred join", "err", err, "condition", condition, "offset", offset)
		return count, nil, errors.New("failed to get article IDs: " + err.Error())
	}

	// 如果没有数据，直接返回空结果
	if len(ids) == 0 {
		return count, []*model.ArticleM{}, nil
	}

	// 第二步：根据 ID 批量获取完整数据（只回表 limit 次）
	// 使用 FIELD 函数保持原有的排序顺序（MySQL 特有语法）
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	orderClause := fmt.Sprintf("FIELD(id, %s)", strings.Join(idStrs, ","))

	if err = s.store.DB(ctx).Model(&model.ArticleM{}).
		Where("id IN ?", ids).
		Order(orderClause).
		Find(&ret).Error; err != nil {
		slog.Error("Failed to get articles by IDs with deferred join", "err", err, "ids", ids)
		return count, nil, errors.New("failed to get articles by IDs: " + err.Error())
	}

	// 4. 回写缓存
	if s.store.cache != nil && queryKey != "" {
		if cacheErr := s.store.cache.Article().SetArticleList(ctx, queryKey, count, ret); cacheErr != nil {
			slog.Warn("Failed to set article deferred join list to cache", "err", cacheErr, "query_key", queryKey)
		}
	}

	return
}
