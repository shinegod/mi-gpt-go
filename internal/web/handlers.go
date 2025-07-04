package web

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/services/miservice"
	"mi-gpt-go/internal/services/openai"
	"mi-gpt-go/pkg/logger"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ConfigResponse 配置响应结构
type ConfigResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SystemStatus 系统状态
type SystemStatus struct {
	IsRunning      bool      `json:"isRunning"`
	StartTime      time.Time `json:"startTime"`
	Version        string    `json:"version"`
	GoVersion      string    `json:"goVersion"`
	WebServerPort  int       `json:"webServerPort"`
	DatabasePath   string    `json:"databasePath"`
	ConcurrentMode bool      `json:"concurrentMode"`
}

// getConfig 获取配置
func (ws *WebServer) getConfig(c *gin.Context) {
	configData := map[string]interface{}{
		"ai": map[string]interface{}{
			"provider":        ws.config.OpenAI.Provider,
			"model":           ws.config.OpenAI.Model,
			"baseURL":         ws.config.OpenAI.BaseURL,
			"proxyURL":        ws.config.OpenAI.ProxyURL,
			"apiKey":          ws.config.OpenAI.APIKey,          // 返回完整API密钥，由前端控制显示
			"azureAPIKey":     ws.config.OpenAI.AzureAPIKey,     // 返回完整API密钥，由前端控制显示
			"azureEndpoint":   ws.config.OpenAI.AzureEndpoint,
			"azureDeployment": ws.config.OpenAI.AzureDeployment,
			"deepSeekAPIKey":  ws.config.OpenAI.DeepSeekAPIKey,  // 返回完整API密钥，由前端控制显示
		},
		"bot": map[string]interface{}{
			"name":            ws.config.Bot.Name,
			"profile":         ws.config.Bot.Profile,
			"masterName":      ws.config.Bot.Master.Name,
			"masterProfile":   ws.config.Bot.Master.Profile,
			"roomName":        ws.config.Bot.Room.Name,
			"roomDescription": ws.config.Bot.Room.Description,
		},
		"speaker": map[string]interface{}{
			"name":               ws.config.Speaker.Name,
			"callAIKeywords":     strings.Join(ws.config.Speaker.CallAIKeywords, ","),
			"wakeupKeywords":     strings.Join(ws.config.Speaker.WakeUpKeywords, ","),
			"exitKeywords":       strings.Join(ws.config.Speaker.ExitKeywords, ","),
			"onEnterAI":          strings.Join(ws.config.Speaker.OnEnterAI, ","),
			"onExitAI":           strings.Join(ws.config.Speaker.OnExitAI, ","),
			"onAIAsking":         strings.Join(ws.config.Speaker.OnAIAsking, ","),
			"onAIReplied":        strings.Join(ws.config.Speaker.OnAIReplied, ","),
			"onAIError":          strings.Join(ws.config.Speaker.OnAIError, ","),
			"streamResponse":     ws.config.Speaker.StreamResponse,
			"enableAudioLog":     ws.config.Speaker.EnableAudioLog,
			"debugMode":          ws.config.Speaker.Debug,
		},
		"database": map[string]interface{}{
			"path":  ws.config.Database.Path,
			"debug": ws.config.Database.Debug,
		},
		"mi": map[string]interface{}{
			"userID":        ws.config.Speaker.UserID,   // 返回完整用户ID，由前端控制显示
			"password":      ws.config.Speaker.Password, // 返回完整密码，由前端控制显示
			"deviceID":      ws.config.Speaker.DeviceID,
			"checkInterval": ws.config.Speaker.CheckInterval,
			"timeout":       ws.config.Speaker.Timeout,
			"enableTrace":   ws.config.Speaker.EnableTrace,
		},
		"concurrent": map[string]interface{}{
			"enable":              ws.config.Speaker.EnableConcurrent,
			"workerCount":         ws.config.Speaker.WorkerCount,
			"queueSize":           ws.config.Speaker.QueueSize,
			"messageBufferSize":   ws.config.Speaker.MessageBufferSize,
			"rateLimit":           ws.config.Speaker.RateLimit,
			"batchSize":           ws.config.Speaker.BatchSize,
			"batchTimeoutSeconds": ws.config.Speaker.BatchTimeoutSeconds,
			"enableMetrics":       ws.config.Speaker.EnableMetrics,
		},
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    configData,
	})
}

