package miservice

import (
	"mi-gpt-go/internal/config"
	"mi-gpt-go/pkg/logger"
)

// CreateMiService 创建基于第三方库的小米服务客户端
func CreateMiService(cfg config.SpeakerConfig) (MiServiceInterface, error) {
	logger.Info("🎯 使用第三方库小米客户端（xiaoai-tts）")
	
	// 使用增强的错误处理创建客户端
	client, err := NewXiaoAiClient(cfg.UserID, cfg.Password)
	if err != nil {
		logger.Errorf("❌ 小米客户端创建失败: %v", err)
		return nil, err
	}
	
	logger.Info("✅ 小米客户端创建成功（xiaoai-tts）")
	return client, nil
} 