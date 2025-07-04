package utils

import (
	"context"
	"fmt"
	"mi-gpt-go/pkg/logger"
	"runtime"
	"time"
)

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
	MaxRetries    int           // 最大重试次数
	InitialDelay  time.Duration // 初始延迟
	MaxDelay      time.Duration // 最大延迟
	BackoffFactor float64       // 退避因子
	EnablePanic   bool          // 是否启用panic恢复
}

// DefaultRecoveryConfig 默认恢复配置
var DefaultRecoveryConfig = RecoveryConfig{
	MaxRetries:    3,
	InitialDelay:  1 * time.Second,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
	EnablePanic:   true,
}

// RecoveryManager 错误恢复管理器
type RecoveryManager struct {
	config RecoveryConfig
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(config RecoveryConfig) *RecoveryManager {
	return &RecoveryManager{
		config: config,
	}
}

// WithRecover 使用恢复机制执行函数
func (rm *RecoveryManager) WithRecover(ctx context.Context, name string, fn func() error) error {
	var lastError error
	
	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// 如果启用了panic恢复，则包装函数
		if rm.config.EnablePanic {
			func() {
				defer func() {
					if r := recover(); r != nil {
						// 获取调用栈
						buf := make([]byte, 4096)
						n := runtime.Stack(buf, false)
						stack := string(buf[:n])
						
						lastError = fmt.Errorf("panic recovered in %s: %v\nstack:\n%s", name, r, stack)
						logger.Errorf("发生panic: %v", lastError)
					}
				}()
				
				lastError = fn()
			}()
		} else {
			lastError = fn()
		}
		
		// 如果成功或者上下文被取消，直接返回
		if lastError == nil || ctx.Err() != nil {
			return lastError
		}
		
		// 记录重试日志
		if attempt < rm.config.MaxRetries {
			delay := rm.calculateDelay(attempt)
			logger.Warnf("%s 执行失败 (尝试 %d/%d): %v，%v后重试", 
				name, attempt+1, rm.config.MaxRetries+1, lastError, delay)
			
			// 等待重试
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}
	
	return fmt.Errorf("%s 重试%d次后仍失败: %v", name, rm.config.MaxRetries, lastError)
}

// calculateDelay 计算延迟时间
func (rm *RecoveryManager) calculateDelay(attempt int) time.Duration {
	delay := float64(rm.config.InitialDelay) * pow(rm.config.BackoffFactor, float64(attempt))
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}
	return time.Duration(delay)
}

// pow 计算幂次方
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// SafeGo 安全启动goroutine
func SafeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 获取调用栈
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stack := string(buf[:n])
				
				logger.Errorf("Goroutine %s panic: %v\nstack:\n%s", name, r, stack)
			}
		}()
		
		fn()
	}()
}

// SafeGoWithContext 带上下文的安全goroutine
func SafeGoWithContext(ctx context.Context, name string, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 获取调用栈
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stack := string(buf[:n])
				
				logger.Errorf("Goroutine %s panic: %v\nstack:\n%s", name, r, stack)
			}
		}()
		
		fn(ctx)
	}()
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	maxFailures int       // 最大失败次数
	resetTime   time.Duration // 重置时间
	failures    int       // 当前失败次数
	lastFailure time.Time // 最后失败时间
	state       string    // 状态: "closed", "open", "half-open"
}

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(maxFailures int, resetTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		resetTime:   resetTime,
		state:       "closed",
	}
}

// Execute 执行函数
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// 检查是否可以执行
	if !cb.canExecute() {
		return fmt.Errorf("断路器已打开，拒绝执行")
	}
	
	err := fn()
	
	// 更新状态
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	
	return err
}

// canExecute 检查是否可以执行
func (cb *CircuitBreaker) canExecute() bool {
	switch cb.state {
	case "closed":
		return true
	case "open":
		// 检查是否可以半开
		if time.Since(cb.lastFailure) > cb.resetTime {
			cb.state = "half-open"
			return true
		}
		return false
	case "half-open":
		return true
	default:
		return false
	}
}

// onFailure 处理失败
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = "open"
		logger.Warnf("断路器已打开，失败次数: %d", cb.failures)
	}
}

// onSuccess 处理成功
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	cb.state = "closed"
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() string {
	return cb.state
}

// GetFailures 获取失败次数
func (cb *CircuitBreaker) GetFailures() int {
	return cb.failures
} 