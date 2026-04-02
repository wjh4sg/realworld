package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/onexstack/realworld/apiserver/model"
)

const (
	// 文章详情缓存键前缀
	articleSlugKeyPrefix = "article:slug:"
	// 文章列表缓存键前缀
	articleListKeyPrefix = "article:list:"
	// 空标记值
	nullValue = "NULL"
	// 空标记过期时间(5分钟)
	nullTTL = 5 * time.Minute
	// 文章列表缓存基础过期时间(1小时)
	articleListBaseTTL = 1 * time.Hour
	// TTL随机偏移范围(0-10分钟)
	articleListTTLJitter = 10 * time.Minute
)

// ArticleCache 定义文章缓存操作接口
type ArticleCache interface {
	// GetArticle 从缓存获取文章详情
	GetArticle(ctx context.Context, slug string) (*model.ArticleM, error)
	// SetArticle 设置文章详情到缓存
	SetArticle(ctx context.Context, slug string, article *model.ArticleM) error
	// DeleteArticle 删除文章详情缓存
	DeleteArticle(ctx context.Context, slug string) error
	// SetArticleNotFound 设置文章不存在标记(防止缓存穿透)
	SetArticleNotFound(ctx context.Context, slug string) error
	// GetArticleList 从缓存获取文章列表
	GetArticleList(ctx context.Context, queryKey string) (int64, []*model.ArticleM, error)
	// SetArticleList 设置文章列表到缓存
	SetArticleList(ctx context.Context, queryKey string, total int64, articles []*model.ArticleM) error
	// GetArticleCursorList 从缓存获取游标分页文章列表
	GetArticleCursorList(ctx context.Context, queryKey string) ([]*model.ArticleM, bool, error)
	// SetArticleCursorList 设置游标分页文章列表到缓存
	SetArticleCursorList(ctx context.Context, queryKey string, articles []*model.ArticleM, hasMore bool) error
	// InvalidateArticleListCache 清除所有文章列表缓存
	InvalidateArticleListCache(ctx context.Context) error
}

// articleCache 文章缓存实现
type articleCache struct {
	client         *redis.Client
	ristrettoCache *RistrettoCache
	metrics        *CacheMetricsCollector
}

// 确保 articleCache 实现了 ArticleCache 接口
var _ ArticleCache = (*articleCache)(nil)

// newArticleCache 创建文章缓存实例
func newArticleCache(client *redis.Client, ristrettoCache *RistrettoCache, metrics *CacheMetricsCollector) ArticleCache {
	return &articleCache{
		client:         client,
		ristrettoCache: ristrettoCache,
		metrics:        metrics,
	}
}

// GetArticle 从缓存获取文章详情（三级缓存查询）
func (c *articleCache) GetArticle(ctx context.Context, slug string) (*model.ArticleM, error) {
	startTime := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.RecordLatency(time.Since(startTime))
		}
	}()

	key := articleSlugKeyPrefix + slug

	// 第一级：查询内存缓存
	if val, ok := c.ristrettoCache.Get(key); ok {
		if c.metrics != nil {
			c.metrics.RecordL1Hit()
		}
		return val.(*model.ArticleM), nil
	}
	if c.metrics != nil {
		c.metrics.RecordL1Miss()
	}

	// 第二级：查询 Redis 缓存
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 缓存未命中
			if c.metrics != nil {
				c.metrics.RecordL2Miss()
			}
			return nil, errors.New("cache miss")
		}
		slog.Error("Failed to get article from cache", "err", err, "slug", slug)
		return nil, err
	}
	if c.metrics != nil {
		c.metrics.RecordL2Hit()
	}

	// 检查是否是空标记
	if val == nullValue {
		return nil, errors.New("article not found")
	}

	// 反序列化文章对象
	var article model.ArticleM
	if err := json.Unmarshal([]byte(val), &article); err != nil {
		slog.Error("Failed to unmarshal article from cache", "err", err, "slug", slug)
		// 删除损坏的缓存
		c.client.Del(ctx, key)
		return nil, errors.New("cache corrupted")
	}

	// 回填内存缓存
	c.ristrettoCache.Set(key, &article, 1)

	return &article, nil
}

// SetArticle 设置文章详情到缓存(同时设置到两级缓存)
func (c *articleCache) SetArticle(ctx context.Context, slug string, article *model.ArticleM) error {
	key := articleSlugKeyPrefix + slug

	// 设置到内存缓存
	c.ristrettoCache.Set(key, article, 1)

	// 序列化文章对象
	data, err := json.Marshal(article)
	if err != nil {
		slog.Error("Failed to marshal article", "err", err, "slug", slug)
		return err
	}

	// 设置到Redis缓存,永不过期
	if err := c.client.Set(ctx, key, data, 0).Err(); err != nil {
		slog.Error("Failed to set article to cache", "err", err, "slug", slug)
		return err
	}

	slog.Debug("Article cached successfully", "slug", slug)
	return nil
}

// DeleteArticle 删除文章详情缓存 (同时删除两级缓存)
func (c *articleCache) DeleteArticle(ctx context.Context, slug string) error {
	key := articleSlugKeyPrefix + slug

	// 删除内存缓存（精细化删除，只删除指定键）
	c.ristrettoCache.Delete(key)

	// 删除 Redis 缓存
	if err := c.client.Del(ctx, key).Err(); err != nil {
		slog.Error("Failed to delete article from cache", "err", err, "slug", slug)
		return err
	}

	slog.Debug("Article cache deleted", "slug", slug)
	return nil
}