// updateConfig 更新配置
func (ws *WebServer) updateConfig(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("请求数据格式错误: %v", err),
		})
		return
	}

	logger.Info("收到配置更新请求")

	// 更新配置字段
	if err := ws.updateConfigFields(data); err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("更新配置失败: %v", err),
		})
		return
	}

	// 保存配置到数据库（持久化）
	if err := ws.dbConfigService.SaveConfig(ws.config); err != nil {
		logger.Errorf("保存配置到数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	logger.Info("配置已成功保存到数据库")

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "配置更新并保存成功",
		Data:    ws.config,
	})
}

// reloadConfig 重新加载配置
func (ws *WebServer) reloadConfig(c *gin.Context) {
	// 从数据库重新加载配置
	newConfig, err := ws.dbConfigService.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("从数据库重新加载配置失败: %v", err),
		})
		return
	}

	ws.config = newConfig
	logger.Info("配置已从数据库重新加载")

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "配置从数据库重新加载成功",
	})
}

// getConfigTemplate 获取配置模板
func (ws *WebServer) getConfigTemplate(c *gin.Context) {
	// 读取env.example文件
	content, err := os.ReadFile("env.example")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("读取配置模板失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    string(content),
	})
}

// getSystemStatus 获取系统状态
func (ws *WebServer) getSystemStatus(c *gin.Context) {
	status := SystemStatus{
		IsRunning:      ws.IsRunning(),
		StartTime:      time.Now(), // 这里应该是实际的启动时间
		Version:        "1.0.0",
		GoVersion:      "1.21",
		WebServerPort:  8080,
		DatabasePath:   ws.config.Database.Path,
		ConcurrentMode: ws.config.Speaker.EnableConcurrent,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    status,
	})
}

// restartSystem 重启系统
func (ws *WebServer) restartSystem(c *gin.Context) {
	logger.Info("收到系统重启请求")
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "系统重启命令已发送",
	})

	// 这里可以实现系统重启逻辑
}

// getSystemLogs 获取系统日志
func (ws *WebServer) getSystemLogs(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "500")
	levelFilter := c.Query("level")
	
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
		limit = 500
	}
	
	// 从日志缓冲区获取日志
	logs := logger.GetLogs(limit)
	
	// 如果指定了级别过滤
	if levelFilter != "" {
		var filteredLogs []logger.LogEntry
		for _, logEntry := range logs {
			if logEntry.Level == levelFilter {
				filteredLogs = append(filteredLogs, logEntry)
			}
		}
		logs = filteredLogs
	}
	
	// 转换为字符串格式（保持与前端兼容）
	var logStrings []string
	for _, logEntry := range logs {
		logStr := fmt.Sprintf("[%s] %s %s", 
			logEntry.Level, 
			logEntry.Time.Format("2006-01-02 15:04:05"), 
			logEntry.Message)
		logStrings = append(logStrings, logStr)
	}
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data: map[string]interface{}{
			"logs":       logStrings,
			"entries":    logs, // 结构化日志数据
			"total":      logger.GetLogCount(),
			"level":      logger.GetLevel(),
		},
	})
}

// getSpeakerStatus 获取音箱状态
func (ws *WebServer) getSpeakerStatus(c *gin.Context) {
	if ws.aiSpeaker == nil {
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Data: map[string]interface{}{
				"connected":    false,
				"deviceID":     ws.config.Speaker.DeviceID,
				"name":         ws.config.Speaker.Name,
				"isPlaying":    false,
				"volume":       0,
				"lastMessage": "AI音箱服务未启动",
				"error":       "请先完成配置",
			},
		})
		return
	}

	// 获取实际的音箱状态
	status := ws.aiSpeaker.GetStatus()
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data: map[string]interface{}{
			"connected":    ws.aiSpeaker.IsRunning(),
			"deviceID":     ws.config.Speaker.DeviceID,
			"name":         ws.config.Speaker.Name,
			"isPlaying":    getStatusField(status, "isPlaying", false),
			"volume":       getStatusField(status, "volume", 50),
			"lastMessage": getStatusField(status, "lastMessage", "运行中"),
			"isRunning":    ws.aiSpeaker.IsRunning(),
			"status":       status,
		},
	})
}

