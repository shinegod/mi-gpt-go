package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log        *logrus.Logger
	logBuffer  *LogBuffer
	bufferSize = 1000 // 最多保存1000条日志
)

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

// LogBuffer 内存日志缓冲区
type LogBuffer struct {
	entries []LogEntry
	mutex   sync.RWMutex
	maxSize int
}

// NewLogBuffer 创建新的日志缓冲区
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add 添加日志条目
func (lb *LogBuffer) Add(entry LogEntry) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	lb.entries = append(lb.entries, entry)
	
	// 如果超过最大大小，移除最老的条目
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[1:]
	}
}

// GetLogs 获取日志条目
func (lb *LogBuffer) GetLogs(limit int) []LogEntry {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()
	
	if limit <= 0 || limit > len(lb.entries) {
		// 复制所有日志
		result := make([]LogEntry, len(lb.entries))
		copy(result, lb.entries)
		return result
	}
	
	// 返回最新的limit条日志
	start := len(lb.entries) - limit
	result := make([]LogEntry, limit)
	copy(result, lb.entries[start:])
	return result
}

// Clear 清空日志缓冲区
func (lb *LogBuffer) Clear() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.entries = lb.entries[:0]
}

// Count 获取日志条目数量
func (lb *LogBuffer) Count() int {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()
	return len(lb.entries)
}

// MemoryHook 内存日志钩子
type MemoryHook struct {
	buffer *LogBuffer
}

// NewMemoryHook 创建内存钩子
func NewMemoryHook(buffer *LogBuffer) *MemoryHook {
	return &MemoryHook{buffer: buffer}
}

// Levels 返回支持的日志级别
func (hook *MemoryHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 处理日志条目
func (hook *MemoryHook) Fire(entry *logrus.Entry) error {
	logEntry := LogEntry{
		Time:    entry.Time,
		Level:   entry.Level.String(),
		Message: entry.Message,
	}
	
	hook.buffer.Add(logEntry)
	return nil
}

// Init 初始化日志
func Init() {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)
	
	// 初始化日志缓冲区
	logBuffer = NewLogBuffer(bufferSize)
	
	// 添加内存钩子
	memoryHook := NewMemoryHook(logBuffer)
	log.AddHook(memoryHook)
}

// SetLevel 设置日志级别
func SetLevel(level string) error {
	var logLevel logrus.Level
	
	switch level {
	case "debug", "DEBUG":
		logLevel = logrus.DebugLevel
	case "info", "INFO":
		logLevel = logrus.InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		logLevel = logrus.WarnLevel
	case "error", "ERROR":
		logLevel = logrus.ErrorLevel
	case "fatal", "FATAL":
		logLevel = logrus.FatalLevel
	case "panic", "PANIC":
		logLevel = logrus.PanicLevel
	default:
		return fmt.Errorf("未知的日志级别: %s", level)
	}
	
	log.SetLevel(logLevel)
	Infof("日志级别已设置为: %s", logLevel.String())
	return nil
}

// GetLevel 获取当前日志级别
func GetLevel() string {
	return log.GetLevel().String()
}

// GetLogs 获取内存中的日志
func GetLogs(limit int) []LogEntry {
	if logBuffer == nil {
		return []LogEntry{}
	}
	return logBuffer.GetLogs(limit)
}

// ClearLogs 清空内存中的日志
func ClearLogs() {
	if logBuffer != nil {
		logBuffer.Clear()
		Info("日志缓冲区已清空")
	}
}

// GetLogCount 获取日志数量
func GetLogCount() int {
	if logBuffer == nil {
		return 0
	}
	return logBuffer.Count()
}

// Info 输出信息日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 格式化输出信息日志
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 格式化输出警告日志
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error 输出错误日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 格式化输出错误日志
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Debug 输出调试日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 格式化输出调试日志
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Fatal 输出致命错误日志
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 格式化输出致命错误日志
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// ShowBanner 显示启动横幅
func ShowBanner() {
	banner := `
    __  ____   __________  ______   ________
   /  |/  (_) / ____/ __ /_  __/  / ____/ __ \
  / /|_/ / / / / __/ /_/ // /    / / __/ / / /
 / /  / / / / /_/ / ____// /    / /_/ / /_/ /
/_/  /_/_/  \____//_/   /_/     \____/\____/

MiGPT Go 版本 - 将小爱音箱接入 ChatGPT
`
	fmt.Print(banner)
} 