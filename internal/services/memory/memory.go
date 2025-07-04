package memory

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/models"
	"mi-gpt-go/internal/services/openai"
	"mi-gpt-go/pkg/logger"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MemoryManager 记忆管理器
type MemoryManager struct {
	db       *gorm.DB           // 数据库连接
	aiClient *openai.Client     // AI客户端，用于生成记忆摘要
}

// NewMemoryManager 创建记忆管理器
func NewMemoryManager(db *gorm.DB, aiClient *openai.Client) *MemoryManager {
	return &MemoryManager{
		db:       db,
		aiClient: aiClient,
	}
}

// AddMessageMemory 为消息添加记忆
func (mm *MemoryManager) AddMessageMemory(ctx context.Context, message *models.Message, ownerID *string, roomID string) error {
	// 创建基础记忆记录
	memory := &models.Memory{
		MessageID: message.ID,
		OwnerID:   ownerID,
		RoomID:    roomID,
	}

	if err := mm.db.Create(memory).Error; err != nil {
		return fmt.Errorf("创建记忆失败: %v", err)
	}

	logger.Debugf("为消息 %d 创建了记忆记录 %d", message.ID, memory.ID)
	return nil
}

// UpdateShortTermMemory 更新短期记忆
func (mm *MemoryManager) UpdateShortTermMemory(ctx context.Context, ownerID *string, roomID string, messageLimit int) error {
	// 获取最近的消息
	var messages []models.Message
	query := mm.db.Preload("Sender").Where("room_id = ?", roomID).Order("created_at DESC").Limit(messageLimit)
	if err := query.Find(&messages).Error; err != nil {
		return fmt.Errorf("获取最近消息失败: %v", err)
	}

	if len(messages) == 0 {
		return nil
	}

	// 生成短期记忆摘要
	summary, err := mm.generateMemorySummary(ctx, messages, "short")
	if err != nil {
		logger.Errorf("生成短期记忆摘要失败: %v", err)
		return err
	}

	// 获取最新的记忆记录作为游标
	var latestMemory models.Memory
	if err := mm.db.Where("room_id = ? AND owner_id = ?", roomID, ownerID).
		Order("created_at DESC").First(&latestMemory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("未找到记忆记录，跳过短期记忆更新")
			return nil
		}
		return fmt.Errorf("获取最新记忆失败: %v", err)
	}

	// 创建或更新短期记忆
	shortTermMemory := &models.ShortTermMemory{
		Text:     summary,
		CursorID: latestMemory.ID,
		OwnerID:  ownerID,
		RoomID:   roomID,
	}

	// 尝试更新现有的短期记忆，如果不存在则创建新的
	var existingMemory models.ShortTermMemory
	if err := mm.db.Where("owner_id = ? AND room_id = ?", ownerID, roomID).
		Order("created_at DESC").First(&existingMemory).Error; err == nil {
		// 更新现有记忆
		existingMemory.Text = summary
		existingMemory.CursorID = latestMemory.ID
		if err := mm.db.Save(&existingMemory).Error; err != nil {
			return fmt.Errorf("更新短期记忆失败: %v", err)
		}
		logger.Debug("已更新短期记忆")
	} else {
		// 创建新记忆
		if err := mm.db.Create(shortTermMemory).Error; err != nil {
			return fmt.Errorf("创建短期记忆失败: %v", err)
		}
		logger.Debug("已创建新的短期记忆")
	}

	return nil
}

// UpdateLongTermMemory 更新长期记忆
func (mm *MemoryManager) UpdateLongTermMemory(ctx context.Context, ownerID *string, roomID string) error {
	// 获取多个短期记忆
	var shortTermMemories []models.ShortTermMemory
	if err := mm.db.Where("owner_id = ? AND room_id = ?", ownerID, roomID).
		Order("created_at DESC").Limit(5).Find(&shortTermMemories).Error; err != nil {
		return fmt.Errorf("获取短期记忆失败: %v", err)
	}

	if len(shortTermMemories) < 2 {
		logger.Debug("短期记忆数量不足，跳过长期记忆更新")
		return nil
	}

	// 合并短期记忆文本
	var memoryTexts []string
	for _, memory := range shortTermMemories {
		memoryTexts = append(memoryTexts, memory.Text)
	}
	combinedText := strings.Join(memoryTexts, "\n")

	// 生成长期记忆摘要
	summary, err := mm.generateLongTermSummary(ctx, combinedText)
	if err != nil {
		logger.Errorf("生成长期记忆摘要失败: %v", err)
		return err
	}

	// 使用最新的短期记忆作为游标
	latestShortTerm := shortTermMemories[0]

	// 创建长期记忆
	longTermMemory := &models.LongTermMemory{
		Text:     summary,
		CursorID: latestShortTerm.ID,
		OwnerID:  ownerID,
		RoomID:   roomID,
	}

	if err := mm.db.Create(longTermMemory).Error; err != nil {
		return fmt.Errorf("创建长期记忆失败: %v", err)
	}

	logger.Debug("已创建新的长期记忆")
	return nil
}

