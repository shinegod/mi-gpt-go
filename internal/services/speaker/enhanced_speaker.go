package speaker

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/internal/services/miservice"
	"mi-gpt-go/internal/services/openai"
	"mi-gpt-go/pkg/logger"
	"strings"
	"sync"
	"time"
)

// EnhancedAISpeaker 增强版AI音箱服务
type EnhancedAISpeaker struct {
	config        *config.Config
	xiaomiService miservice.MiServiceInterface
	openaiService *openai.Client
	mutex         sync.RWMutex
	isRunning     bool
	stopChannel   chan struct{}
	isHealthy     bool
	lastActivity  time.Time
}

// NewEnhancedAISpeaker 创建增强版AI音箱服务
func NewEnhancedAISpeaker(cfg *config.Config) (*EnhancedAISpeaker, error) {
	logger.Info("正在创建增强版AI音箱服务...")
	
	// 使用工厂函数创建小爱音箱客户端
	xiaomiService, err := miservice.CreateMiService(cfg.Speaker)
	if err != nil {
		return nil, fmt.Errorf("创建小爱音箱客户端失败: %v", err)
	}

	// 创建OpenAI客户端
	openaiClient, err := openai.NewClient(cfg.OpenAI)
	if err != nil {
		return nil, fmt.Errorf("创建OpenAI客户端失败: %v", err)
	}

	enhanced := &EnhancedAISpeaker{
		config:        cfg,
		xiaomiService: xiaomiService,
		openaiService: openaiClient,
		stopChannel:   make(chan struct{}),
		isHealthy:     true,
		lastActivity:  time.Now(),
	}

	logger.Info("增强版AI音箱服务初始化成功")
	return enhanced, nil
}

// Start 启动服务
func (eas *EnhancedAISpeaker) Start() error {
	eas.mutex.Lock()
	defer eas.mutex.Unlock()

	if eas.isRunning {
		return fmt.Errorf("服务已在运行中")
	}

	logger.Info("正在启动增强版AI音箱服务...")

	// 启动消息处理循环
	eas.startMessageProcessor()

	eas.isRunning = true
	logger.Info("🎉 增强版AI音箱服务启动成功（演示模式）")
	return nil
}

// Stop 停止服务
func (eas *EnhancedAISpeaker) Stop() error {
	eas.mutex.Lock()
	defer eas.mutex.Unlock()

	if !eas.isRunning {
		return nil
	}

	logger.Info("正在停止增强版AI音箱服务...")
	close(eas.stopChannel)

	if err := eas.xiaomiService.Close(); err != nil {
		logger.Warnf("关闭小米服务失败: %v", err)
	}

	eas.isRunning = false
	logger.Info("增强版AI音箱服务已停止")
	return nil
}

// IsRunning 检查服务是否运行中
func (eas *EnhancedAISpeaker) IsRunning() bool {
	eas.mutex.RLock()
	defer eas.mutex.RUnlock()
	return eas.isRunning
}

// startMessageProcessor 启动消息处理器
func (eas *EnhancedAISpeaker) startMessageProcessor() {
	go func() {
		logger.Info("消息处理器已启动")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 启动对话轮询
		go func() {
			err := eas.xiaomiService.PollConversations(ctx, eas.config.Speaker.DeviceID, func(record *miservice.ConversationRecord) {
				logger.Infof("🎯 收到用户提问: %s", record.Query)
				eas.handleMessage(record.Query)
			})
			if err != nil && err != context.Canceled {
				// 如果是不支持轮询的错误，只记录一次日志，不重试
				if strings.Contains(err.Error(), "不支持对话记录轮询") {
					logger.Infof("ℹ️ %s", err.Error())
				} else {
					logger.Errorf("对话轮询失败: %v", err)
				}
			}
		}()

		// 等待停止信号
		<-eas.stopChannel
		cancel()
		logger.Info("消息处理器已停止")
	}()
}