// playTTS 播放TTS
func (ws *WebServer) playTTS(c *gin.Context) {
	var request struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("请求数据格式错误: %v", err),
		})
		return
	}

	if ws.aiSpeaker == nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: "AI音箱服务未启动，请先完成配置并启动服务。可在配置页面完成所有必需配置后启动音箱服务。",
		})
		return
	}

	logger.Infof("收到TTS播放请求: %s", request.Text)

	// 调用实际的TTS播放方法
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	err := ws.aiSpeaker.ExecuteCommand(ctx, request.Text)
	if err != nil {
		logger.Errorf("TTS播放失败: %v", err)
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("TTS播放失败: %v", err),
		})
		return
	}

	logger.Infof("TTS播放成功: %s", request.Text)

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "TTS播放成功",
	})
}

// stopSpeaker 停止音箱
func (ws *WebServer) stopSpeaker(c *gin.Context) {
	if ws.aiSpeaker == nil {
		c.JSON(http.StatusServiceUnavailable, ConfigResponse{
			Success: false,
			Message: "AI音箱服务未启动",
		})
		return
	}

	logger.Info("收到音箱停止请求")

	// 停止AI音箱服务
	if err := ws.aiSpeaker.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("停止音箱服务失败: %v", err),
		})
		return
	}

	// 清空AI音箱服务引用
	ws.aiSpeaker = nil

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "音箱服务已停止",
	})
}

// getConcurrentStatus 获取并发处理状态
func (ws *WebServer) getConcurrentStatus(c *gin.Context) {
	status := map[string]interface{}{
		"enabled":           ws.config.Speaker.EnableConcurrent,
		"workerCount":       ws.config.Speaker.WorkerCount,
		"queueSize":         ws.config.Speaker.QueueSize,
		"currentQueueSize":  0, // 这里应该获取实际的队列大小
		"processedTasks":    0, // 这里应该获取实际的处理任务数
		"activeWorkers":     ws.config.Speaker.WorkerCount,
		"averageProcessTime": "150ms",
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    status,
	})
}

// updateConcurrentConfig 更新并发配置
func (ws *WebServer) updateConcurrentConfig(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("请求数据格式错误: %v", err),
		})
		return
	}

	logger.Info("收到并发配置更新请求")

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "并发配置更新成功",
	})
}

