package http

import (
	"context"
	"net/http"

	"kv-storage/internal/config"
	"kv-storage/internal/interfaces"
	"kv-storage/internal/service"
	"kv-storage/internal/transport/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	engine  *gin.Engine
	server  *http.Server
	logger  interfaces.Logger
	config  *config.Config
	service *service.KVService
}

func NewRouter(cfg *config.Config, logger interfaces.Logger, kvService *service.KVService) interfaces.Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	rateLimiter := middleware.NewRateLimiter(100, 200, logger)

	engine.Use(
		gin.Recovery(),
		middleware.Logger(logger),
		cors.Default(),
		rateLimiter.RateLimit(),
	)

	router := &Router{
		engine:  engine,
		logger:  logger,
		config:  cfg,
		service: kvService,
	}

	router.setupRoutes()

	router.server = &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      engine,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
	}

	return router
}

func (r *Router) setupRoutes() {
	api := r.engine.Group("/api/v1")
	{
		kv := api.Group("/kv")
		{
			handler := NewHandler(r.service, r.logger)
			kv.POST("", handler.Create)
			kv.GET("", handler.List)
			kv.GET("/all", handler.ListIncludingDeleted)
			kv.GET("/:key", handler.Get)
			kv.PUT("/:key", handler.Update)
			kv.DELETE("/:key", handler.Delete)
			kv.POST("/:key/restore", handler.Restore)
		}
	}

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (r *Router) Run(addr string) error {
	r.logger.Info("Starting HTTP server", "addr", addr)
	return r.server.ListenAndServe()
}

func (r *Router) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down HTTP server")
	return r.server.Shutdown(ctx)
}
