package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// 全局配置服务实例
var dbConfigService interface {
	LoadConfig() (*Config, error)
	SaveConfig(*Config) error
}

// SetDBConfigService 设置数据库配置服务
func SetDBConfigService(service interface {
	LoadConfig() (*Config, error)
	SaveConfig(*Config) error
}) {
	dbConfigService = service
}

// Config 系统配置
type Config struct {
	Database DatabaseConfig `json:"database"`
	Speaker  SpeakerConfig  `json:"speaker"`
	Bot      BotConfig      `json:"bot"`
	OpenAI   OpenAIConfig   `json:"openai"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path  string `json:"path"`
	Debug bool   `json:"debug"`
}

// SpeakerConfig 音箱配置
type SpeakerConfig struct {
	UserID                 string   `json:"userId"`
	Password               string   `json:"password"`
	DeviceID               string   `json:"deviceId"`                // 小爱音箱设备ID
	Name                   string   `json:"name"`
	CallAIKeywords         []string `json:"callAIKeywords"`
	WakeUpKeywords         []string `json:"wakeUpKeywords"`
	ExitKeywords           []string `json:"exitKeywords"`
	SwitchSpeakerKeywords  []string `json:"switchSpeakerKeywords"`
	OnEnterAI              []string `json:"onEnterAI"`
	OnExitAI               []string `json:"onExitAI"`
	OnAIAsking             []string `json:"onAIAsking"`
	OnAIReplied            []string `json:"onAIReplied"`
	OnAIError              []string `json:"onAIError"`
	StreamResponse         bool     `json:"streamResponse"`
	EnableAudioLog         bool     `json:"enableAudioLog"`
	KeepAlive              bool     `json:"keepAlive"`
	AudioActive            string   `json:"audioActive"`
	AudioError             string   `json:"audioError"`
	AudioBeep              string   `json:"audioBeep"`              // 提示音URL
	AudioSilent            string   `json:"audioSilent"`            // 静音URL
	CheckInterval          int      `json:"checkInterval"`          // 检查间隔(毫秒)
	CheckTTSStatusAfter    int      `json:"checkTTSStatusAfter"`    // TTS后检查延迟(秒)
	Timeout                int      `json:"timeout"`                // 超时时间(毫秒)
	EnableTrace            bool     `json:"enableTrace"`            // 启用跟踪日志
	Debug                  bool     `json:"debug"`
	
	// 并发处理配置
	EnableConcurrent       bool     `json:"enableConcurrent"`       // 是否启用并发处理
	WorkerCount            int      `json:"workerCount"`            // 工作协程数量
	QueueSize              int      `json:"queueSize"`              // 任务队列大小
	MessageBufferSize      int      `json:"messageBufferSize"`      // 消息缓冲区大小
	RateLimit              int      `json:"rateLimit"`              // 速率限制（每秒任务数）
	BatchSize              int      `json:"batchSize"`              // 批处理大小
	BatchTimeoutSeconds    int      `json:"batchTimeoutSeconds"`    // 批处理超时（秒）
	EnableMetrics          bool     `json:"enableMetrics"`          // 是否启用指标统计
}

// BotConfig 机器人配置
type BotConfig struct {
	Name            string `json:"name"`
	Profile         string `json:"profile"`
	SystemTemplate  string `json:"systemTemplate"`
	Master          MasterConfig `json:"master"`
	Room            RoomConfig   `json:"room"`
}

// MasterConfig 主人配置
type MasterConfig struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

// RoomConfig 房间配置
type RoomConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OpenAIConfig AI服务配置
type OpenAIConfig struct {
	// 通用配置
	APIKey               string `json:"apiKey"`               // API密钥
	BaseURL              string `json:"baseUrl"`              // API基础URL
	Model                string `json:"model"`                // 模型名称
	ProxyURL             string `json:"proxyUrl"`             // 代理URL
	EnableSearch         bool   `json:"enableSearch"`         // 是否启用搜索功能
	
	// Azure OpenAI 配置
	AzureAPIKey          string `json:"azureApiKey"`          // Azure API密钥
	AzureEndpoint        string `json:"azureEndpoint"`        // Azure端点
	AzureDeployment      string `json:"azureDeployment"`      // Azure部署名称
	
	// DeepSeek 配置
	DeepSeekAPIKey       string `json:"deepSeekApiKey"`       // DeepSeek API密钥
	DeepSeekBaseURL      string `json:"deepSeekBaseUrl"`      // DeepSeek API基础URL
	
	// 服务提供商选择
	Provider             string `json:"provider"`             // 服务提供商：openai, azure, deepseek
}

const ConfigFileName = "config.json"

// LoadWithDefaults 加载配置，优先从数据库加载，然后是文件，最后是默认值
func LoadWithDefaults() (*Config, error) {
	// 优先从数据库加载配置
	if dbConfigService != nil {
		if config, err := dbConfigService.LoadConfig(); err == nil {
			return config, nil
		}
	}
	
	// 然后尝试从配置文件加载
	if config, err := LoadFromFile(); err == nil {
		return config, nil
	}
	
	// 如果都不存在，使用默认配置
	config := createDefaultConfig()
	return config, nil
}

// LoadFromDB 从数据库加载配置
func LoadFromDB() (*Config, error) {
	if dbConfigService == nil {
		return nil, fmt.Errorf("数据库配置服务未初始化")
	}
	return dbConfigService.LoadConfig()
}

// SaveToDB 保存配置到数据库
func (c *Config) SaveToDB() error {
	if dbConfigService == nil {
		return fmt.Errorf("数据库配置服务未初始化")
	}
	return dbConfigService.SaveConfig(c)
}

// createDefaultConfig 创建默认配置
func createDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path:  "app.db",
			Debug: false,
		},
		Speaker: SpeakerConfig{
			UserID:                 "",
			Password:               "",
			DeviceID:               "",
			Name:                   "傻妞",
			CallAIKeywords:         []string{"请", "你", "傻妞"},
			WakeUpKeywords:         []string{"打开", "进入", "召唤"},
			ExitKeywords:           []string{"关闭", "退出", "再见"},
			SwitchSpeakerKeywords:  []string{"音色切换到"},
			OnEnterAI:              []string{"你好，我是傻妞，很高兴认识你"},
			OnExitAI:               []string{"傻妞已退出"},
			OnAIAsking:             []string{"让我先想想", "请稍等"},
			OnAIReplied:            []string{"我说完了", "还有其他问题吗"},
			OnAIError:              []string{"啊哦，出错了，请稍后再试吧！"},
			StreamResponse:         true,
			EnableAudioLog:         false,
			KeepAlive:              false,
			AudioActive:            "",
			AudioError:             "",
			AudioBeep:              "",
			AudioSilent:            "",
			CheckInterval:          1000,
			CheckTTSStatusAfter:    3,
			Timeout:                5000,
			EnableTrace:            false,
			Debug:                  false,
			
			// 并发处理配置
			EnableConcurrent:       true,
			WorkerCount:            4,
			QueueSize:              100,
			MessageBufferSize:      50,
			RateLimit:              10,
			BatchSize:              5,
			BatchTimeoutSeconds:    5,
			EnableMetrics:          true,
		},
		Bot: BotConfig{
			Name:    "傻妞",
			Profile: "你是一个聪明、可爱、有趣的AI助手，拥有丰富的知识和幽默感",
			Master: MasterConfig{
				Name:    "主人",
				Profile: "我的主人，一个有趣的人",
			},
			Room: RoomConfig{
				Name:        "客厅",
				Description: "这是一个温馨的客厅，我们经常在这里聊天",
			},
		},
		OpenAI: OpenAIConfig{
			// 通用配置
			APIKey:          "",
			BaseURL:         "",
			Model:           "deepseek-chat",  // 更改默认模型
			ProxyURL:        "",
			EnableSearch:    false,
			
			// Azure OpenAI 配置
			AzureAPIKey:     "",
			AzureEndpoint:   "",
			AzureDeployment: "",
			
			// DeepSeek 配置
			DeepSeekAPIKey:  "",
			DeepSeekBaseURL: "https://api.deepseek.com/v1",
			
			// 服务提供商选择
			Provider:        "deepseek", // 默认使用DeepSeek（需要配置API Key）
		},
	}
}

// LoadFromFile 从配置文件加载配置
func LoadFromFile() (*Config, error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	return &config, nil
}

// SaveToFile 保存配置到文件
func (c *Config) SaveToFile() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	if err := os.WriteFile(ConfigFileName, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}
	
	return nil
}

// Load 加载配置（兼容旧版本，支持环境变量覆盖）
func Load() (*Config, error) {
	// 首先加载默认配置或文件配置
	config, err := LoadWithDefaults()
	if err != nil {
		return nil, err
	}
	
	// 如果存在环境变量，则覆盖配置（向后兼容）
	if value := os.Getenv("MI_USER_ID"); value != "" {
		config.Speaker.UserID = value
	}
	if value := os.Getenv("MI_PASSWORD"); value != "" {
		config.Speaker.Password = value
	}
	if value := os.Getenv("MI_DEVICE_ID"); value != "" {
		config.Speaker.DeviceID = value
	}
	if value := os.Getenv("OPENAI_API_KEY"); value != "" {
		config.OpenAI.APIKey = value
	}
	if value := os.Getenv("DEEPSEEK_API_KEY"); value != "" {
		config.OpenAI.DeepSeekAPIKey = value
	}
	if value := os.Getenv("AI_PROVIDER"); value != "" {
		config.OpenAI.Provider = value
	}
	
	return config, nil
}



// ValidateBasic 验证基础配置（程序启动必需的）
func (c *Config) ValidateBasic() error {
	// 只验证数据库配置
	if c.Database.Path == "" {
		return fmt.Errorf("数据库路径不能为空")
	}
	return nil
}

// ValidateAI 验证AI服务配置
func (c *Config) ValidateAI() error {
	if c.OpenAI.Provider == "" {
		return fmt.Errorf("AI服务提供商不能为空")
	}
	
	switch c.OpenAI.Provider {
	case "openai":
		if c.OpenAI.APIKey == "" {
			return fmt.Errorf("OpenAI API Key不能为空")
		}
		if c.OpenAI.Model == "" {
			return fmt.Errorf("OpenAI模型不能为空")
		}
	case "azure":
		if c.OpenAI.AzureAPIKey == "" {
			return fmt.Errorf("azure API Key不能为空")
		}
		if c.OpenAI.AzureEndpoint == "" {
			return fmt.Errorf("Azure端点不能为空")
		}
		if c.OpenAI.AzureDeployment == "" {
			return fmt.Errorf("Azure部署名称不能为空")
		}
	case "deepseek":
		if c.OpenAI.DeepSeekAPIKey == "" {
			return fmt.Errorf("DeepSeek API Key不能为空")
		}
	default:
		return fmt.Errorf("不支持的AI服务提供商: %s", c.OpenAI.Provider)
	}
	
	return nil
}

// ValidateMi 验证小米设备配置
func (c *Config) ValidateMi() error {
	if c.Speaker.UserID == "" {
		return fmt.Errorf("小米用户ID不能为空")
	}
	if c.Speaker.Password == "" {
		return fmt.Errorf("小米密码不能为空")
	}
	if c.Speaker.DeviceID == "" {
		return fmt.Errorf("小爱音箱设备ID不能为空")
	}
	return nil
}

// ValidateAll 验证所有配置
func (c *Config) ValidateAll() error {
	if err := c.ValidateBasic(); err != nil {
		return err
	}
	if err := c.ValidateAI(); err != nil {
		return err
	}
	if err := c.ValidateMi(); err != nil {
		return err
	}
	return nil
}

// IsConfigured 检查是否已完整配置
func (c *Config) IsConfigured() bool {
	return c.ValidateAll() == nil
} 