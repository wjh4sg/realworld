package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// JWT黑名单缓存键前缀
	jwtBlacklistKeyPrefix = "jwt:blacklist:"
	// JWT验证缓存键前缀
	jwtValidationKeyPrefix = "jwt:validation:"
	// 黑名单过期时间（与JWT过期时间一致）
	jwtBlacklistTTL = 24 * time.Hour
	// 验证缓存过期时间
	jwtValidationTTL = 1 * time.Hour
)

// JWTCache 定义JWT缓存操作接口
type JWTCache interface {
	// AddToBlacklist 将令牌添加到黑名单
	AddToBlacklist(ctx context.Context, tokenID string) error
	// IsInBlacklist 检查令牌是否在黑名单中
	IsInBlacklist(ctx context.Context, tokenID string) (bool, error)
	// SetValidationResult 缓存令牌验证结果
	SetValidationResult(ctx context.Context, tokenHash string, userID int64) error
	// GetValidationResult 获取缓存的验证结果
	GetValidationResult(ctx context.Context, tokenHash string) (int64, error)
	// InvalidateValidationCache 使验证缓存失效
	InvalidateValidationCache(ctx context.Context, tokenHash string) error
}

// jwtCache JWT缓存实现
type jwtCache struct {
	client         *redis.Client
	ristrettoCache *RistrettoCache
}

// 确保 jwtCache 实现了 JWTCache 接口
var _ JWTCache = (*jwtCache)(nil)

// newJWTCache 创建JWT缓存实例
func newJWTCache(client *redis.Client, ristrettoCache *RistrettoCache) JWTCache {
	return &jwtCache{
		client:         client,
		ristrettoCache: ristrettoCache,
	}
}

// AddToBlacklist 将令牌添加到黑名单(同时添加到两级缓存)
func (c *jwtCache) AddToBlacklist(ctx context.Context, tokenID string) error {
	key := jwtBlacklistKeyPrefix + tokenID
	
	c.ristrettoCache.Set(key, true, 1)
	
	if err := c.client.Set(ctx, key, "1", jwtBlacklistTTL).Err(); err != nil {
		return err
	}
	
	return nil
}

// IsInBlacklist 检查令牌是否在黑名单中(三级缓存查询)
func (c *jwtCache) IsInBlacklist(ctx context.Context, tokenID string) (bool, error) {
	key := jwtBlacklistKeyPrefix + tokenID
	
	if val, ok := c.ristrettoCache.Get(key); ok {
		return val.(bool), nil
	}
	
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	
	result := val > 0
	
	c.ristrettoCache.Set(key, result, 1)
	
	return result, nil
}

// SetValidationResult 缓存令牌验证结果(同时设置到两级缓存)
func (c *jwtCache) SetValidationResult(ctx context.Context, tokenHash string, userID int64) error {
	key := jwtValidationKeyPrefix + tokenHash
	
	c.ristrettoCache.Set(key, userID, 1)
	
	return c.client.Set(ctx, key, userID, jwtValidationTTL).Err()
}

// GetValidationResult 获取缓存的验证结果(三级缓存查询)
func (c *jwtCache) GetValidationResult(ctx context.Context, tokenHash string) (int64, error) {
	key := jwtValidationKeyPrefix + tokenHash
	
	if val, ok := c.ristrettoCache.Get(key); ok {
		return val.(int64), nil
	}
	
	val, err := c.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	c.ristrettoCache.Set(key, val, 1)
	
	return val, nil
}

// InvalidateValidationCache 使验证缓存失效(同时删除两级缓存)
func (c *jwtCache) InvalidateValidationCache(ctx context.Context, tokenHash string) error {
	key := jwtValidationKeyPrefix + tokenHash
	
	c.ristrettoCache.Clear()
	
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	
	return nil
}