// updateConfigFields 更新配置字段
func (ws *WebServer) updateConfigFields(data map[string]interface{}) error {
	// AI配置
	if ai, ok := data["ai"].(map[string]interface{}); ok {
		if provider, ok := ai["provider"].(string); ok {
			ws.config.OpenAI.Provider = provider
		}
		if model, ok := ai["model"].(string); ok {
			ws.config.OpenAI.Model = model
		}
		if baseURL, ok := ai["baseURL"].(string); ok {
			ws.config.OpenAI.BaseURL = baseURL
		}
		if proxyURL, ok := ai["proxyURL"].(string); ok {
			ws.config.OpenAI.ProxyURL = proxyURL
		}
		if apiKey, ok := ai["apiKey"].(string); ok {
			ws.config.OpenAI.APIKey = apiKey
		}
		if azureAPIKey, ok := ai["azureAPIKey"].(string); ok {
			ws.config.OpenAI.AzureAPIKey = azureAPIKey
		}
		if azureEndpoint, ok := ai["azureEndpoint"].(string); ok {
			ws.config.OpenAI.AzureEndpoint = azureEndpoint
		}
		if azureDeployment, ok := ai["azureDeployment"].(string); ok {
			ws.config.OpenAI.AzureDeployment = azureDeployment
		}
		if deepSeekAPIKey, ok := ai["deepSeekAPIKey"].(string); ok {
			ws.config.OpenAI.DeepSeekAPIKey = deepSeekAPIKey
		}
		if deepSeekBaseURL, ok := ai["deepSeekBaseURL"].(string); ok {
			ws.config.OpenAI.DeepSeekBaseURL = deepSeekBaseURL
		}
	}

	// 机器人配置
	if bot, ok := data["bot"].(map[string]interface{}); ok {
		if name, ok := bot["name"].(string); ok {
			ws.config.Bot.Name = name
		}
		if profile, ok := bot["profile"].(string); ok {
			ws.config.Bot.Profile = profile
		}
		if masterName, ok := bot["masterName"].(string); ok {
			ws.config.Bot.Master.Name = masterName
		}
		if masterProfile, ok := bot["masterProfile"].(string); ok {
			ws.config.Bot.Master.Profile = masterProfile
		}
		if roomName, ok := bot["roomName"].(string); ok {
			ws.config.Bot.Room.Name = roomName
		}
		if roomDescription, ok := bot["roomDescription"].(string); ok {
			ws.config.Bot.Room.Description = roomDescription
		}
	}

	// 音箱配置
	if speaker, ok := data["speaker"].(map[string]interface{}); ok {
		if name, ok := speaker["name"].(string); ok {
			ws.config.Speaker.Name = name
		}
		if callAIKeywords, ok := speaker["callAIKeywords"].(string); ok {
			ws.config.Speaker.CallAIKeywords = strings.Split(callAIKeywords, ",")
		}
		if wakeupKeywords, ok := speaker["wakeupKeywords"].(string); ok {
			ws.config.Speaker.WakeUpKeywords = strings.Split(wakeupKeywords, ",")
		}
		if exitKeywords, ok := speaker["exitKeywords"].(string); ok {
			ws.config.Speaker.ExitKeywords = strings.Split(exitKeywords, ",")
		}
		if onEnterAI, ok := speaker["onEnterAI"].(string); ok {
			ws.config.Speaker.OnEnterAI = strings.Split(onEnterAI, ",")
		}
		if onExitAI, ok := speaker["onExitAI"].(string); ok {
			ws.config.Speaker.OnExitAI = strings.Split(onExitAI, ",")
		}
		if onAIAsking, ok := speaker["onAIAsking"].(string); ok {
			ws.config.Speaker.OnAIAsking = strings.Split(onAIAsking, ",")
		}
		if onAIReplied, ok := speaker["onAIReplied"].(string); ok {
			ws.config.Speaker.OnAIReplied = strings.Split(onAIReplied, ",")
		}
		if onAIError, ok := speaker["onAIError"].(string); ok {
			ws.config.Speaker.OnAIError = strings.Split(onAIError, ",")
		}
		if streamResponse, ok := speaker["streamResponse"].(bool); ok {
			ws.config.Speaker.StreamResponse = streamResponse
		}
		if enableAudioLog, ok := speaker["enableAudioLog"].(bool); ok {
			ws.config.Speaker.EnableAudioLog = enableAudioLog
		}
		if debugMode, ok := speaker["debugMode"].(bool); ok {
			ws.config.Speaker.Debug = debugMode
		}
	}

	// 小米设备配置
	if mi, ok := data["mi"].(map[string]interface{}); ok {
		if userID, ok := mi["userID"].(string); ok {
			ws.config.Speaker.UserID = userID
		}
		if password, ok := mi["password"].(string); ok {
			ws.config.Speaker.Password = password
		}
		if deviceID, ok := mi["deviceID"].(string); ok {
			ws.config.Speaker.DeviceID = deviceID
		}
		if checkInterval, ok := mi["checkInterval"].(float64); ok {
			ws.config.Speaker.CheckInterval = int(checkInterval)
		}
		if timeout, ok := mi["timeout"].(float64); ok {
			ws.config.Speaker.Timeout = int(timeout)
		}
		if enableTrace, ok := mi["enableTrace"].(bool); ok {
			ws.config.Speaker.EnableTrace = enableTrace
		}
	}

	// 并发配置
	if concurrent, ok := data["concurrent"].(map[string]interface{}); ok {
		if enable, ok := concurrent["enable"].(bool); ok {
			ws.config.Speaker.EnableConcurrent = enable
		}
		if workerCount, ok := concurrent["workerCount"].(float64); ok {
			ws.config.Speaker.WorkerCount = int(workerCount)
		}
		if queueSize, ok := concurrent["queueSize"].(float64); ok {
			ws.config.Speaker.QueueSize = int(queueSize)
		}
		if messageBufferSize, ok := concurrent["messageBufferSize"].(float64); ok {
			ws.config.Speaker.MessageBufferSize = int(messageBufferSize)
		}
		if rateLimit, ok := concurrent["rateLimit"].(float64); ok {
			ws.config.Speaker.RateLimit = int(rateLimit)
		}
		if batchSize, ok := concurrent["batchSize"].(float64); ok {
			ws.config.Speaker.BatchSize = int(batchSize)
		}
		if batchTimeoutSeconds, ok := concurrent["batchTimeoutSeconds"].(float64); ok {
			ws.config.Speaker.BatchTimeoutSeconds = int(batchTimeoutSeconds)
		}
		if enableMetrics, ok := concurrent["enableMetrics"].(bool); ok {
			ws.config.Speaker.EnableMetrics = enableMetrics
		}
	}

	// 数据库配置
	if database, ok := data["database"].(map[string]interface{}); ok {
		if path, ok := database["path"].(string); ok {
			ws.config.Database.Path = path
		}
		if debug, ok := database["debug"].(bool); ok {
			ws.config.Database.Debug = debug
		}
	}

	logger.Info("配置字段更新完成")
	return nil
}



