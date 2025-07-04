package bot

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/internal/database"
	"mi-gpt-go/internal/models"
	"mi-gpt-go/internal/services/openai"
	"mi-gpt-go/internal/services/speaker"
	"mi-gpt-go/pkg/logger"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// MyBot 机器人服务
type MyBot struct {
	config   config.BotConfig     // 机器人配置
	speaker  *speaker.AISpeaker   // 音箱服务
	openai   *openai.Client       // AI客户端
	db       *gorm.DB             // 数据库连接
	ctx      context.Context      // 上下文
	cancel   context.CancelFunc   // 取消函数
	
	// 缓存的实体
	bot    *models.User // 机器人用户
	master *models.User // 主人用户
	room   *models.Room // 房间
}

// NewMyBot 创建新的机器人服务
func NewMyBot(cfg config.BotConfig, speakerService *speaker.AISpeaker, openaiConfig config.OpenAIConfig) *MyBot {
	// 创建AI客户端（支持OpenAI、DeepSeek等）
	openaiClient, err := openai.NewClient(openaiConfig)
	if err != nil {
		logger.Errorf("创建AI客户端失败: %v", err)
	}

	bot := &MyBot{
		config:  cfg,
		speaker: speakerService,
		openai:  openaiClient,
		db:      database.GetDB(),
	}

	// 设置音箱的 AI 问答函数
	speakerService.SetAskAI(bot.askAI)

	// 添加人设切换等特殊命令
	bot.addCustomCommands()

	return bot
}

// Start 启动机器人服务
func (b *MyBot) Start() error {
	logger.Info("机器人服务启动中...")

	// 创建上下文
	b.ctx, b.cancel = context.WithCancel(context.Background())

	// 初始化实体
	if err := b.initEntities(); err != nil {
		return fmt.Errorf("初始化实体失败: %v", err)
	}

	// 设置音箱名称
	if b.bot != nil {
		b.speaker.SetName(b.bot.Name)
	}

	// 启动音箱服务
	if err := b.speaker.Start(b.ctx); err != nil {
		return fmt.Errorf("启动音箱服务失败: %v", err)
	}

	logger.Info("机器人服务启动成功")
	return nil
}

// Stop 停止机器人服务
func (b *MyBot) Stop() error {
	logger.Info("机器人服务停止中...")

	if b.cancel != nil {
		b.cancel()
	}

	if b.speaker != nil {
		if err := b.speaker.Stop(); err != nil {
			logger.Errorf("停止音箱服务失败: %v", err)
		}
	}

	logger.Info("机器人服务已停止")
	return nil
}

// initEntities 初始化实体
func (b *MyBot) initEntities() error {
	// 创建或获取机器人用户
	bot := &models.User{
		Name:    b.config.Name,
		Profile: b.config.Profile,
	}
	if err := b.db.Where("name = ?", bot.Name).FirstOrCreate(bot).Error; err != nil {
		return fmt.Errorf("初始化机器人用户失败: %v", err)
	}
	b.bot = bot

	// 创建或获取主人用户
	master := &models.User{
		Name:    b.config.Master.Name,
		Profile: b.config.Master.Profile,
	}
	if err := b.db.Where("name = ?", master.Name).FirstOrCreate(master).Error; err != nil {
		return fmt.Errorf("初始化主人用户失败: %v", err)
	}
	b.master = master

	// 创建或获取房间
	room := &models.Room{
		Name:        b.config.Room.Name,
		Description: b.config.Room.Description,
	}
	if err := b.db.Where("name = ?", room.Name).FirstOrCreate(room).Error; err != nil {
		return fmt.Errorf("初始化房间失败: %v", err)
	}
	b.room = room

	return nil
}

// askAI AI 问答处理
func (b *MyBot) askAI(ctx context.Context, msg speaker.QueryMessage) (speaker.SpeakerAnswer, error) {
	if b.openai == nil {
		return speaker.SpeakerAnswer{}, fmt.Errorf("OpenAI 客户端未初始化")
	}

	// 保存用户消息
	if err := b.saveMessage(msg); err != nil {
		logger.Errorf("保存消息失败: %v", err)
	}

	// 构建系统提示词
	systemPrompt := b.buildSystemPrompt()

	// 构建用户提示词
	userPrompt := fmt.Sprintf("%s: %s", b.master.Name, msg.Text)

	// 调用 OpenAI
	options := openai.ChatOptions{
		User:   userPrompt,
		System: systemPrompt,
		Trace:  true,
	}

	var response string
	var err error

	if b.speaker.IsKeepAlive() && b.config.Name != "" {
		// 流式响应
		response, err = b.openai.ChatStream(ctx, options)
	} else {
		// 普通响应
		response, err = b.openai.Chat(ctx, options)
	}

	if err != nil {
		return speaker.SpeakerAnswer{}, err
	}

	// 保存机器人回复
	if err := b.saveBotMessage(response); err != nil {
		logger.Errorf("保存机器人消息失败: %v", err)
	}

	return speaker.SpeakerAnswer{
		Text: response,
	}, nil
}

