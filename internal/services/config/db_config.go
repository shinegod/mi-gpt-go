package config

import (
	"encoding/json"
	"fmt"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/internal/database"
	"mi-gpt-go/internal/models"
	"mi-gpt-go/pkg/logger"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DBConfigService 数据库配置服务
type DBConfigService struct {
	defaultConfig *config.Config
}

// NewDBConfigService 创建新的数据库配置服务
func NewDBConfigService() *DBConfigService {
	// 获取默认配置作为后备
	defaultConfig, _ := config.LoadWithDefaults()
	return &DBConfigService{
		defaultConfig: defaultConfig,
	}
}

// LoadConfig 从数据库加载完整配置
func (s *DBConfigService) LoadConfig() (*config.Config, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 从数据库读取所有配置
	var configItems []models.Config
	if err := db.Find(&configItems).Error; err != nil {
		logger.Warn("读取数据库配置失败，使用默认配置:", err)
		return s.initializeDefaultConfig()
	}

	// 如果数据库中没有配置，初始化默认配置
	if len(configItems) == 0 {
		logger.Info("数据库中没有配置，初始化默认配置")
		return s.initializeDefaultConfig()
	}

	// 将数据库配置转换为Config结构
	cfg, err := s.convertToConfig(configItems)
	if err != nil {
		logger.Error("转换数据库配置失败:", err)
		return s.defaultConfig, err
	}

	logger.Info("成功从数据库加载配置")
	return cfg, nil
}

// SaveConfig 保存完整配置到数据库
func (s *DBConfigService) SaveConfig(cfg *config.Config) error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 将Config结构转换为数据库配置项
	configItems, err := s.convertFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("转换配置失败: %v", err)
	}

	// 使用事务保存配置
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range configItems {
			// 使用Clauses实现真正的UPSERT操作
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "type", "group", "updated_at"}),
			}).Create(&item)
			
			if result.Error != nil {
				return fmt.Errorf("保存配置项 %s 失败: %v", item.Key, result.Error)
			}
		}
		return nil
	})
}

// GetConfigValue 获取单个配置值
func (s *DBConfigService) GetConfigValue(key string) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", fmt.Errorf("数据库未初始化")
	}

	var configItem models.Config
	if err := db.Where("key = ?", key).First(&configItem).Error; err != nil {
		return "", fmt.Errorf("配置项 %s 不存在: %v", key, err)
	}

	return configItem.Value, nil
}

// SetConfigValue 设置单个配置值
func (s *DBConfigService) SetConfigValue(key, value, valueType, group string) error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	configItem := models.Config{
		Key:   key,
		Value: value,
		Type:  valueType,
		Group: group,
	}

	return db.Save(&configItem).Error
}

// initializeDefaultConfig 初始化默认配置到数据库
func (s *DBConfigService) initializeDefaultConfig() (*config.Config, error) {
	logger.Info("初始化默认配置到数据库")
	
	if err := s.SaveConfig(s.defaultConfig); err != nil {
		logger.Error("初始化默认配置失败:", err)
		return s.defaultConfig, err
	}

	return s.defaultConfig, nil
}

// convertToConfig 将数据库配置项转换为Config结构
func (s *DBConfigService) convertToConfig(items []models.Config) (*config.Config, error) {
	// 创建默认配置的副本，避免修改原始默认配置
	cfg := *s.defaultConfig

	for _, item := range items {
		if err := s.setConfigField(&cfg, item.Key, item.Value); err != nil {
			logger.Warnf("设置配置字段 %s 失败: %v", item.Key, err)
			continue
		}
	}

	return &cfg, nil
}

