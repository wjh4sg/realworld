package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/cache"
	"github.com/onexstack/realworld/apiserver/handler"
	"github.com/onexstack/realworld/apiserver/jwt"
	"github.com/onexstack/realworld/apiserver/middleware"
	mockstore "github.com/onexstack/realworld/apiserver/mock/store"
	"github.com/onexstack/realworld/apiserver/monitoring"
	"github.com/onexstack/realworld/apiserver/store"
	"github.com/onexstack/realworld/apiserver/ui"
	common "github.com/onexstack/realworld/common"
	config "github.com/onexstack/realworld/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type serverDeps struct {
	db        *gorm.DB
	redis     *redis.Client
	useMock   bool
	cache     cache.ICache
	storeInst store.IStore
}

func main() {
	useMock := os.Getenv("USE_MOCK") == "1"
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := defaultConfig()
	if !useMock {
		cfg = config.GetConfig()
	}
	deps, jwtSecret, port := initializeDependencies(cfg, useMock)

	bizLayer := biz.NewBiz(deps.storeInst)

	var jwtCache cache.JWTCache
	if deps.cache != nil {
		jwtCache = deps.cache.JWT()
	}
	jwtManager := jwt.NewManager(jwtSecret, jwtCache)

	concurrencyMonitor := monitoring.NewConcurrencyMonitor()
	h := handler.NewHandler(bizLayer, jwtManager)

	router := gin.Default()
	router.Use(middleware.Cors())

	setupUIRoutes(router)
	setupRoutes(router, h, jwtManager, concurrencyMonitor, deps, cfg)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Server is running on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func defaultConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Server.Port = "8080"
	cfg.Server.RateLimitPerMinute = 60
	cfg.Server.RateLimitPerHour = 1000
	cfg.Server.RateLimitRPS = 1000
	return cfg
}

func initializeDependencies(cfg *config.Config, useMock bool) (*serverDeps, string, string) {
	if useMock {
		log.Println("Starting in mock mode")
		return &serverDeps{
			useMock:   true,
			storeInst: mockstore.NewMockStore(),
		}, "mock-jwt-secret-for-testing", cfg.Server.Port
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	redisClient, err := cache.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	cacheManager, err := cache.NewCache(redisClient, context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize cache manager: %v", err)
	}

	warmupService := cache.NewCacheWarmupService(cacheManager, db)
	if err := warmupService.Warmup(); err != nil {
		log.Printf("Warning: cache warmup failed: %v", err)
	}

	storeInst := store.NewStore(db, cacheManager)
	return &serverDeps{
		db:        db,
		redis:     redisClient,
		cache:     cacheManager,
		storeInst: storeInst,
	}, cfg.JWT.Secret, cfg.Server.Port
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	mysqlOpts := &common.MySQLOptions{
		Addr:                  cfg.MySQL.Addr,
		Username:              cfg.MySQL.Username,
		Password:              cfg.MySQL.Password,
		Database:              cfg.MySQL.Database,
		MaxIdleConnections:    cfg.MySQL.MaxIdleConnections,
		MaxOpenConnections:    cfg.MySQL.MaxOpenConnections,
		MaxConnectionLifeTime: cfg.MySQL.MaxConnectionLifeTime,
	}

	return common.NewMySQL(mysqlOpts)
}

func setupUIRoutes(r *gin.Engine) {
	uiFS, err := ui.StaticFS()
	if err != nil {
		log.Printf("Failed to load embedded UI: %v", err)
		return
	}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui/")
	})

	r.StaticFS("/ui", http.FS(uiFS))
}

func setupRoutes(r *gin.Engine, h *handler.Handler, jwtManager *jwt.Manager, concurrencyMonitor *monitoring.ConcurrencyMonitor, deps *serverDeps, cfg *config.Config) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		if deps.useMock {
			c.JSON(http.StatusOK, gin.H{"status": "ready", "mode": "mock"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := common.MustRawDB(deps.db).PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "mysql", "error": err.Error()})
			return
		}

		if err := deps.redis.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "redis", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	api := r.Group("/api")
	api.Use(middleware.RateLimiter(cfg.Server.RateLimitPerMinute, cfg.Server.RateLimitPerHour))

	api.POST("/users", h.User().Register)
	api.POST("/users/login", h.User().Login)
	api.POST("/users/refresh", h.User().RefreshToken)

	optionalAuth := api.Group("/")
	optionalAuth.Use(middleware.OptionalAuth(jwtManager))
	{
		optionalAuth.GET("/profiles/:username", h.User().GetProfile)
		optionalAuth.GET("/articles", h.Article().GetArticles)
		optionalAuth.GET("/articles/cursor", h.Article().GetArticlesWithCursor)
		optionalAuth.GET("/articles/optimized", h.Article().GetArticlesWithDeferredJoin)
		optionalAuth.GET("/articles/:slug", h.Article().GetArticle)
		optionalAuth.GET("/articles/:slug/comments", h.Comment().GetCommentsByArticle)
		optionalAuth.GET("/tags", h.Tag().GetTags)
	}

	auth := api.Group("/")
	auth.Use(middleware.Auth(jwtManager))
	{
		auth.GET("/user", h.User().GetCurrentUser)
		auth.PUT("/user", h.User().UpdateUser)
		auth.POST("/profiles/:username/follow", h.User().FollowUser)
		auth.DELETE("/profiles/:username/follow", h.User().UnfollowUser)

		auth.GET("/articles/feed", h.Article().GetFeed)
		auth.POST("/articles", h.Article().CreateArticle)
		auth.PUT("/articles/:slug", h.Article().UpdateArticle)
		auth.DELETE("/articles/:slug", h.Article().DeleteArticle)
		auth.POST("/articles/:slug/favorite", h.Article().FavoriteArticle)
		auth.DELETE("/articles/:slug/favorite", h.Article().UnfavoriteArticle)

		auth.POST("/articles/:slug/comments", h.Comment().CreateComment)
		auth.DELETE("/articles/:slug/comments/:id", h.Comment().DeleteComment)
		auth.DELETE("/comments/:id", h.Comment().DeleteComment)
	}

	r.GET("/metrics/concurrency", func(c *gin.Context) {
		concurrencyMonitor.HandleMetrics(c.Writer, c.Request)
	})

	r.GET("/metrics", func(c *gin.Context) {
		concurrencyMonitor.HandleAllMetrics(c.Writer, c.Request)
	})

	if !deps.useMock && deps.storeInst.Cache() != nil {
		r.GET("/metrics/cache", func(c *gin.Context) {
			metrics := deps.storeInst.Cache().Metrics().ExportMetrics()
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   metrics,
			})
		})
	}
}
