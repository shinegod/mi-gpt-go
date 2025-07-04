package utils

import (
	"context"
	"fmt"
	"mi-gpt-go/pkg/logger"
	"sync"
	"time"
)

// WorkerPool 工作池
type WorkerPool struct {
	workerCount int              // 工作协程数量
	jobQueue    chan Job         // 任务队列
	workers     []*Worker        // 工作协程列表
	wg          sync.WaitGroup   // 等待组
	ctx         context.Context  // 上下文
	cancel      context.CancelFunc // 取消函数
	mutex       sync.RWMutex     // 读写锁
	isRunning   bool             // 运行状态
}

// Job 任务接口
type Job interface {
	Execute(ctx context.Context) error
	GetID() string
	GetPriority() int
}

// Worker 工作协程
type Worker struct {
	id       int
	pool     *WorkerPool
	jobQueue chan Job
	quit     chan bool
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		workers:     make([]*Worker, workerCount),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.isRunning {
		return fmt.Errorf("工作池已在运行")
	}

	logger.Infof("启动工作池，工作协程数: %d", wp.workerCount)

	// 创建并启动工作协程
	for i := 0; i < wp.workerCount; i++ {
		worker := &Worker{
			id:       i,
			pool:     wp,
			jobQueue: wp.jobQueue,
			quit:     make(chan bool),
		}
		wp.workers[i] = worker
		wp.wg.Add(1)
		go worker.start()
	}

	wp.isRunning = true
	return nil
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if !wp.isRunning {
		return nil
	}

	logger.Info("正在停止工作池...")

	// 取消上下文
	wp.cancel()

	// 停止所有工作协程
	for _, worker := range wp.workers {
		if worker != nil {
			worker.stop()
		}
	}

	// 等待所有协程完成
	wp.wg.Wait()

	// 关闭任务队列
	close(wp.jobQueue)

	wp.isRunning = false
	logger.Info("工作池已停止")
	return nil
}

// Submit 提交任务
func (wp *WorkerPool) Submit(job Job) error {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.isRunning {
		return fmt.Errorf("工作池未运行")
	}

	select {
	case wp.jobQueue <- job:
		logger.Debugf("任务已提交: %s", job.GetID())
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("工作池已停止")
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// GetStatus 获取工作池状态
func (wp *WorkerPool) GetStatus() map[string]interface{} {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	return map[string]interface{}{
		"running":      wp.isRunning,
		"workerCount":  wp.workerCount,
		"queueLength":  len(wp.jobQueue),
		"queueCapacity": cap(wp.jobQueue),
	}
}

// start 启动工作协程
func (w *Worker) start() {
	defer w.pool.wg.Done()
	
	logger.Debugf("工作协程 %d 已启动", w.id)

	for {
		select {
		case job := <-w.jobQueue:
			if job != nil {
				w.executeJob(job)
			}
		case <-w.quit:
			logger.Debugf("工作协程 %d 收到停止信号", w.id)
			return
		case <-w.pool.ctx.Done():
			logger.Debugf("工作协程 %d 上下文已取消", w.id)
			return
		}
	}
}

// stop 停止工作协程
func (w *Worker) stop() {
	close(w.quit)
}

// executeJob 执行任务
func (w *Worker) executeJob(job Job) {
	startTime := time.Now()
	logger.Debugf("工作协程 %d 开始执行任务: %s", w.id, job.GetID())

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(w.pool.ctx, 30*time.Second)
	defer cancel()

	// 执行任务
	err := job.Execute(ctx)
	duration := time.Since(startTime)

	if err != nil {
		logger.Errorf("工作协程 %d 执行任务失败: %s, 耗时: %v, 错误: %v", 
			w.id, job.GetID(), duration, err)
	} else {
		logger.Debugf("工作协程 %d 执行任务成功: %s, 耗时: %v", 
			w.id, job.GetID(), duration)
	}
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	items []Job
	mutex sync.RWMutex
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]Job, 0),
	}
}