// SetArticleNotFound 设置文章不存在标记(防止缓存穿透)
func (c *articleCache) SetArticleNotFound(ctx context.Context, slug string) error {
	key := articleSlugKeyPrefix + slug

	// 设置空标记,5分钟过期
	if err := c.client.Set(ctx, key, nullValue, nullTTL).Err(); err != nil {
		slog.Error("Failed to set article not found mark", "err", err, "slug", slug)
		return err
	}

	slog.Debug("Article not found mark set", "slug", slug)
	return nil
}

// ArticleListCache 文章列表缓存数据结构
type ArticleListCache struct {
	Total    int64             `json:"total"`
	Articles []*model.ArticleM `json:"articles"`
	HasMore  bool              `json:"hasMore,omitempty"`
}

// GetArticleList 从缓存获取文章列表
func (c *articleCache) GetArticleList(ctx context.Context, queryKey string) (int64, []*model.ArticleM, error) {
	key := articleListKeyPrefix + queryKey

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 缓存未命中
			return 0, nil, errors.New("cache miss")
		}
		slog.Error("Failed to get article list from cache", "err", err, "queryKey", queryKey)
		return 0, nil, err
	}

	// 反序列化列表对象
	var listCache ArticleListCache
	if err := json.Unmarshal([]byte(val), &listCache); err != nil {
		slog.Error("Failed to unmarshal article list from cache", "err", err, "queryKey", queryKey)
		// 删除损坏的缓存
		c.client.Del(ctx, key)
		return 0, nil, errors.New("cache corrupted")
	}

	return listCache.Total, listCache.Articles, nil
}

// SetArticleList 设置文章列表到缓存
func (c *articleCache) SetArticleList(ctx context.Context, queryKey string, total int64, articles []*model.ArticleM) error {
	key := articleListKeyPrefix + queryKey

	// 构建缓存数据
	listCache := ArticleListCache{
		Total:    total,
		Articles: articles,
	}

	// 序列化列表对象
	data, err := json.Marshal(listCache)
	if err != nil {
		slog.Error("Failed to marshal article list", "err", err, "queryKey", queryKey)
		return err
	}

	// TTL随机化：基础时间 + 随机偏移，防止缓存雪崩
	jitter := time.Duration(rand.Int63n(int64(articleListTTLJitter)))
	ttl := articleListBaseTTL + jitter

	// 设置到缓存
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Error("Failed to set article list to cache", "err", err, "queryKey", queryKey)
		return err
	}

	slog.Debug("Article list cached successfully", "queryKey", queryKey, "ttl", ttl.String())
	return nil
}

// GetArticleCursorList 从缓存获取游标分页文章列表
func (c *articleCache) GetArticleCursorList(ctx context.Context, queryKey string) ([]*model.ArticleM, bool, error) {
	key := articleListKeyPrefix + queryKey

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, errors.New("cache miss")
		}
		slog.Error("Failed to get article cursor list from cache", "err", err, "queryKey", queryKey)
		return nil, false, err
	}

	var listCache ArticleListCache
	if err := json.Unmarshal([]byte(val), &listCache); err != nil {
		slog.Error("Failed to unmarshal article cursor list from cache", "err", err, "queryKey", queryKey)
		c.client.Del(ctx, key)
		return nil, false, errors.New("cache corrupted")
	}

	return listCache.Articles, listCache.HasMore, nil
}

// SetArticleCursorList 设置游标分页文章列表到缓存
func (c *articleCache) SetArticleCursorList(ctx context.Context, queryKey string, articles []*model.ArticleM, hasMore bool) error {
	key := articleListKeyPrefix + queryKey

	listCache := ArticleListCache{
		Articles: articles,
		HasMore:  hasMore,
	}

	data, err := json.Marshal(listCache)
	if err != nil {
		slog.Error("Failed to marshal article cursor list", "err", err, "queryKey", queryKey)
		return err
	}

	// TTL随机化：基础时间 + 随机偏移，防止缓存雪崩
	jitter := time.Duration(rand.Int63n(int64(articleListTTLJitter)))
	ttl := articleListBaseTTL + jitter

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Error("Failed to set article cursor list to cache", "err", err, "queryKey", queryKey)
		return err
	}

	slog.Debug("Article cursor list cached successfully", "queryKey", queryKey, "hasMore", hasMore, "ttl", ttl.String())
	return nil
}

// InvalidateArticleListCache 清除所有文章列表缓存
func (c *articleCache) InvalidateArticleListCache(ctx context.Context) error {
	// 使用SCAN命令查找所有列表缓存键
	pattern := articleListKeyPrefix + "*"

	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			slog.Error("Failed to scan article list cache keys", "err", err)
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	// 批量删除
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			slog.Error("Failed to delete article list caches", "err", err, "count", len(keys))
			return err
		}
		slog.Debug("Article list caches invalidated", "count", len(keys))
	}

	return nil
}

// GenerateQueryKey 生成查询键(根据查询条件生成唯一标识)
func GenerateQueryKey(queryType string, params map[string]interface{}) string {
	// 将参数序列化为JSON并计算哈希
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	// 使用完整的SHA256哈希值增强唯一性
	return fmt.Sprintf("%s:%x", queryType, hash)
}
