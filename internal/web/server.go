package web

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/internal/services/speaker"
	"mi-gpt-go/pkg/logger"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DBConfigService 数据库配置服务接口
type DBConfigService interface {
	LoadConfig() (*config.Config, error)
	SaveConfig(*config.Config) error
}

// WebServer Web服务器
type WebServer struct {
	router           *gin.Engine
	server           *http.Server
	config           *config.Config
	aiSpeaker        *speaker.EnhancedAISpeaker
	dbConfigService  DBConfigService
	isRunning        bool
}

// NewWebServer 创建Web服务器
func NewWebServer(cfg *config.Config, aiSpeaker *speaker.EnhancedAISpeaker, dbConfigService DBConfigService) *WebServer {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	
	// 配置CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	ws := &WebServer{
		router:          router,
		config:          cfg,
		aiSpeaker:       aiSpeaker,
		dbConfigService: dbConfigService,
	}

	// 设置路由
	ws.setupRoutes()

	return ws
}

// setupRoutes 设置路由
func (ws *WebServer) setupRoutes() {
	// 静态文件服务 - 使用嵌入的文件系统提供Vue前端构建文件
	if HasStaticFiles() {
		// 使用嵌入的静态文件
		ws.router.StaticFS("/assets", GetAssetsFS()) // 直接映射到assets子目录
		ws.router.GET("/vite.svg", ws.handleViteSVG)  // 单独处理vite.svg
		ws.router.GET("/", ws.handleEmbedIndex)
		ws.router.NoRoute(ws.handleSPA) // SPA路由回退
	} else {
		// 开发模式：回退到文件系统
		ws.router.Static("/static", "./internal/web/static")
		ws.router.Static("/assets", "./internal/web/static/assets")
		ws.router.GET("/", ws.handleIndex)
	}
	
	// 前端测试页面
	ws.router.GET("/test", ws.handleTestPage)

	// API路由组
	api := ws.router.Group("/api/v1")
	{
		// 配置管理
		config := api.Group("/config")
		{
			config.GET("/", ws.getConfig)
			config.PUT("/", ws.updateConfig)
			config.POST("/reload", ws.reloadConfig)
			config.GET("/template", ws.getConfigTemplate)
			config.POST("/validate", ws.validateConfig)
			config.POST("/test-ai", ws.testAIConnection)
			config.POST("/test-mi", ws.testMiConnection)
			config.POST("/start-speaker", ws.startSpeaker)
		}

		// 系统状态
		system := api.Group("/system")
		{
			system.GET("/status", ws.getSystemStatus)
			system.POST("/restart", ws.restartSystem)
			system.GET("/logs", ws.getSystemLogs)
			system.DELETE("/logs", ws.clearLogs)
			system.PUT("/log-level", ws.setLogLevel)
			system.GET("/log-level", ws.getLogLevel)
		}

		// 音箱控制
		speaker := api.Group("/speaker")
		{
			speaker.GET("/status", ws.getSpeakerStatus)
			speaker.POST("/play", ws.playTTS)
			speaker.POST("/stop", ws.stopSpeaker)
			speaker.POST("/restart", ws.restartSpeaker)  // 配置热重载端点
			speaker.POST("/execute", ws.executeVoiceCommand)  // 语音命令执行端点
		}

		// 并发处理状态
		concurrent := api.Group("/concurrent")
		{
			concurrent.GET("/status", ws.getConcurrentStatus)
			concurrent.PUT("/config", ws.updateConcurrentConfig)
		}
	}
}

// handleIndex 首页处理（开发模式）
func (ws *WebServer) handleIndex(c *gin.Context) {
	// 从文件系统读取Vue应用的index.html
	indexHTML, err := os.ReadFile("internal/web/static/index.html")
	if err != nil {
		// 如果Vue应用不可用，返回简单的状态页面
		logger.Warn("Vue应用index.html未找到，返回状态页面")
		ws.handleStatusPage(c)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
}

// handleEmbedIndex 首页处理（生产模式，使用嵌入文件）
func (ws *WebServer) handleEmbedIndex(c *gin.Context) {
	indexHTML, err := GetIndexHTML()
	if err != nil {
		logger.Warn("嵌入的Vue应用index.html未找到，返回状态页面")
		ws.handleStatusPage(c)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
}

// handleSPA SPA路由回退处理（用于Vue Router的history模式）
func (ws *WebServer) handleSPA(c *gin.Context) {
	// 对于非API路径且非静态资源路径，返回index.html
	path := c.Request.URL.Path
	if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/assets") && path != "/vite.svg" {
		ws.handleEmbedIndex(c)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "路径未找到"})
}