// Push 添加任务
func (pq *PriorityQueue) Push(job Job) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// 按优先级插入
	inserted := false
	for i, item := range pq.items {
		if job.GetPriority() > item.GetPriority() {
			// 插入到当前位置
			pq.items = append(pq.items[:i], append([]Job{job}, pq.items[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.items = append(pq.items, job)
	}
}

// Pop 取出任务
func (pq *PriorityQueue) Pop() Job {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	job := pq.items[0]
	pq.items = pq.items[1:]
	return job
}

// Len 获取队列长度
func (pq *PriorityQueue) Len() int {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return len(pq.items)
}

// MessageBuffer 消息缓冲区
type MessageBuffer struct {
	buffer    []interface{}
	capacity  int
	mutex     sync.RWMutex
	notEmpty  *sync.Cond
	notFull   *sync.Cond
}

// NewMessageBuffer 创建消息缓冲区
func NewMessageBuffer(capacity int) *MessageBuffer {
	mb := &MessageBuffer{
		buffer:   make([]interface{}, 0, capacity),
		capacity: capacity,
	}
	mb.notEmpty = sync.NewCond(&mb.mutex)
	mb.notFull = sync.NewCond(&mb.mutex)
	return mb
}

// Put 放入消息
func (mb *MessageBuffer) Put(msg interface{}) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// 等待缓冲区有空间
	for len(mb.buffer) >= mb.capacity {
		mb.notFull.Wait()
	}

	mb.buffer = append(mb.buffer, msg)
	mb.notEmpty.Signal()
}

// Get 取出消息
func (mb *MessageBuffer) Get() interface{} {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// 等待缓冲区有消息
	for len(mb.buffer) == 0 {
		mb.notEmpty.Wait()
	}

	msg := mb.buffer[0]
	mb.buffer = mb.buffer[1:]
	mb.notFull.Signal()
	return msg
}

// TryGet 尝试取出消息（非阻塞）
func (mb *MessageBuffer) TryGet() interface{} {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	if len(mb.buffer) == 0 {
		return nil
	}

	msg := mb.buffer[0]
	mb.buffer = mb.buffer[1:]
	mb.notFull.Signal()
	return msg
}

// Len 获取缓冲区长度
func (mb *MessageBuffer) Len() int {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	return len(mb.buffer)
}

// RateLimiter 速率限制器
type RateLimiter struct {
	rate     int           // 每秒允许的请求数
	bucket   chan struct{} // 令牌桶
	ticker   *time.Ticker  // 定时器
	ctx      context.Context
	cancel   context.CancelFunc
	isRunning bool
	mutex    sync.RWMutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RateLimiter{
		rate:   rate,
		bucket: make(chan struct{}, rate),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动速率限制器
func (rl *RateLimiter) Start() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.isRunning {
		return
	}

	// 初始化令牌桶
	for i := 0; i < rl.rate; i++ {
		rl.bucket <- struct{}{}
	}

	// 启动令牌补充
	rl.ticker = time.NewTicker(time.Second / time.Duration(rl.rate))
	go rl.refillTokens()

	rl.isRunning = true
	logger.Debugf("速率限制器已启动，速率: %d/秒", rl.rate)
}

// Stop 停止速率限制器
func (rl *RateLimiter) Stop() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if !rl.isRunning {
		return
	}

	rl.cancel()
	if rl.ticker != nil {
		rl.ticker.Stop()
	}

	rl.isRunning = false
	logger.Debug("速率限制器已停止")
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.bucket:
		return true
	default:
		return false
	}
}

// Wait 等待获取令牌
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.bucket:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// refillTokens 补充令牌
func (rl *RateLimiter) refillTokens() {
	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.bucket <- struct{}{}:
				// 成功添加令牌
			default:
				// 令牌桶已满
			}
		case <-rl.ctx.Done():
			return
		}
	}
} 