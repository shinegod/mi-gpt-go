package speaker

import (
	"context"
	"fmt"
	"mi-gpt-go/internal/services/miservice"
	"mi-gpt-go/internal/utils"
	"mi-gpt-go/pkg/logger"
	"sync"
	"time"
)

// ConcurrentConfig 并发配置
type ConcurrentConfig struct {
	WorkerCount       int           // 工作协程数量
	QueueSize         int           // 任务队列大小
	MessageBufferSize int           // 消息缓冲区大小
	RateLimit         int           // 速率限制（每秒任务数）
	BatchSize         int           // 批处理大小
	BatchTimeout      time.Duration // 批处理超时
	EnableMetrics     bool          // 是否启用指标统计
}

// DefaultConcurrentConfig 默认并发配置
var DefaultConcurrentConfig = ConcurrentConfig{
	WorkerCount:       4,
	QueueSize:         100,
	MessageBufferSize: 50,
	RateLimit:         10,
	BatchSize:         5,
	BatchTimeout:      5 * time.Second,
	EnableMetrics:     true,
}

// MessageProcessJob 消息处理任务
type MessageProcessJob struct {
	ID        string
	Message   miservice.QueryMessage
	Speaker   *EnhancedAISpeaker
	Priority  int
	Timestamp time.Time
}

// NewMessageProcessJob 创建消息处理任务
func NewMessageProcessJob(id string, message miservice.QueryMessage, speaker *EnhancedAISpeaker, priority int) *MessageProcessJob {
	return &MessageProcessJob{
		ID:        id,
		Message:   message,
		Speaker:   speaker,
		Priority:  priority,
		Timestamp: time.Now(),
	}
}

// Execute 执行消息处理任务
func (job *MessageProcessJob) Execute(ctx context.Context) error {
	logger.Debugf("开始处理消息任务: %s, 内容: %s", job.ID, job.Message.Text)
	
	// 处理消息逻辑 - handleMessage只接受string参数
	job.Speaker.handleMessage(job.Message.Text)
	return nil
}

// GetID 获取任务ID
func (job *MessageProcessJob) GetID() string {
	return job.ID
}

// GetPriority 获取任务优先级
func (job *MessageProcessJob) GetPriority() int {
	return job.Priority
}

// TTSPlayJob TTS播放任务
type TTSPlayJob struct {
	ID        string
	Text      string
	Service   miservice.MiServiceInterface
	Priority  int
	Timestamp time.Time
}

// NewTTSPlayJob 创建TTS播放任务
func NewTTSPlayJob(id string, text string, service miservice.MiServiceInterface, priority int) *TTSPlayJob {
	return &TTSPlayJob{
		ID:        id,
		Text:      text,
		Service:   service,
		Priority:  priority,
		Timestamp: time.Now(),
	}
}

// Execute 执行TTS播放任务
func (job *TTSPlayJob) Execute(ctx context.Context) error {
	logger.Debugf("开始TTS播放任务: %s, 内容: %s", job.ID, job.Text)
	
	return job.Service.SafePlayTTS(ctx, job.Text)
}

// GetID 获取任务ID
func (job *TTSPlayJob) GetID() string {
	return job.ID
}

// GetPriority 获取任务优先级
func (job *TTSPlayJob) GetPriority() int {
	return job.Priority
}

// 任务优先级常量
const (
	PriorityUrgent = 100 // 紧急任务（如紧急停止）
	PriorityHigh   = 80  // 高优先级（如用户命令）
	PriorityNormal = 50  // 普通优先级（如常规消息处理）
	PriorityLow    = 30  // 低优先级（如健康检查）
	PriorityBatch  = 10  // 批处理任务
)

// ConcurrentHandler 并发处理器
type ConcurrentHandler struct {
	config         ConcurrentConfig
	workerPool     *utils.WorkerPool
	messageBuffer  *utils.MessageBuffer
	rateLimiter    *utils.RateLimiter
	speaker        *EnhancedAISpeaker
	miService      miservice.MiServiceInterface
	
	// 批处理相关
	batchJobs      []utils.Job
	batchMutex     sync.Mutex
	batchTimer     *time.Timer
	
	// 状态管理
	ctx            context.Context
	cancel         context.CancelFunc
	isRunning      bool
	mutex          sync.RWMutex
}

