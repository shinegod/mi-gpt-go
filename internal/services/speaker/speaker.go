package speaker

import (
	"context"
	"math/rand"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/pkg/logger"
	"strings"
	"sync"
	"time"
)

// QueryMessage 查询消息
type QueryMessage struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// SpeakerAnswer 音箱回答
type SpeakerAnswer struct {
	Text      string `json:"text,omitempty"`
	KeepAlive bool   `json:"keepAlive,omitempty"`
	PlaySFX   bool   `json:"playSfx,omitempty"`
}

// Command 命令接口
type Command interface {
	Match(msg QueryMessage) bool
	Run(ctx context.Context, msg QueryMessage) error
}

// SimpleCommand 简单命令实现
type SimpleCommand struct {
	MatchFunc func(QueryMessage) bool
	RunFunc   func(context.Context, QueryMessage) error
}

func (c *SimpleCommand) Match(msg QueryMessage) bool {
	return c.MatchFunc(msg)
}

func (c *SimpleCommand) Run(ctx context.Context, msg QueryMessage) error {
	return c.RunFunc(ctx, msg)
}

// Speaker 基础音箱
type Speaker struct {
	mu              sync.RWMutex
	keepAlive       bool
	streamResponse  bool
	enableAudioLog  bool
	debug           bool
	commands        []Command
	cancelFunc      context.CancelFunc
}

// NewSpeaker 创建新的音箱
func NewSpeaker() *Speaker {
	return &Speaker{
		commands:       make([]Command, 0),
		streamResponse: true,
		enableAudioLog: false,
		debug:          false,
	}
}

// AddCommand 添加命令
func (s *Speaker) AddCommand(cmd Command) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append(s.commands, cmd)
}

// IsKeepAlive 是否保持唤醒状态
func (s *Speaker) IsKeepAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keepAlive
}

// SetKeepAlive 设置保持唤醒状态
func (s *Speaker) SetKeepAlive(keepAlive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keepAlive = keepAlive
}

// EnterKeepAlive 进入保持唤醒状态
func (s *Speaker) EnterKeepAlive() error {
	s.SetKeepAlive(true)
	logger.Info("已进入连续对话模式")
	return nil
}

// ExitKeepAlive 退出保持唤醒状态
func (s *Speaker) ExitKeepAlive() error {
	s.SetKeepAlive(false)
	logger.Info("已退出连续对话模式")
	return nil
}

// Response 响应消息
func (s *Speaker) Response(answer SpeakerAnswer) error {
	if answer.Text == "" {
		return nil
	}

	logger.Infof("🔊 音箱回复: %s", answer.Text)
	
	// 这里应该调用实际的小爱音箱 API 来播放语音
	// 目前只是打印日志
	
	return nil
}

// ProcessMessage 处理消息
func (s *Speaker) ProcessMessage(ctx context.Context, msg QueryMessage) error {
	s.mu.RLock()
	commands := make([]Command, len(s.commands))
	copy(commands, s.commands)
	s.mu.RUnlock()

	// 检查命令
	for _, cmd := range commands {
		if cmd.Match(msg) {
			if s.debug {
				logger.Debugf("匹配到命令，执行中...")
			}
			return cmd.Run(ctx, msg)
		}
	}

	return nil
}

// Start 启动音箱服务
func (s *Speaker) Start(ctx context.Context) error {
	logger.Info("音箱服务启动中...")
	
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// 这里应该启动小爱音箱的连接和消息监听
	// 目前只是一个示例循环
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("音箱服务停止")
				return
			case <-ticker.C:
				if s.debug {
					logger.Debug("音箱服务心跳")
				}
			}
		}
	}()

	logger.Info("音箱服务启动成功")
	return nil
}

// Stop 停止音箱服务
func (s *Speaker) Stop() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	logger.Info("音箱服务已停止")
	return nil
}

// AISpeaker AI 音箱
type AISpeaker struct {
	*Speaker
	config          config.SpeakerConfig
	name            string
	callAIKeywords  []string
	wakeUpKeywords  []string
	exitKeywords    []string
	onEnterAI       []string
	onExitAI        []string
	onAIAsking      []string
	onAIReplied     []string
	onAIError       []string
	askAI           func(context.Context, QueryMessage) (SpeakerAnswer, error)
}

// NewAISpeaker 创建新的 AI 音箱
func NewAISpeaker(cfg config.SpeakerConfig) *AISpeaker {
	speaker := &AISpeaker{
		Speaker:         NewSpeaker(),
		config:          cfg,
		name:            cfg.Name,
		callAIKeywords:  cfg.CallAIKeywords,
		wakeUpKeywords:  cfg.WakeUpKeywords,
		exitKeywords:    cfg.ExitKeywords,
		onEnterAI:       cfg.OnEnterAI,
		onExitAI:        cfg.OnExitAI,
		onAIAsking:      cfg.OnAIAsking,
		onAIReplied:     cfg.OnAIReplied,
		onAIError:       cfg.OnAIError,
	}

	speaker.streamResponse = cfg.StreamResponse
	speaker.enableAudioLog = cfg.EnableAudioLog
	speaker.debug = cfg.Debug

	// 添加内置命令
	speaker.addBuiltinCommands()

	return speaker
}

