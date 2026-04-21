package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/api/handlers"
	"mcp-control-hub/internal/api/middleware"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

func Setup(
	serverHandler *handlers.ServerHandler,
	toolHandler *handlers.ToolHandler,
	namespaceHandler *handlers.NamespaceHandler,
	apiKeyHandler *handlers.APIKeyHandler,
	configHandler *handlers.ConfigHandler,
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	// CORS middleware (still useful for MCP protocol clients)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-API-Key", "Mcp-Session-Id", "Mcp-Protocol-Version"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err == nil {
		r.StaticFS("/static", http.FS(staticFS))

		// Serve specific HTML files
		r.GET("/", func(c *gin.Context) {
			data, err := staticFiles.ReadFile("static/index.html")
			if err != nil {
				c.String(404, "Not found")
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
		})

		r.GET("/login.html", func(c *gin.Context) {
			data, err := staticFiles.ReadFile("static/login.html")
			if err != nil {
				c.String(404, "Not found")
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
		})
	}

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes (with auth)
	api := r.Group("/api/v1")
	api.Use(authMiddleware.Authenticate())
	{
		// Servers
		servers := api.Group("/servers")
		{
			servers.POST("", serverHandler.Create)
			servers.GET("", serverHandler.List)
			servers.GET("/:id", serverHandler.Get)
			servers.PUT("/:id", serverHandler.Update)
			servers.DELETE("/:id", serverHandler.Delete)
			servers.POST("/:id/enable", serverHandler.Enable)
			servers.POST("/:id/disable", serverHandler.Disable)
			servers.POST("/:id/sync", serverHandler.Sync)
		}

		// Tools
		tools := api.Group("/tools")
		{
			tools.GET("", toolHandler.List)
			tools.GET("/:id", toolHandler.Get)
			tools.PUT("/:id", toolHandler.Update)
			tools.POST("/:id/enable", toolHandler.Enable)
			tools.POST("/:id/disable", toolHandler.Disable)
			tools.POST("/refresh", toolHandler.Refresh)
		}

		// Namespaces
		namespaces := api.Group("/namespaces")
		{
			namespaces.POST("", namespaceHandler.Create)
			namespaces.GET("", namespaceHandler.List)
			namespaces.GET("/:id", namespaceHandler.Get)
			namespaces.PUT("/:id", namespaceHandler.Update)
			namespaces.DELETE("/:id", namespaceHandler.Delete)
			namespaces.POST("/:id/tools", namespaceHandler.AddTool)
			namespaces.DELETE("/:id/tools/:tool_id", namespaceHandler.RemoveTool)
			namespaces.GET("/:id/tools", namespaceHandler.ListTools)
		}

		// API Keys
		apiKeys := api.Group("/apikeys")
		{
			apiKeys.POST("", apiKeyHandler.Create)
			apiKeys.GET("", apiKeyHandler.List)
			apiKeys.DELETE("/:id", apiKeyHandler.Delete)
		}

		// Config
		config := api.Group("/config")
		{
			config.GET("", configHandler.List)
			config.PUT("/:key", configHandler.Update)
			config.PUT("", configHandler.UpdateAll)
		}
	}

	return r
}