// NewConcurrentHandler 创建并发处理器
func NewConcurrentHandler(config ConcurrentConfig, speaker *EnhancedAISpeaker, miService miservice.MiServiceInterface) *ConcurrentHandler {
	ctx, cancel := context.WithCancel(context.Background())
	
	ch := &ConcurrentHandler{
		config:     config,
		speaker:    speaker,
		miService:  miService,
		ctx:        ctx,
		cancel:     cancel,
		batchJobs:  make([]utils.Job, 0, config.BatchSize),
	}
	
	// 创建组件
	ch.workerPool = utils.NewWorkerPool(config.WorkerCount, config.QueueSize)
	ch.messageBuffer = utils.NewMessageBuffer(config.MessageBufferSize)
	ch.rateLimiter = utils.NewRateLimiter(config.RateLimit)
	
	logger.Infof("并发处理器已创建，工作协程: %d，队列大小: %d", 
		config.WorkerCount, config.QueueSize)
	
	return ch
}

// Start 启动并发处理器
func (ch *ConcurrentHandler) Start() error {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	
	if ch.isRunning {
		return fmt.Errorf("并发处理器已在运行")
	}
	
	logger.Info("启动并发处理器...")
	
	// 启动工作池
	if err := ch.workerPool.Start(); err != nil {
		return fmt.Errorf("启动工作池失败: %v", err)
	}
	
	// 启动速率限制器
	ch.rateLimiter.Start()
	
	// 启动批处理器
	ch.startBatchProcessor()
	
	// 启动消息处理器
	ch.startMessageProcessor()
	
	ch.isRunning = true
	logger.Info("并发处理器启动成功")
	return nil
}

// Stop 停止并发处理器
func (ch *ConcurrentHandler) Stop() error {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	
	if !ch.isRunning {
		return nil
	}
	
	logger.Info("停止并发处理器...")
	
	// 取消上下文
	ch.cancel()
	
	// 停止批处理定时器
	if ch.batchTimer != nil {
		ch.batchTimer.Stop()
	}
	
	// 处理剩余的批处理任务
	ch.flushBatchJobs()
	
	// 停止速率限制器
	ch.rateLimiter.Stop()
	
	// 停止工作池
	if err := ch.workerPool.Stop(); err != nil {
		logger.Warnf("停止工作池失败: %v", err)
	}
	
	ch.isRunning = false
	logger.Info("并发处理器已停止")
	return nil
}

// SubmitMessage 提交消息处理任务
func (ch *ConcurrentHandler) SubmitMessage(message miservice.QueryMessage, priority int) error {
	if !ch.isRunning {
		return fmt.Errorf("并发处理器未运行")
	}
	
	// 等待速率限制
	if err := ch.rateLimiter.Wait(ch.ctx); err != nil {
		return fmt.Errorf("速率限制等待失败: %v", err)
	}
	
	// 创建任务
	jobID := fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), message.Text[:min(10, len(message.Text))])
	job := NewMessageProcessJob(jobID, message, ch.speaker, priority)
	
	// 提交任务
	return ch.submitJob(job)
}

// SubmitTTS 提交TTS播放任务
func (ch *ConcurrentHandler) SubmitTTS(text string, priority int) error {
	if !ch.isRunning {
		return fmt.Errorf("并发处理器未运行")
	}
	
	// 等待速率限制
	if err := ch.rateLimiter.Wait(ch.ctx); err != nil {
		return fmt.Errorf("速率限制等待失败: %v", err)
	}
	
	// 创建任务
	jobID := fmt.Sprintf("tts_%d_%s", time.Now().UnixNano(), text[:min(10, len(text))])
	job := NewTTSPlayJob(jobID, text, ch.miService, priority)
	
	// 提交任务
	return ch.submitJob(job)
}