// handleMessage 处理消息
func (eas *EnhancedAISpeaker) handleMessage(text string) {
	eas.lastActivity = time.Now()

	// 检查AI服务是否正确配置
	if eas.openaiService == nil {
		logger.Warn("AI服务未配置，跳过消息处理")
		response := "AI服务未配置，请在Web管理面板中配置AI服务。"
		err := eas.xiaomiService.Say(response)
		if err != nil {
			logger.Errorf("发送配置提示失败: %v", err)
		}
		return
	}

	// 调用OpenAI获取回复
	ctx := context.Background()
	response, err := eas.openaiService.Chat(ctx, openai.ChatOptions{
		User: text,
	})
	if err != nil {
		logger.Errorf("获取AI回复失败: %v", err)
		
		// 根据错误类型提供不同的提示
		var userResponse string
		errorStr := err.Error()
		if strings.Contains(errorStr, "unsupported protocol scheme") {
			userResponse = "AI服务配置不完整，请在Web管理面板中配置API密钥和服务地址。"
		} else if strings.Contains(errorStr, "401") || strings.Contains(errorStr, "403") {
			userResponse = "AI服务认证失败，请检查API密钥是否正确。"
		} else if strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "connection") {
			userResponse = "AI服务连接超时，请检查网络连接或代理设置。"
		} else {
			userResponse = "抱歉，我现在无法回答您的问题。请稍后再试。"
		}
		
		// 发送错误提示
		err = eas.xiaomiService.Say(userResponse)
		if err != nil {
			logger.Errorf("发送错误提示失败: %v", err)
		}
		return
	}

	// 发送回复到小爱音箱
	err = eas.xiaomiService.Say(response)
	if err != nil {
		logger.Errorf("发送回复失败: %v", err)
	}
}

// GetStatus 获取音箱状态
func (eas *EnhancedAISpeaker) GetStatus() map[string]interface{} {
	eas.mutex.RLock()
	defer eas.mutex.RUnlock()
	
	status := map[string]interface{}{
		"isRunning":     eas.isRunning,
		"isHealthy":     eas.isHealthy,
		"lastActivity":  eas.lastActivity.Format("2006-01-02 15:04:05"),
		"service":       "xiaoai-tts",
		"deviceID":      eas.config.Speaker.DeviceID,
		"deviceName":    eas.config.Speaker.Name,
	}
	
	// 获取小米服务状态
	if eas.xiaomiService != nil {
		xiaomiStatus := eas.xiaomiService.GetHealthStatus()
		status["xiaomiService"] = xiaomiStatus
		status["xiaomiHealthy"] = eas.xiaomiService.IsHealthy()
	}
	
	// 获取AI服务状态
	if eas.openaiService != nil {
		status["aiService"] = "configured"
		status["aiProvider"] = eas.config.OpenAI.Provider
		status["aiModel"] = eas.config.OpenAI.Model
	} else {
		status["aiService"] = "not_configured"
	}
	
	return status
}

// ExecuteCommand 执行命令
func (eas *EnhancedAISpeaker) ExecuteCommand(ctx context.Context, command string) error {
	if !eas.IsRunning() {
		return fmt.Errorf("音箱服务未运行")
	}
	
	logger.Infof("🎯 执行命令: %s", command)
	
	// 检查AI服务
	if eas.openaiService == nil {
		logger.Warn("AI服务未配置，直接播放TTS")
		return eas.xiaomiService.Say(command)
	}
	
	// 通过消息处理器处理命令
	eas.handleMessage(command)
	return nil
}

// Restart 重启服务
func (eas *EnhancedAISpeaker) Restart() error {
	logger.Info("重启AI音箱服务...")
	
	// 停止服务
	if err := eas.Stop(); err != nil {
		logger.Warnf("停止服务失败: %v", err)
	}
	
	// 重新创建stopChannel
	eas.stopChannel = make(chan struct{})
	
	// 重启服务
	return eas.Start()
}

// GetHealthStatus 获取健康状态
func (eas *EnhancedAISpeaker) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"healthy":       eas.isHealthy,
		"running":       eas.isRunning,
		"lastActivity":  eas.lastActivity,
		"xiaomiHealthy": eas.xiaomiService != nil && eas.xiaomiService.IsHealthy(),
		"aiConfigured":  eas.openaiService != nil,
	}
}

// SetHealthy 设置健康状态
func (eas *EnhancedAISpeaker) SetHealthy(healthy bool) {
	eas.mutex.Lock()
	defer eas.mutex.Unlock()
	eas.isHealthy = healthy
} 