// handleViteSVG 处理vite.svg图标请求
func (ws *WebServer) handleViteSVG(c *gin.Context) {
	svgContent, err := GetViteSVG()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图标未找到"})
		return
	}
	c.Data(http.StatusOK, "image/svg+xml", svgContent)
}

// handleStatusPage 状态页面处理（备用）
func (ws *WebServer) handleStatusPage(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MiGPT Go 管理面板</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
        }
        .container {
            text-align: center;
            background: rgba(255, 255, 255, 0.1);
            padding: 40px;
            border-radius: 20px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        h1 { margin: 0 0 20px 0; font-size: 2.5em; }
        p { margin: 10px 0; font-size: 1.2em; opacity: 0.9; }
        .status { margin: 20px 0; }
        .status-item { 
            display: inline-block; 
            margin: 5px 15px; 
            padding: 8px 16px; 
            background: rgba(255, 255, 255, 0.2); 
            border-radius: 10px; 
        }
        .btn {
            display: inline-block;
            margin: 20px 10px;
            padding: 12px 24px;
            background: rgba(255, 255, 255, 0.2);
            color: white;
            text-decoration: none;
            border-radius: 10px;
            transition: all 0.3s ease;
            border: 1px solid rgba(255, 255, 255, 0.3);
        }
        .btn:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎉 MiGPT Go 管理面板</h1>
        <p>增强版 - 支持多线程处理 + Web管理</p>
        <div class="status">
            <div class="status-item">🌐 Web服务正常运行</div>
            <div class="status-item">⚙️ 管理面板已就绪</div>
        </div>
        <p>Vue前端管理界面已集成，请刷新页面！</p>
        <a href="/api/v1/system/status" class="btn">📊 系统状态API</a>
        <a href="/api/v1/config" class="btn">⚙️ 配置API</a>
        <p style="margin-top: 30px; opacity: 0.7; font-size: 0.9em;">
            API地址: http://0.0.0.0:8080/api/v1<br>
            外网访问: http://[你的IP]:8080<br>
            项目地址: <a href="https://github.com/shinegod/mi-gpt-go" style="color: #fff;">GitHub</a>
        </p>
    </div>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// handleTestPage 测试页面处理
func (ws *WebServer) handleTestPage(c *gin.Context) {
	// 读取测试页面文件
	content, err := os.ReadFile("frontend/test.html")
	if err != nil {
		c.String(http.StatusNotFound, "测试页面未找到")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

// Start 启动Web服务器
func (ws *WebServer) Start(port int) error {
	if ws.isRunning {
		return fmt.Errorf("Web服务器已在运行")
	}

	address := fmt.Sprintf(":%d", port)
	ws.server = &http.Server{
		Addr:    address,
		Handler: ws.router,
	}

	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Web服务器启动失败: %v", err)
		}
	}()

	ws.isRunning = true
	logger.Infof("🌐 Web管理面板已启动: http://0.0.0.0%s", address)
	logger.Infof("🌍 外网访问地址: http://[你的IP]%s", address)
	return nil
}

// Stop 停止Web服务器
func (ws *WebServer) Stop() error {
	if !ws.isRunning {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("关闭Web服务器失败: %v", err)
	}

	ws.isRunning = false
	logger.Info("Web服务器已停止")
	return nil
}

// IsRunning 检查是否运行中
func (ws *WebServer) IsRunning() bool {
	return ws.isRunning
}

// SetAISpeaker 设置AI音箱服务
func (ws *WebServer) SetAISpeaker(aiSpeaker *speaker.EnhancedAISpeaker) {
	ws.aiSpeaker = aiSpeaker
}

// CreateAISpeaker 创建AI音箱服务
func (ws *WebServer) CreateAISpeaker() error {
	if !ws.config.IsConfigured() {
		return fmt.Errorf("配置不完整，无法创建AI音箱服务")
	}
	
	// 停止现有服务
	if ws.aiSpeaker != nil {
		if err := ws.aiSpeaker.Stop(); err != nil {
			logger.Warn("停止现有AI音箱服务失败:", err)
		}
	}
	
	// 创建新服务
	var err error
	ws.aiSpeaker, err = speaker.NewEnhancedAISpeaker(ws.config)
	if err != nil {
		return fmt.Errorf("创建AI音箱服务失败: %v", err)
	}
	
	// 启动服务
	if err := ws.aiSpeaker.Start(); err != nil {
		return fmt.Errorf("启动AI音箱服务失败: %v", err)
	}
	
	logger.Info("AI音箱服务创建并启动成功")
	return nil
} 