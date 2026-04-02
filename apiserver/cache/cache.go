package cache

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// ICache 定义缓存层接口
type ICache interface {
	// 返回文章缓存操作接口
	Article() ArticleCache
	// 返回 JWT 缓存操作接口
	JWT() JWTCache
	// 返回监控指标收集器
	Metrics() *CacheMetricsCollector
	// 关闭缓存连接
	Close() error
}

// cacheManager 缓存管理器实现
type cacheManager struct {
	client         *redis.Client
	ristrettoCache *RistrettoCache
	articleCache   ArticleCache
	jwtCache       JWTCache
	metrics        *CacheMetricsCollector
}

// 确保 cacheManager 实现了 ICache 接口
var _ ICache = (*cacheManager)(nil)

// NewRedisClient 创建Redis客户端
func NewRedisClient(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "err", err, "addr", addr)
		return nil, err
	}

	slog.Info("Redis client connected successfully", "addr", addr, "db", db)
	return client, nil
}

// NewCache 创建缓存管理器实例
func NewCache(client *redis.Client, ctx context.Context) (ICache, error) {
	// 初始化 Ristretto 缓存
	ristrettoCache, err := NewRistrettoCache()
	if err != nil {
		return nil, err
	}

	// 初始化监控指标收集器
	metrics := NewCacheMetricsCollector()

	cm := &cacheManager{
		client:         client,
		ristrettoCache: ristrettoCache,
		metrics:        metrics,
	}

	// 初始化文章缓存
	cm.articleCache = newArticleCache(client, ristrettoCache, metrics)

	// 初始化 JWT 缓存
	cm.jwtCache = newJWTCache(client, ristrettoCache)

	slog.Info("Cache manager initialized successfully")

	return cm, nil
}

// Article 返回文章缓存操作接口
func (c *cacheManager) Article() ArticleCache {
	return c.articleCache
}

// JWT 返回 JWT 缓存操作接口
func (c *cacheManager) JWT() JWTCache {
	return c.jwtCache
}

// Metrics 返回监控指标收集器
func (c *cacheManager) Metrics() *CacheMetricsCollector {
	return c.metrics
}

// Close 关闭缓存连接
func (c *cacheManager) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