// submitJob 提交任务（内部方法）
func (ch *ConcurrentHandler) submitJob(job utils.Job) error {
	// 检查是否启用批处理
	if ch.config.BatchSize > 1 && job.GetPriority() <= PriorityNormal {
		return ch.addToBatch(job)
	}
	
	// 直接提交高优先级任务
	return ch.workerPool.Submit(job)
}

// addToBatch 添加任务到批处理
func (ch *ConcurrentHandler) addToBatch(job utils.Job) error {
	ch.batchMutex.Lock()
	defer ch.batchMutex.Unlock()
	
	ch.batchJobs = append(ch.batchJobs, job)
	
	// 检查是否达到批处理大小
	if len(ch.batchJobs) >= ch.config.BatchSize {
		ch.flushBatchJobsLocked()
		return nil
	}
	
	// 设置批处理定时器
	if ch.batchTimer == nil {
		ch.batchTimer = time.AfterFunc(ch.config.BatchTimeout, func() {
			ch.batchMutex.Lock()
			defer ch.batchMutex.Unlock()
			ch.flushBatchJobsLocked()
		})
	}
	
	return nil
}

// flushBatchJobs 刷新批处理任务
func (ch *ConcurrentHandler) flushBatchJobs() {
	ch.batchMutex.Lock()
	defer ch.batchMutex.Unlock()
	ch.flushBatchJobsLocked()
}

// flushBatchJobsLocked 刷新批处理任务（已锁定）
func (ch *ConcurrentHandler) flushBatchJobsLocked() {
	if len(ch.batchJobs) == 0 {
		return
	}
	
	// 并发提交所有批处理任务
	for _, job := range ch.batchJobs {
		if err := ch.workerPool.Submit(job); err != nil {
			logger.Errorf("提交批处理任务失败: %v", err)
		}
	}
	
	logger.Debugf("提交批处理任务: %d 个", len(ch.batchJobs))
	
	// 清空批处理队列
	ch.batchJobs = make([]utils.Job, 0, ch.config.BatchSize)
	
	// 重置定时器
	if ch.batchTimer != nil {
		ch.batchTimer.Stop()
		ch.batchTimer = nil
	}
}

// startBatchProcessor 启动批处理器
func (ch *ConcurrentHandler) startBatchProcessor() {
	utils.SafeGoWithContext(ch.ctx, "batch-processor", func(ctx context.Context) {
		ticker := time.NewTicker(ch.config.BatchTimeout / 2)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				logger.Debug("批处理器已停止")
				return
			case <-ticker.C:
				// 定期检查并刷新批处理任务
				ch.batchMutex.Lock()
				if len(ch.batchJobs) > 0 {
					// 检查第一个任务是否超时
					if job, ok := ch.batchJobs[0].(*MessageProcessJob); ok {
						if time.Since(job.Timestamp) > ch.config.BatchTimeout {
							ch.flushBatchJobsLocked()
						}
					}
				}
				ch.batchMutex.Unlock()
			}
		}
	})
}

// startMessageProcessor 启动消息处理器
func (ch *ConcurrentHandler) startMessageProcessor() {
	utils.SafeGoWithContext(ch.ctx, "message-processor", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				logger.Debug("消息处理器已停止")
				return
			default:
				// 尝试从消息缓冲区获取消息
				if msg := ch.messageBuffer.TryGet(); msg != nil {
					if message, ok := msg.(miservice.QueryMessage); ok {
						ch.SubmitMessage(message, PriorityNormal)
					}
				} else {
					// 没有消息时短暂休眠
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	})
}

// PutMessage 放入消息到缓冲区
func (ch *ConcurrentHandler) PutMessage(message miservice.QueryMessage) {
	ch.messageBuffer.Put(message)
}

// GetStatus 获取并发处理状态
func (ch *ConcurrentHandler) GetStatus() map[string]interface{} {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()
	
	status := map[string]interface{}{
		"running":        ch.isRunning,
		"config":         ch.config,
		"messageBuffer":  ch.messageBuffer.Len(),
		"batchJobs":      len(ch.batchJobs),
	}
	
	if ch.workerPool != nil {
		status["workerPool"] = ch.workerPool.GetStatus()
	}
	
	return status
}

// min 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 