// addBuiltinCommands 添加内置命令
func (ai *AISpeaker) addBuiltinCommands() {
	// 唤醒命令
	ai.AddCommand(&SimpleCommand{
		MatchFunc: func(msg QueryMessage) bool {
			if ai.IsKeepAlive() {
				return false
			}
			for _, keyword := range ai.wakeUpKeywords {
				if strings.Contains(msg.Text, keyword) {
					return true
				}
			}
			return false
		},
		RunFunc: func(ctx context.Context, msg QueryMessage) error {
			return ai.enterAI(ctx)
		},
	})

	// 退出命令
	ai.AddCommand(&SimpleCommand{
		MatchFunc: func(msg QueryMessage) bool {
			if !ai.IsKeepAlive() {
				return false
			}
			for _, keyword := range ai.exitKeywords {
				if strings.Contains(msg.Text, keyword) {
					return true
				}
			}
			return false
		},
		RunFunc: func(ctx context.Context, msg QueryMessage) error {
			return ai.exitAI(ctx)
		},
	})

	// AI 对话命令
	ai.AddCommand(&SimpleCommand{
		MatchFunc: func(msg QueryMessage) bool {
			// 如果在保持唤醒状态，所有消息都交给 AI
			if ai.IsKeepAlive() {
				return true
			}
			// 检查是否以召唤关键词开始
			for _, keyword := range ai.callAIKeywords {
				if strings.HasPrefix(msg.Text, keyword) {
					return true
				}
			}
			return false
		},
		RunFunc: func(ctx context.Context, msg QueryMessage) error {
			return ai.askAIForAnswer(ctx, msg)
		},
	})
}

// SetAskAI 设置 AI 问答函数
func (ai *AISpeaker) SetAskAI(askFunc func(context.Context, QueryMessage) (SpeakerAnswer, error)) {
	ai.askAI = askFunc
}

// enterAI 进入 AI 模式
func (ai *AISpeaker) enterAI(_ context.Context) error {
	if !ai.streamResponse {
		return ai.Response(SpeakerAnswer{
			Text: "您已关闭流式响应，无法使用连续对话模式",
		})
	}

	// 回应
	if len(ai.onEnterAI) > 0 {
		text := ai.pickOne(ai.onEnterAI)
		if err := ai.Response(SpeakerAnswer{
			Text:      text,
			KeepAlive: true,
		}); err != nil {
			return err
		}
	}

	// 唤醒
	return ai.EnterKeepAlive()
}

// exitAI 退出 AI 模式
func (ai *AISpeaker) exitAI(_ context.Context) error {
	// 退出唤醒状态
	if err := ai.ExitKeepAlive(); err != nil {
		return err
	}

	// 回应
	if len(ai.onExitAI) > 0 {
		text := ai.pickOne(ai.onExitAI)
		return ai.Response(SpeakerAnswer{
			Text:    text,
			PlaySFX: false,
		})
	}

	return nil
}

// askAIForAnswer 请求 AI 回答
func (ai *AISpeaker) askAIForAnswer(ctx context.Context, msg QueryMessage) error {
	if ai.askAI == nil {
		return ai.Response(SpeakerAnswer{
			Text: "AI 服务未初始化",
		})
	}

	// 显示思考中的提示
	if len(ai.onAIAsking) > 0 {
		thinkingText := ai.pickOne(ai.onAIAsking)
		if err := ai.Response(SpeakerAnswer{
			Text: thinkingText,
		}); err != nil {
			return err
		}
	}

	// 请求 AI 回答
	answer, err := ai.askAI(ctx, msg)
	if err != nil {
		logger.Errorf("AI 回答错误: %v", err)
		if len(ai.onAIError) > 0 {
			errorText := ai.pickOne(ai.onAIError)
			return ai.Response(SpeakerAnswer{
				Text: errorText,
			})
		}
		return err
	}

	// 回复 AI 的答案
	if answer.Text != "" {
		answer.KeepAlive = ai.IsKeepAlive()
		if err := ai.Response(answer); err != nil {
			return err
		}
	}

	// 显示回答完毕的提示
	if len(ai.onAIReplied) > 0 && ai.IsKeepAlive() {
		repliedText := ai.pickOne(ai.onAIReplied)
		return ai.Response(SpeakerAnswer{
			Text:      repliedText,
			KeepAlive: true,
		})
	}

	return nil
}

// pickOne 随机选择一个元素
func (ai *AISpeaker) pickOne(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[rand.Intn(len(items))]
}

// GetName 获取音箱名称
func (ai *AISpeaker) GetName() string {
	return ai.name
}

// SetName 设置音箱名称
func (ai *AISpeaker) SetName(name string) {
	ai.name = name
}

// SpeakerService 音箱服务类型别名
type SpeakerService = AISpeaker

// ProcessMessage 处理消息的便捷方法
func (ai *AISpeaker) ProcessMessage(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	msg := QueryMessage{
		Text:      text,
		Timestamp: time.Now(),
	}

	ctx := context.Background()

	// 如果包含AI关键词，调用AI
	for _, keyword := range ai.callAIKeywords {
		if strings.Contains(text, keyword) {
			if ai.askAI != nil {
				answer, err := ai.askAI(ctx, msg)
				if err != nil {
					logger.Errorf("AI回答失败: %v", err)
					return ai.pickOne(ai.onAIError), nil
				}
				return answer.Text, nil
			}
		}
	}

	// 检查是否是进入/退出命令
	for _, keyword := range ai.wakeUpKeywords {
		if strings.Contains(text, keyword) {
			ai.enterAI(ctx)
			return ai.pickOne(ai.onEnterAI), nil
		}
	}

	for _, keyword := range ai.exitKeywords {
		if strings.Contains(text, keyword) {
			ai.exitAI(ctx)
			return ai.pickOne(ai.onExitAI), nil
		}
	}

	// 默认无响应
	return "", nil
} 