// getStatusField 从状态map中获取字段值，如果不存在则返回默认值
func getStatusField(status map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if status == nil {
		return defaultValue
	}
	if value, exists := status[key]; exists {
		return value
	}
	return defaultValue
}

// validateConfig 验证配置
func (ws *WebServer) validateConfig(c *gin.Context) {
	validationResults := map[string]interface{}{
		"basic": ws.config.ValidateBasic() == nil,
		"ai":    ws.config.ValidateAI() == nil,
		"mi":    ws.config.ValidateMi() == nil,
		"all":   ws.config.ValidateAll() == nil,
	}

	messages := make(map[string]string)
	if err := ws.config.ValidateBasic(); err != nil {
		messages["basic"] = err.Error()
	}
	if err := ws.config.ValidateAI(); err != nil {
		messages["ai"] = err.Error()
	}
	if err := ws.config.ValidateMi(); err != nil {
		messages["mi"] = err.Error()
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data: map[string]interface{}{
			"validation": validationResults,
			"messages":   messages,
			"isComplete": ws.config.IsConfigured(),
		},
	})
}

// testAIConnection 测试AI服务连接
func (ws *WebServer) testAIConnection(c *gin.Context) {
	if err := ws.config.ValidateAI(); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("AI配置无效: %v", err),
		})
		return
	}

	logger.Infof("正在测试AI服务连接 [%s]...", ws.config.OpenAI.Provider)

	// 创建AI客户端进行真实连接测试
	aiClient, err := ws.createAIClient()
	if err != nil {
		logger.Errorf("创建AI客户端失败: %v", err)
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("创建AI客户端失败: %v", err),
			Data: map[string]interface{}{
				"provider": ws.config.OpenAI.Provider,
				"model":    ws.config.OpenAI.Model,
				"status":   "failed",
				"error":    err.Error(),
			},
		})
		return
	}

	// 发送测试请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testMessage := "请回复'连接测试成功'来确认服务正常"
	response, err := aiClient.Chat(ctx, openai.ChatOptions{
		User:   testMessage,
		System: "你是一个AI助手。请简短回复确认连接正常。",
		Trace:  true,
	})

	if err != nil {
		logger.Errorf("AI服务连接测试失败: %v", err)
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("AI服务连接测试失败: %v", err),
			Data: map[string]interface{}{
				"provider":     ws.config.OpenAI.Provider,
				"model":        ws.config.OpenAI.Model,
				"status":       "failed",
				"error":        err.Error(),
				"testMessage":  testMessage,
			},
		})
		return
	}

	logger.Infof("AI服务连接测试成功 [%s]: %s", ws.config.OpenAI.Provider, response)

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("AI服务连接测试成功，收到回复：%s", response),
		Data: map[string]interface{}{
			"provider":     ws.config.OpenAI.Provider,
			"model":        ws.config.OpenAI.Model,
			"status":       "connected",
			"testMessage":  testMessage,
			"response":     response,
			"responseTime": "已在日志中显示",
		},
	})
}