// buildSystemPrompt 构建系统提示词
func (b *MyBot) buildSystemPrompt() string {
	template := `你是%s，%s。

## 对话环境
- 房间: %s (%s)
- 用户: %s (%s)

## 对话规则
1. 保持角色一致性，体现你的个性特点
2. 回答要简洁明了，适合语音交互
3. 用中文回复，语气要自然友好
4. 如果不确定用户意图，可以礼貌地询问`

	return fmt.Sprintf(template,
		b.bot.Name,
		b.bot.Profile,
		b.room.Name,
		b.room.Description,
		b.master.Name,
		b.master.Profile,
	)
}

// saveMessage 保存用户消息
func (b *MyBot) saveMessage(msg speaker.QueryMessage) error {
	message := &models.Message{
		Text:     msg.Text,
		SenderID: b.master.ID,
		RoomID:   b.room.ID,
	}

	return b.db.Create(message).Error
}

// saveBotMessage 保存机器人消息
func (b *MyBot) saveBotMessage(text string) error {
	message := &models.Message{
		Text:     text,
		SenderID: b.bot.ID,
		RoomID:   b.room.ID,
	}

	return b.db.Create(message).Error
}

// addCustomCommands 添加自定义命令
func (b *MyBot) addCustomCommands() {
	// 人设切换命令：你是xxx你xxx
	b.speaker.AddCommand(&speaker.SimpleCommand{
		MatchFunc: func(msg speaker.QueryMessage) bool {
			// 匹配格式：你是(姓名)你(描述)
			pattern := `.*你是([^你]*)你(.*)`
			matched, _ := regexp.MatchString(pattern, msg.Text)
			return matched
		},
		RunFunc: func(ctx context.Context, msg speaker.QueryMessage) error {
			return b.handleBotPersonalityChange(ctx, msg)
		},
	})

	// 主人信息切换命令：我是xxx我xxx
	b.speaker.AddCommand(&speaker.SimpleCommand{
		MatchFunc: func(msg speaker.QueryMessage) bool {
			// 匹配格式：我是(姓名)我(描述)
			pattern := `.*我是([^我]*)我(.*)`
			matched, _ := regexp.MatchString(pattern, msg.Text)
			return matched
		},
		RunFunc: func(ctx context.Context, msg speaker.QueryMessage) error {
			return b.handleMasterInfoChange(ctx, msg)
		},
	})
}

// handleBotPersonalityChange 处理机器人人设切换
func (b *MyBot) handleBotPersonalityChange(_ context.Context, msg speaker.QueryMessage) error {
	// 使用正则表达式提取姓名和描述
	pattern := `.*你是([^你]*)你(.*)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(msg.Text)
	
	if len(matches) < 3 {
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "格式不正确，请使用\"你是[姓名]你[描述]\"的格式",
		})
	}

	newName := strings.TrimSpace(matches[1])
	newProfile := strings.TrimSpace(matches[2])

	if newName == "" || newProfile == "" {
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "姓名和描述都不能为空",
		})
	}

	// 更新机器人信息
	b.bot.Name = newName
	b.bot.Profile = newProfile
	if err := b.db.Save(b.bot).Error; err != nil {
		logger.Errorf("更新机器人信息失败: %v", err)
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "更新失败，请稍后再试",
		})
	}

	// 更新音箱名称
	b.speaker.SetName(newName)

	logger.Infof("机器人人设已更新 - 姓名: %s, 描述: %s", newName, newProfile)
	
	return b.speaker.Response(speaker.SpeakerAnswer{
		Text: fmt.Sprintf("好的，我现在是%s了！%s", newName, newProfile),
	})
}

// handleMasterInfoChange 处理主人信息切换
func (b *MyBot) handleMasterInfoChange(_ context.Context, msg speaker.QueryMessage) error {
	// 使用正则表达式提取姓名和描述
	pattern := `.*我是([^我]*)我(.*)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(msg.Text)
	
	if len(matches) < 3 {
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "格式不正确，请使用\"我是[姓名]我[描述]\"的格式",
		})
	}

	newName := strings.TrimSpace(matches[1])
	newProfile := strings.TrimSpace(matches[2])

	if newName == "" || newProfile == "" {
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "姓名和描述都不能为空",
		})
	}

	// 更新主人信息
	b.master.Name = newName
	b.master.Profile = newProfile
	if err := b.db.Save(b.master).Error; err != nil {
		logger.Errorf("更新主人信息失败: %v", err)
		return b.speaker.Response(speaker.SpeakerAnswer{
			Text: "更新失败，请稍后再试",
		})
	}

	logger.Infof("主人信息已更新 - 姓名: %s, 描述: %s", newName, newProfile)
	
	return b.speaker.Response(speaker.SpeakerAnswer{
		Text: fmt.Sprintf("知道了，%s！我已经记住你的信息了：%s", newName, newProfile),
	})
} 