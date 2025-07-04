package openai

import (
	"context"
	"fmt"
	"io"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/pkg/logger"
	"net/http"
	"net/url"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Client AI客户端（支持OpenAI、Azure OpenAI、DeepSeek等）
type Client struct {
	client       *openai.Client  // 底层OpenAI客户端
	model        string          // 使用的模型名称
	enableSearch bool           // 是否启用搜索功能
	provider     string         // 服务提供商类型
}

// ChatOptions 聊天选项
type ChatOptions struct {
	User         string
	System       string
	Model        string
	JSONMode     bool
	RequestID    string
	Trace        bool
	EnableSearch bool
	OnStream     func(string)
}

// NewClient 创建新的AI客户端，支持多种服务提供商
func NewClient(cfg config.OpenAIConfig) (*Client, error) {
	var client *openai.Client

	// 验证基础配置
	if err := validateOpenAIConfig(cfg); err != nil {
		return nil, fmt.Errorf("AI配置验证失败: %v", err)
	}

	// 配置HTTP客户端和代理
	httpClient := &http.Client{}
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("解析代理URL失败: %v", err)
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// 根据服务提供商创建相应的客户端
	switch cfg.Provider {
	case "azure":
		// Azure OpenAI 服务
		if cfg.AzureAPIKey == "" {
			return nil, fmt.Errorf("使用Azure OpenAI时必须提供AzureAPIKey")
		}
		if cfg.AzureEndpoint == "" {
			return nil, fmt.Errorf("使用Azure OpenAI时必须提供AzureEndpoint")
		}
		clientConfig := openai.DefaultAzureConfig(cfg.AzureAPIKey, cfg.AzureEndpoint)
		clientConfig.AzureModelMapperFunc = func(model string) string {
			return cfg.AzureDeployment
		}
		clientConfig.HTTPClient = httpClient
		client = openai.NewClientWithConfig(clientConfig)
		logger.Info("已初始化Azure OpenAI客户端")

	case "deepseek":
		// DeepSeek 服务
		if cfg.DeepSeekAPIKey == "" {
			return nil, fmt.Errorf("使用DeepSeek时必须提供DeepSeekAPIKey")
		}
		baseURL := cfg.DeepSeekBaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
		if !isValidURL(baseURL) {
			return nil, fmt.Errorf("deepSeek BaseURL格式无效: %s", baseURL)
		}
		clientConfig := openai.DefaultConfig(cfg.DeepSeekAPIKey)
		clientConfig.BaseURL = baseURL
		clientConfig.HTTPClient = httpClient
		client = openai.NewClientWithConfig(clientConfig)
		logger.Info("已初始化DeepSeek客户端")

	case "openai":
		fallthrough
	default:
		// OpenAI 官方服务
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("使用OpenAI时必须提供APIKey")
		}
		clientConfig := openai.DefaultConfig(cfg.APIKey)
		if cfg.BaseURL != "" {
			if !isValidURL(cfg.BaseURL) {
				return nil, fmt.Errorf("openAI BaseURL格式无效: %s", cfg.BaseURL)
			}
			clientConfig.BaseURL = cfg.BaseURL
		}
		clientConfig.HTTPClient = httpClient
		client = openai.NewClientWithConfig(clientConfig)
		logger.Info("已初始化OpenAI客户端")
	}

	return &Client{
		client:       client,
		model:        cfg.Model,
		enableSearch: cfg.EnableSearch,
		provider:     cfg.Provider,
	}, nil
}

// Chat 普通聊天（支持OpenAI、DeepSeek等）
func (c *Client) Chat(ctx context.Context, options ChatOptions) (string, error) {
	if options.Trace {
		logger.Infof("🔥 AI对话请求 [%s]\n🤖️ 系统提示: %s\n😊 用户输入: %s", 
			c.provider, getDefault(options.System, "无"), options.User)
	}

	messages := []openai.ChatCompletionMessage{}
	if options.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: options.System,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: options.User,
	})

	req := openai.ChatCompletionRequest{
		Model:    getDefault(options.Model, c.model),
		Messages: messages,
	}

	if options.JSONMode {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.Errorf("LLM 响应异常: %v", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("未收到 AI 响应")
	}

	content := resp.Choices[0].Message.Content
	if options.Trace {
		logger.Infof("✅ AI回复 [%s]: %s", c.provider, getDefault(content, "无回复"))
	}

	return content, nil
}

// ChatStream 流式聊天（支持OpenAI、DeepSeek等）
func (c *Client) ChatStream(ctx context.Context, options ChatOptions) (string, error) {
	if options.Trace {
		logger.Infof("🔥 AI流式对话请求 [%s]\n🤖️ 系统提示: %s\n😊 用户输入: %s", 
			c.provider, getDefault(options.System, "无"), options.User)
	}

	messages := []openai.ChatCompletionMessage{}
	if options.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: options.System,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: options.User,
	})

	req := openai.ChatCompletionRequest{
		Model:    getDefault(options.Model, c.model),
		Messages: messages,
		Stream:   true,
	}

	if options.JSONMode {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.Errorf("LLM 响应异常: %v", err)
		return "", err
	}
	defer stream.Close()

	var content strings.Builder
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("流式响应错误: %v", err)
			return "", err
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				content.WriteString(delta)
				if options.OnStream != nil {
					options.OnStream(delta)
				}
			}
		}
	}

	result := content.String()
	if options.Trace {
		logger.Infof("✅ AI流式回复完成 [%s]: %s", c.provider, getDefault(result, "无回复"))
	}

	return result, nil
}

// getDefault 获取默认值
func getDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// validateOpenAIConfig 验证OpenAI配置
func validateOpenAIConfig(cfg config.OpenAIConfig) error {
	if cfg.Provider == "" {
		return fmt.Errorf("AI服务提供商不能为空")
	}

	switch cfg.Provider {
	case "openai":
		if cfg.APIKey == "" {
			return fmt.Errorf("openAI API Key不能为空")
		}
		if cfg.BaseURL != "" && !isValidURL(cfg.BaseURL) {
			return fmt.Errorf("openAI BaseURL格式无效: %s", cfg.BaseURL)
		}
	case "azure":
		if cfg.AzureAPIKey == "" {
			return fmt.Errorf("azure API Key不能为空")
		}
		if cfg.AzureEndpoint == "" {
			return fmt.Errorf("azure端点不能为空")
		}
		if !isValidURL(cfg.AzureEndpoint) {
			return fmt.Errorf("azure端点URL格式无效: %s", cfg.AzureEndpoint)
		}
	case "deepseek":
		if cfg.DeepSeekAPIKey == "" {
			return fmt.Errorf("deepSeek API Key不能为空")
		}
		baseURL := cfg.DeepSeekBaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
		if !isValidURL(baseURL) {
			return fmt.Errorf("deepSeek BaseURL格式无效: %s", baseURL)
		}
	default:
		return fmt.Errorf("不支持的AI服务提供商: %s", cfg.Provider)
	}

	return nil
}

// isValidURL 验证URL格式是否有效
func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	
	// 检查是否有有效的scheme
	if parsedURL.Scheme == "" {
		return false
	}
	
	// 检查是否有有效的host
	if parsedURL.Host == "" {
		return false
	}
	
	// 只允许http和https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	
	return true
} 