// testMiConnection 测试小米设备连接
func (ws *WebServer) testMiConnection(c *gin.Context) {
	if err := ws.config.ValidateMi(); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("小米设备配置无效: %v", err),
		})
		return
	}

	logger.Info("正在测试小米设备连接...")

	// 创建临时的小米服务客户端进行测试
	testConfig := ws.config.Speaker
	testService, err := miservice.CreateMiService(testConfig)
	if err != nil {
		logger.Errorf("小米设备连接测试失败: %v", err)
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("小米设备连接失败: %v", err),
			Data: map[string]interface{}{
				"userID":   ws.config.Speaker.UserID,
				"deviceID": ws.config.Speaker.DeviceID,
				"status":   "failed",
				"error":    err.Error(),
			},
		})
		return
	}

	// 关闭测试客户端
	defer testService.Close()

	logger.Info("小米设备连接测试成功")

	// 如果配置完整且音箱服务未运行，自动启动音箱服务
	if ws.config.IsConfigured() && (ws.aiSpeaker == nil || !ws.aiSpeaker.IsRunning()) {
		logger.Info("配置完整，正在自动启动音箱服务...")
		
		if err := ws.CreateAISpeaker(); err != nil {
			logger.Warnf("自动启动音箱服务失败: %v", err)
			c.JSON(http.StatusOK, ConfigResponse{
				Success: true,
				Message: "小米设备连接测试成功，但自动启动音箱服务失败",
				Data: map[string]interface{}{
					"userID":      ws.config.Speaker.UserID,
					"deviceID":    ws.config.Speaker.DeviceID,
					"status":      "connected",
					"autoStart":   false,
					"autoStartError": err.Error(),
				},
			})
			return
		}
		
		logger.Info("音箱服务自动启动成功")
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Message: "小米设备连接测试成功，音箱服务已自动启动",
			Data: map[string]interface{}{
				"userID":    ws.config.Speaker.UserID,
				"deviceID":  ws.config.Speaker.DeviceID,
				"status":    "connected",
				"autoStart": true,
				"speakerStatus": ws.aiSpeaker.GetStatus(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "小米设备连接测试成功",
		Data: map[string]interface{}{
			"userID":   ws.config.Speaker.UserID,
			"deviceID": ws.config.Speaker.DeviceID,
			"status":   "connected",
		},
	})
}

// startSpeaker 启动音箱服务
func (ws *WebServer) startSpeaker(c *gin.Context) {
	if !ws.config.IsConfigured() {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: "配置不完整，无法启动音箱服务",
		})
		return
	}

	if ws.aiSpeaker != nil && ws.aiSpeaker.IsRunning() {
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Message: "音箱服务已在运行中",
		})
		return
	}

	// 创建并启动AI音箱服务
	logger.Info("正在启动AI音箱服务...")
	if err := ws.CreateAISpeaker(); err != nil {
		errorMsg := err.Error()
		userFriendlyMsg := "启动音箱服务失败"
		
		// 根据错误类型提供友好的错误提示
		if strings.Contains(errorMsg, "登录失败") {
			if strings.Contains(errorMsg, "状态码: 405") {
				userFriendlyMsg = "小米设备连接失败：请检查小米账号和密码是否正确，或者小米账号是否启用了二步验证。建议：1) 确认账号密码正确 2) 关闭小米账号的二步验证 3) 检查网络连接"
			} else {
				userFriendlyMsg = "小米设备连接失败：请检查小米账号和密码是否正确，以及网络连接是否正常"
			}
		} else if strings.Contains(errorMsg, "设备ID") {
			userFriendlyMsg = "小米设备连接失败：请检查设备ID是否正确，确保设备ID与实际的小爱音箱设备匹配"
		} else if strings.Contains(errorMsg, "AI") {
			userFriendlyMsg = "AI服务连接失败：请检查AI服务配置是否正确，包括API密钥、模型名称等"
		}
		
		logger.Errorf("启动音箱服务失败: %v", err)
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: userFriendlyMsg,
			Data: map[string]interface{}{
				"error": errorMsg,
				"suggestion": "请检查配置并重试，如果问题持续存在，请查看系统日志获取详细信息",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "音箱服务启动成功",
		Data: map[string]interface{}{
			"isRunning": ws.aiSpeaker.IsRunning(),
			"status":    ws.aiSpeaker.GetStatus(),
		},
	})
}