// GetShortTermMemories 获取短期记忆
func (mm *MemoryManager) GetShortTermMemories(ownerID *string, roomID string, limit int) ([]models.ShortTermMemory, error) {
	var memories []models.ShortTermMemory
	query := mm.db.Where("owner_id = ? AND room_id = ?", ownerID, roomID).
		Order("created_at DESC").Limit(limit)
	
	if err := query.Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("获取短期记忆失败: %v", err)
	}

	return memories, nil
}

// GetLongTermMemories 获取长期记忆
func (mm *MemoryManager) GetLongTermMemories(ownerID *string, roomID string, limit int) ([]models.LongTermMemory, error) {
	var memories []models.LongTermMemory
	query := mm.db.Where("owner_id = ? AND room_id = ?", ownerID, roomID).
		Order("created_at DESC").Limit(limit)
	
	if err := query.Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("获取长期记忆失败: %v", err)
	}

	return memories, nil
}

// generateMemorySummary 生成记忆摘要
func (mm *MemoryManager) generateMemorySummary(ctx context.Context, messages []models.Message, memoryType string) (string, error) {
	if mm.aiClient == nil {
		return "AI客户端未初始化，无法生成记忆摘要", nil
	}

	// 构建对话内容
	var dialogues []string
	for i := len(messages) - 1; i >= 0; i-- { // 按时间顺序排列
		msg := messages[i]
		dialogues = append(dialogues, fmt.Sprintf("%s: %s", msg.Sender.Name, msg.Text))
	}
	conversationText := strings.Join(dialogues, "\n")

	// 生成摘要提示词
	var systemPrompt string
	switch memoryType {
	case "short":
		systemPrompt = `你是一个记忆管理助手。请根据以下对话内容，生成一个简洁的短期记忆摘要。
摘要应该：
1. 提取对话中的关键信息和重要事件
2. 保留情感色彩和互动细节
3. 长度控制在100字以内
4. 使用第三人称描述

请只返回摘要内容，不要包含其他说明。`

	default:
		systemPrompt = `你是一个记忆管理助手。请根据以下对话内容，生成一个记忆摘要。
摘要应该：
1. 提取关键信息
2. 保持简洁明了
3. 长度控制在50字以内

请只返回摘要内容。`
	}

	userPrompt := fmt.Sprintf("以下是需要总结的对话内容：\n\n%s", conversationText)

	// 调用AI生成摘要
	options := openai.ChatOptions{
		User:   userPrompt,
		System: systemPrompt,
		Trace:  false,
	}

	summary, err := mm.aiClient.Chat(ctx, options)
	if err != nil {
		return "", fmt.Errorf("AI生成摘要失败: %v", err)
	}

	return strings.TrimSpace(summary), nil
}

// generateLongTermSummary 生成长期记忆摘要
func (mm *MemoryManager) generateLongTermSummary(ctx context.Context, shortTermMemories string) (string, error) {
	if mm.aiClient == nil {
		return "AI客户端未初始化，无法生成长期记忆摘要", nil
	}

	systemPrompt := `你是一个记忆管理助手。请根据多个短期记忆内容，生成一个综合的长期记忆摘要。
长期记忆摘要应该：
1. 整合多个短期记忆中的关键信息
2. 识别和保留重要的模式、偏好和特征
3. 去除重复和临时性信息
4. 长度控制在150字以内
5. 有助于未来对话的个性化

请只返回摘要内容，不要包含其他说明。`

	userPrompt := fmt.Sprintf("以下是需要整合的短期记忆内容：\n\n%s", shortTermMemories)

	options := openai.ChatOptions{
		User:   userPrompt,
		System: systemPrompt,
		Trace:  false,
	}

	summary, err := mm.aiClient.Chat(ctx, options)
	if err != nil {
		return "", fmt.Errorf("AI生成长期记忆摘要失败: %v", err)
	}

	return strings.TrimSpace(summary), nil
}

// CleanOldMemories 清理过期的记忆（可选功能）
func (mm *MemoryManager) CleanOldMemories(ctx context.Context, daysToKeep int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)

	// 清理过期的短期记忆
	if err := mm.db.Where("created_at < ?", cutoffDate).Delete(&models.ShortTermMemory{}).Error; err != nil {
		logger.Errorf("清理过期短期记忆失败: %v", err)
	}

	// 保留长期记忆更长时间
	longTermCutoff := time.Now().AddDate(0, 0, -daysToKeep*3)
	if err := mm.db.Where("created_at < ?", longTermCutoff).Delete(&models.LongTermMemory{}).Error; err != nil {
		logger.Errorf("清理过期长期记忆失败: %v", err)
	}

	logger.Infof("已清理 %d 天前的过期记忆", daysToKeep)
	return nil
} 