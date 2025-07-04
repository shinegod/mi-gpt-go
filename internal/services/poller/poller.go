package poller

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/services/miservice"
	"mi-gpt-go/internal/services/speaker"
	"mi-gpt-go/pkg/logger"
	"sync"
	"time"
)

// MessagePoller 消息拉取器
type MessagePoller struct {
	miService      miservice.MiServiceInterface
	speakerService *speaker.SpeakerService
	interval       time.Duration
	lastTimestamp  int64
	running        bool
	mutex          sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewMessagePoller 创建消息拉取器
func NewMessagePoller(miSvc miservice.MiServiceInterface, speakerSvc *speaker.SpeakerService, interval time.Duration) *MessagePoller {
	return &MessagePoller{
		miService:      miSvc,
		speakerService: speakerSvc,
		interval:       interval,
		lastTimestamp:  time.Now().UnixMilli(),
		stopChan:       make(chan struct{}),
	}
}

// Start 启动消息拉取
func (mp *MessagePoller) Start(ctx context.Context) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.running {
		return fmt.Errorf("消息拉取器已在运行")
	}

	mp.running = true
	mp.wg.Add(1)

	go mp.pollLoop(ctx)

	logger.Infof("消息拉取器已启动，检查间隔: %v", mp.interval)
	return nil
}

// Stop 停止消息拉取
func (mp *MessagePoller) Stop() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if !mp.running {
		return nil
	}

	mp.running = false
	close(mp.stopChan)
	mp.wg.Wait()

	logger.Info("消息拉取器已停止")
	return nil
}

// IsRunning 检查是否正在运行
func (mp *MessagePoller) IsRunning() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.running
}

// pollLoop 消息拉取循环
func (mp *MessagePoller) pollLoop(ctx context.Context) {
	defer mp.wg.Done()

	ticker := time.NewTicker(mp.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("消息拉取器接收到上下文取消信号")
			return
		case <-mp.stopChan:
			logger.Debug("消息拉取器接收到停止信号")
			return
		case <-ticker.C:
			if err := mp.pollMessages(); err != nil {
				logger.Errorf("拉取消息失败: %v", err)
			}
		}
	}
}

// pollMessages 拉取消息
func (mp *MessagePoller) pollMessages() error {
	// 使用SafeIsPlaying检查播放状态
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	isPlaying, err := mp.miService.SafeIsPlaying(ctx)
	if err != nil {
		logger.Debugf("检查播放状态失败: %v", err)
		// 即使检查播放状态失败，也继续拉取消息
	}

	// 如果设备正在播放，跳过此次拉取
	if isPlaying {
		logger.Debug("设备正在播放，跳过消息拉取")
		return nil
	}

	// 使用SafeGetMessages获取最新消息
	options := map[string]interface{}{
		"limit": 10,
	}

	messages, err := mp.miService.SafeGetMessages(ctx, options)
	if err != nil {
		return fmt.Errorf("获取消息失败: %v", err)
	}

	// 过滤新消息
	newMessages := mp.filterNewMessages(messages)
	if len(newMessages) == 0 {
		return nil
	}

	logger.Debugf("获取到 %d 条新消息", len(newMessages))

	// 处理每条新消息
	for _, message := range newMessages {
		if err := mp.processMessage(message); err != nil {
			logger.Errorf("处理消息失败: %v", err)
			continue
		}

		// 更新最后处理的时间戳 - 从interface{}转换为queryMessage
		if queryMsg, ok := message.(miservice.QueryMessage); ok {
			if queryMsg.Timestamp > mp.lastTimestamp {
				mp.lastTimestamp = queryMsg.Timestamp
			}
		}
	}

	return nil
}

// filterNewMessages 过滤新消息
func (mp *MessagePoller) filterNewMessages(messages []interface{}) []interface{} {
	var newMessages []interface{}

	for _, message := range messages {
		// 尝试转换为QueryMessage类型
		if queryMsg, ok := message.(miservice.QueryMessage); ok {
			// 只处理比上次时间戳更新的消息
			if queryMsg.Timestamp > mp.lastTimestamp {
				newMessages = append(newMessages, message)
			}
		}
	}

	return newMessages
}

// processMessage 处理单条消息
func (mp *MessagePoller) processMessage(message interface{}) error {
	// 转换消息类型
	queryMsg, ok := message.(miservice.QueryMessage)
	if !ok {
		return fmt.Errorf("消息类型转换失败")
	}
	
	if queryMsg.Text == "" {
		return nil
	}

	logger.Debugf("处理消息: %s", queryMsg.Text)

	// 通过音箱服务处理消息
	response, err := mp.speakerService.ProcessMessage(queryMsg.Text)
	if err != nil {
		return fmt.Errorf("音箱服务处理消息失败: %v", err)
	}

	// 如果有响应内容，播放TTS
	if response != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := mp.miService.SafePlayTTS(ctx, response); err != nil {
			return fmt.Errorf("播放TTS失败: %v", err)
		}

		logger.Debugf("已播放响应: %s", response)
	}

	return nil
}

// SetLastTimestamp 设置最后处理的时间戳
func (mp *MessagePoller) SetLastTimestamp(timestamp int64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.lastTimestamp = timestamp
}

// GetLastTimestamp 获取最后处理的时间戳
func (mp *MessagePoller) GetLastTimestamp() int64 {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.lastTimestamp
}

// GetStats 获取统计信息
func (mp *MessagePoller) GetStats() map[string]interface{} {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	return map[string]interface{}{
		"running":        mp.running,
		"interval":       mp.interval.String(),
		"lastTimestamp":  mp.lastTimestamp,
		"lastCheckTime":  time.Now().Format("2006-01-02 15:04:05"),
	}
} 