// convertFromConfig 将Config结构转换为数据库配置项
func (s *DBConfigService) convertFromConfig(cfg *config.Config) ([]models.Config, error) {
	var items []models.Config

	// AI配置
	items = append(items, s.createConfigItems("ai", map[string]interface{}{
		"ai.provider":             cfg.OpenAI.Provider,
		"ai.apiKey":              cfg.OpenAI.APIKey,
		"ai.baseURL":             cfg.OpenAI.BaseURL,
		"ai.model":               cfg.OpenAI.Model,
		"ai.proxyURL":            cfg.OpenAI.ProxyURL,
		"ai.enableSearch":        cfg.OpenAI.EnableSearch,
		"ai.azureAPIKey":         cfg.OpenAI.AzureAPIKey,
		"ai.azureEndpoint":       cfg.OpenAI.AzureEndpoint,
		"ai.azureDeployment":     cfg.OpenAI.AzureDeployment,
		"ai.deepSeekAPIKey":      cfg.OpenAI.DeepSeekAPIKey,
		"ai.deepSeekBaseURL":     cfg.OpenAI.DeepSeekBaseURL,
	})...)

	// 音箱配置
	items = append(items, s.createConfigItems("speaker", map[string]interface{}{
		"speaker.userID":                cfg.Speaker.UserID,
		"speaker.password":              cfg.Speaker.Password,
		"speaker.deviceID":              cfg.Speaker.DeviceID,
		"speaker.name":                  cfg.Speaker.Name,
		"speaker.callAIKeywords":        cfg.Speaker.CallAIKeywords,
		"speaker.wakeUpKeywords":        cfg.Speaker.WakeUpKeywords,
		"speaker.exitKeywords":          cfg.Speaker.ExitKeywords,
		"speaker.switchSpeakerKeywords": cfg.Speaker.SwitchSpeakerKeywords,
		"speaker.onEnterAI":             cfg.Speaker.OnEnterAI,
		"speaker.onExitAI":              cfg.Speaker.OnExitAI,
		"speaker.onAIAsking":            cfg.Speaker.OnAIAsking,
		"speaker.onAIReplied":           cfg.Speaker.OnAIReplied,
		"speaker.onAIError":             cfg.Speaker.OnAIError,
		"speaker.streamResponse":        cfg.Speaker.StreamResponse,
		"speaker.enableAudioLog":        cfg.Speaker.EnableAudioLog,
		"speaker.keepAlive":             cfg.Speaker.KeepAlive,
		"speaker.audioActive":           cfg.Speaker.AudioActive,
		"speaker.audioError":            cfg.Speaker.AudioError,
		"speaker.audioBeep":             cfg.Speaker.AudioBeep,
		"speaker.audioSilent":           cfg.Speaker.AudioSilent,
		"speaker.checkInterval":         cfg.Speaker.CheckInterval,
		"speaker.checkTTSStatusAfter":   cfg.Speaker.CheckTTSStatusAfter,
		"speaker.timeout":               cfg.Speaker.Timeout,
		"speaker.enableTrace":           cfg.Speaker.EnableTrace,
		"speaker.debug":                 cfg.Speaker.Debug,
		"speaker.enableConcurrent":      cfg.Speaker.EnableConcurrent,
		"speaker.workerCount":           cfg.Speaker.WorkerCount,
		"speaker.queueSize":             cfg.Speaker.QueueSize,
		"speaker.messageBufferSize":     cfg.Speaker.MessageBufferSize,
		"speaker.rateLimit":             cfg.Speaker.RateLimit,
		"speaker.batchSize":             cfg.Speaker.BatchSize,
		"speaker.batchTimeoutSeconds":   cfg.Speaker.BatchTimeoutSeconds,
		"speaker.enableMetrics":         cfg.Speaker.EnableMetrics,
	})...)

	// 机器人配置
	items = append(items, s.createConfigItems("bot", map[string]interface{}{
		"bot.name":               cfg.Bot.Name,
		"bot.profile":            cfg.Bot.Profile,
		"bot.systemTemplate":     cfg.Bot.SystemTemplate,
		"bot.master.name":        cfg.Bot.Master.Name,
		"bot.master.profile":     cfg.Bot.Master.Profile,
		"bot.room.name":          cfg.Bot.Room.Name,
		"bot.room.description":   cfg.Bot.Room.Description,
	})...)

	// 数据库配置
	items = append(items, s.createConfigItems("database", map[string]interface{}{
		"database.path":  cfg.Database.Path,
		"database.debug": cfg.Database.Debug,
	})...)

	return items, nil
}

// createConfigItems 创建配置项列表
func (s *DBConfigService) createConfigItems(group string, values map[string]interface{}) []models.Config {
	var items []models.Config
	
	for key, value := range values {
		valueStr, valueType := s.convertValue(value)
		items = append(items, models.Config{
			Key:   key,
			Value: valueStr,
			Type:  valueType,
			Group: group,
		})
	}
	
	return items
}

// convertValue 转换值为字符串和类型
func (s *DBConfigService) convertValue(value interface{}) (string, string) {
	switch v := value.(type) {
	case string:
		return v, "string"
	case bool:
		return strconv.FormatBool(v), "bool"
	case int:
		return strconv.Itoa(v), "int"
	case []string:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes), "array"
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes), "object"
	}
}

// setConfigField 设置配置字段
func (s *DBConfigService) setConfigField(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("无效的配置键: %s", key)
	}

	switch parts[0] {
	case "ai":
		return s.setAIField(cfg, parts[1:], value)
	case "speaker":
		return s.setSpeakerField(cfg, parts[1:], value)
	case "bot":
		return s.setBotField(cfg, parts[1:], value)
	case "database":
		return s.setDatabaseField(cfg, parts[1:], value)
	default:
		return fmt.Errorf("未知的配置分组: %s", parts[0])
	}
}

// setAIField 设置AI配置字段
func (s *DBConfigService) setAIField(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("AI配置字段名为空")
	}

	switch parts[0] {
	case "provider":
		cfg.OpenAI.Provider = value
	case "apiKey":
		cfg.OpenAI.APIKey = value
	case "baseURL":
		cfg.OpenAI.BaseURL = value
	case "model":
		cfg.OpenAI.Model = value
	case "proxyURL":
		cfg.OpenAI.ProxyURL = value
	case "enableSearch":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.OpenAI.EnableSearch = b
		}
	case "azureAPIKey":
		cfg.OpenAI.AzureAPIKey = value
	case "azureEndpoint":
		cfg.OpenAI.AzureEndpoint = value
	case "azureDeployment":
		cfg.OpenAI.AzureDeployment = value
	case "deepSeekAPIKey":
		cfg.OpenAI.DeepSeekAPIKey = value
	case "deepSeekBaseURL":
		cfg.OpenAI.DeepSeekBaseURL = value
	default:
		return fmt.Errorf("未知的AI配置字段: %s", parts[0])
	}
	return nil
}