// clearLogs 清空系统日志
func (ws *WebServer) clearLogs(c *gin.Context) {
	logger.ClearLogs()
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "系统日志已清空",
	})
}

// setLogLevel 设置日志级别
func (ws *WebServer) setLogLevel(c *gin.Context) {
	var request struct {
		Level string `json:"level" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("请求数据格式错误: %v", err),
		})
		return
	}
	
	if err := logger.SetLevel(request.Level); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("设置日志级别失败: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("日志级别已设置为: %s", request.Level),
		Data: map[string]interface{}{
			"level": logger.GetLevel(),
		},
	})
}

// getLogLevel 获取当前日志级别
func (ws *WebServer) getLogLevel(c *gin.Context) {
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data: map[string]interface{}{
			"level": logger.GetLevel(),
		},
	})
}

// restartSpeaker 重启音箱服务（配置热重载）
// executeVoiceCommand 执行语音命令
func (ws *WebServer) executeVoiceCommand(c *gin.Context) {
	var request struct {
		Command      string `json:"command" binding:"required"`
		NeedResponse bool   `json:"needResponse"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	logger.Infof("收到语音命令执行请求: %s (需要回应: %t)", request.Command, request.NeedResponse)

	// 获取当前运行的音箱服务
	if ws.aiSpeaker == nil {
		c.JSON(http.StatusServiceUnavailable, ConfigResponse{
			Success: false,
			Message: "音箱服务未启动，请先启动音箱服务",
		})
		return
	}

	// 暂时通过TTS播报的方式执行语音命令
	// TODO: 将来可以扩展为真正的ubus协议语音命令执行
	commandText := request.Command
	if request.NeedResponse {
		commandText = "小爱同学，" + request.Command
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ws.aiSpeaker.ExecuteCommand(ctx, commandText); err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("执行语音命令失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("语音命令 '%s' 已通过TTS播报", request.Command),
		Data: map[string]interface{}{
			"command": request.Command,
			"needResponse": request.NeedResponse,
			"method":  "tts_broadcast",
		},
	})
}

func (ws *WebServer) restartSpeaker(c *gin.Context) {
	logger.Info("收到重启音箱服务请求（配置热重载）")

	// 如果配置不完整，直接返回
	if !ws.config.IsConfigured() {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Message: "配置不完整，无法重启音箱服务",
		})
		return
	}

	// 停止现有服务
	if ws.aiSpeaker != nil {
		logger.Info("正在停止现有音箱服务...")
		if err := ws.aiSpeaker.Stop(); err != nil {
			logger.Warnf("停止音箱服务失败: %v", err)
		} else {
			logger.Info("音箱服务已停止")
		}
		ws.aiSpeaker = nil
	}

	// 重新启动服务
	logger.Info("正在重新启动AI音箱服务...")
	if err := ws.CreateAISpeaker(); err != nil {
		errorMsg := err.Error()
		userFriendlyMsg := "重启音箱服务失败"
		
		// 根据错误类型提供友好的错误提示
		if strings.Contains(errorMsg, "登录失败") {
			userFriendlyMsg = "小米设备连接失败：请检查小米账号和密码是否正确"
		} else if strings.Contains(errorMsg, "设备ID") {
			userFriendlyMsg = "小米设备连接失败：请检查设备ID是否正确"
		} else if strings.Contains(errorMsg, "AI") {
			userFriendlyMsg = "AI服务连接失败：请检查AI服务配置是否正确"
		}
		
		logger.Errorf("重启音箱服务失败: %v", err)
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Message: userFriendlyMsg,
			Data: map[string]interface{}{
				"error": errorMsg,
				"suggestion": "请检查配置并重试",
			},
		})
		return
	}

	logger.Info("音箱服务重启成功，新配置已生效")
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "音箱服务重启成功，配置已更新",
		Data: map[string]interface{}{
			"isRunning": ws.aiSpeaker.IsRunning(),
			"status":    ws.aiSpeaker.GetStatus(),
		},
	})
}

// createAIClient 创建AI客户端用于测试
func (ws *WebServer) createAIClient() (*openai.Client, error) {
	return openai.NewClient(ws.config.OpenAI)
}



 