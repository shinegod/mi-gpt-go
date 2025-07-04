package miservice

import "context"

// Device 设备信息
type Device struct {
	DeviceID        string   `json:"deviceId"`
	SerialNumber    string   `json:"serialNumber"`
	Name            string   `json:"name"`
	Alias           string   `json:"alias"`
	Model           string   `json:"model"`
	HardwareVersion string   `json:"hardwareVersion"`
	SoftwareVersion string   `json:"softwareVersion"`
	MacAddress      string   `json:"macAddress"`
	MasterFlag      int      `json:"masterFlag"`
	Presence        string   `json:"presence"`
	Capabilities    []string `json:"capabilities"`
}

// ConversationRecord 对话记录
type ConversationRecord struct {
	Query string `json:"query"`
	Time  int64  `json:"time"`
}

// DeviceStatus 设备状态
type DeviceStatus struct {
	IsOnline bool `json:"isOnline"`
	Volume   int  `json:"volume"`
	Playing  bool `json:"playing"`
}

// ActionCommand 设备操作命令
type ActionCommand []int

// PropertyCommand 设备属性查询命令
type PropertyCommand []int

// QueryMessage 查询消息
type QueryMessage struct {
	Text      string `json:"text"`      // 消息文本
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// MiServiceInterface 小米服务通用接口
type MiServiceInterface interface {
	// 基础功能
	Say(text string) error
	Close() error
	
	// 设备管理
	GetDevices() ([]Device, error)
	UseDevice(index int) error
	
	// 音频控制
	SetVolume(deviceID string, volume int) error
	GetVolume(deviceID string) (int, error)
	Play(deviceID string) error
	Pause(deviceID string) error
	Next(deviceID string) error
	Previous(deviceID string) error
	TogglePlayState(deviceID string) error
	PlayURL(deviceID, url string) error
	
	// 状态查询
	GetStatus(deviceID string) (*DeviceStatus, error)
	IsHealthy() bool
	GetLastError() error
	GetHealthStatus() map[string]interface{}
	
	// 对话功能
	GetLastConversation(deviceID string) (*ConversationRecord, error)
	PollConversations(ctx context.Context, deviceID string, callback func(*ConversationRecord)) error
	
	// 安全操作
	SafeCall(ctx context.Context, fn func() error) error
	SafePlayTTS(ctx context.Context, text string) error
	SafeReconnect(ctx context.Context) error
	SafeIsPlaying(ctx context.Context) (bool, error)
	SafeGetMessages(ctx context.Context, params map[string]interface{}) ([]interface{}, error)
} 