// setSpeakerField 设置音箱配置字段
func (s *DBConfigService) setSpeakerField(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("音箱配置字段名为空")
	}

	switch parts[0] {
	case "userID":
		cfg.Speaker.UserID = value
	case "password":
		cfg.Speaker.Password = value
	case "deviceID":
		cfg.Speaker.DeviceID = value
	case "name":
		cfg.Speaker.Name = value
	case "callAIKeywords":
		var keywords []string
		if err := json.Unmarshal([]byte(value), &keywords); err == nil {
			cfg.Speaker.CallAIKeywords = keywords
		}
	case "wakeUpKeywords":
		var keywords []string
		if err := json.Unmarshal([]byte(value), &keywords); err == nil {
			cfg.Speaker.WakeUpKeywords = keywords
		}
	case "exitKeywords":
		var keywords []string
		if err := json.Unmarshal([]byte(value), &keywords); err == nil {
			cfg.Speaker.ExitKeywords = keywords
		}
	case "switchSpeakerKeywords":
		var keywords []string
		if err := json.Unmarshal([]byte(value), &keywords); err == nil {
			cfg.Speaker.SwitchSpeakerKeywords = keywords
		}
	case "onEnterAI":
		var messages []string
		if err := json.Unmarshal([]byte(value), &messages); err == nil {
			cfg.Speaker.OnEnterAI = messages
		}
	case "onExitAI":
		var messages []string
		if err := json.Unmarshal([]byte(value), &messages); err == nil {
			cfg.Speaker.OnExitAI = messages
		}
	case "onAIAsking":
		var messages []string
		if err := json.Unmarshal([]byte(value), &messages); err == nil {
			cfg.Speaker.OnAIAsking = messages
		}
	case "onAIReplied":
		var messages []string
		if err := json.Unmarshal([]byte(value), &messages); err == nil {
			cfg.Speaker.OnAIReplied = messages
		}
	case "onAIError":
		var messages []string
		if err := json.Unmarshal([]byte(value), &messages); err == nil {
			cfg.Speaker.OnAIError = messages
		}
	case "streamResponse":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.StreamResponse = b
		}
	case "enableAudioLog":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.EnableAudioLog = b
		}
	case "keepAlive":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.KeepAlive = b
		}
	case "audioActive":
		cfg.Speaker.AudioActive = value
	case "audioError":
		cfg.Speaker.AudioError = value
	case "audioBeep":
		cfg.Speaker.AudioBeep = value
	case "audioSilent":
		cfg.Speaker.AudioSilent = value
	case "checkInterval":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.CheckInterval = i
		}
	case "checkTTSStatusAfter":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.CheckTTSStatusAfter = i
		}
	case "timeout":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.Timeout = i
		}
	case "enableTrace":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.EnableTrace = b
		}
	case "debug":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.Debug = b
		}
	case "enableConcurrent":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.EnableConcurrent = b
		}
	case "workerCount":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.WorkerCount = i
		}
	case "queueSize":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.QueueSize = i
		}
	case "messageBufferSize":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.MessageBufferSize = i
		}
	case "rateLimit":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.RateLimit = i
		}
	case "batchSize":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.BatchSize = i
		}
	case "batchTimeoutSeconds":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.Speaker.BatchTimeoutSeconds = i
		}
	case "enableMetrics":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Speaker.EnableMetrics = b
		}
	default:
		return fmt.Errorf("未知的音箱配置字段: %s", parts[0])
	}
	return nil
}

// setBotField 设置机器人配置字段
func (s *DBConfigService) setBotField(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("机器人配置字段名为空")
	}

	switch parts[0] {
	case "name":
		cfg.Bot.Name = value
	case "profile":
		cfg.Bot.Profile = value
	case "systemTemplate":
		cfg.Bot.SystemTemplate = value
	case "master":
		if len(parts) > 1 {
			switch parts[1] {
			case "name":
				cfg.Bot.Master.Name = value
			case "profile":
				cfg.Bot.Master.Profile = value
			}
		}
	case "room":
		if len(parts) > 1 {
			switch parts[1] {
			case "name":
				cfg.Bot.Room.Name = value
			case "description":
				cfg.Bot.Room.Description = value
			}
		}
	default:
		return fmt.Errorf("未知的机器人配置字段: %s", parts[0])
	}
	return nil
}

// setDatabaseField 设置数据库配置字段
func (s *DBConfigService) setDatabaseField(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("数据库配置字段名为空")
	}

	switch parts[0] {
	case "path":
		cfg.Database.Path = value
	case "debug":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.Database.Debug = b
		}
	default:
		return fmt.Errorf("未知的数据库配置字段: %s", parts[0])
	}
	return nil
} 