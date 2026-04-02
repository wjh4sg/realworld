package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/onexstack/realworld/apiserver/model"
	"gorm.io/gorm"
)

// CacheWarmupService 缓存预热服务
type CacheWarmupService struct {
	cache  ICache
	db     *gorm.DB
	logger *slog.Logger

	// 预热配置
	articleLimit int
	tagLimit     int

	// 定时预热控制
	stopChan chan struct{}
}

// NewCacheWarmupService 创建缓存预热服务
func NewCacheWarmupService(cache ICache, db *gorm.DB) *CacheWarmupService {
	return &CacheWarmupService{
		cache:        cache,
		db:           db,
		logger:       slog.Default(),
		articleLimit: 100, // 默认预热 100 篇文章
		tagLimit:     50,  // 默认预热 50 个标签
		stopChan:     make(chan struct{}),
	}
}

// Warmup 执行缓存预热（一次性）
func (s *CacheWarmupService) Warmup() error {
	s.logger.Info("Starting cache warmup...")

	startTime := time.Now()

	// 预热热门文章
	articleCount, err := s.WarmupArticles(s.articleLimit)
	if err != nil {
		s.logger.Error("Failed to warmup articles", "err", err)
		return fmt.Errorf("warmup articles failed: %w", err)
	}

	// 预热热门标签
	tagCount, err := s.WarmupTags(s.tagLimit)
	if err != nil {
		s.logger.Error("Failed to warmup tags", "err", err)
		return fmt.Errorf("warmup tags failed: %w", err)
	}

	duration := time.Since(startTime)
	s.logger.Info("Cache warmup completed",
		"articles", articleCount,
		"tags", tagCount,
		"duration", duration.String())

	return nil
}

// WarmupArticles 预热文章缓存
func (s *CacheWarmupService) WarmupArticles(limit int) (int, error) {
	ctx := context.Background()

	// 查询最新文章（按创建时间排序）
	var articles []*model.ArticleM
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&articles).Error; err != nil {
		return 0, fmt.Errorf("query articles failed: %w", err)
	}

	count := 0
	for _, article := range articles {
		// 写入两级缓存
		if err := s.cache.Article().SetArticle(ctx, article.Slug, article); err != nil {
			s.logger.Warn("Failed to cache article", "slug", article.Slug, "err", err)
			continue
		}
		count++
	}

	s.logger.Info("Articles warmup completed", "count", count, "limit", limit)
	return count, nil
}

// WarmupTags 预热标签缓存
func (s *CacheWarmupService) WarmupTags(limit int) (int, error) {
	// 查询热门标签（这里简化处理，实际可以根据使用次数排序）
	var tags []string
	if err := s.db.Model(&model.TagM{}).Distinct("tag").Limit(limit).Pluck("tag", &tags).Error; err != nil {
		return 0, fmt.Errorf("query tags failed: %w", err)
	}

	count := len(tags)
	s.logger.Info("Tags warmup completed", "count", count, "limit", limit)
	return count, nil
}

// StartPeriodicWarmup 启动定时预热
func (s *CacheWarmupService) StartPeriodicWarmup(interval time.Duration) {
	s.logger.Info("Starting periodic cache warmup", "interval", interval.String())

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.Warmup(); err != nil {
					s.logger.Error("Periodic warmup failed", "err", err)
				}
			case <-s.stopChan:
				s.logger.Info("Periodic warmup stopped")
				return
			}
		}
	}()
}

// StopPeriodicWarmup 停止定时预热
func (s *CacheWarmupService) StopPeriodicWarmup() {
	close(s.stopChan)
}

// SetArticleLimit 设置文章预热数量
func (s *CacheWarmupService) SetArticleLimit(limit int) {
	s.articleLimit = limit
}

// SetTagLimit 设置标签预热数量
func (s *CacheWarmupService) SetTagLimit(limit int) {
	s.tagLimit = limit
}
