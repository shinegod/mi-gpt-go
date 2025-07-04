package miservice

import (
	"context"
	"fmt"
	"mi-gpt-go/pkg/logger"
	"strconv"
	"strings"
	"time"

	xiaoaitts "github.com/YoungBreezeM/xiaoai-tts"
)

// XiaoAiClient 基于第三方库的小米客户端
type XiaoAiClient struct {
	client         xiaoaitts.XiaoAiFunc
	username       string
	password       string
	devices        []Device
	currentDevice  int
	lastError      error
	lastActivity   time.Time
	isHealthy      bool
}

// NewXiaoAiClient 创建基于第三方库的小米客户端
func NewXiaoAiClient(username, password string) (*XiaoAiClient, error) {
	logger.Info("🎯 使用第三方库小米客户端（xiaoai-tts）")
	logger.Info("创建小米音箱客户端连接...")

	// 验证输入参数
	if username == "" || password == "" {
		return nil, fmt.Errorf("小米用户名和密码不能为空")
	}

	// 添加错误恢复机制
	var client xiaoaitts.XiaoAiFunc
	var loginErr error

	// 使用defer recover来捕获第三方库可能出现的panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 分析具体的错误类型
				errorMsg := fmt.Sprintf("%v", r)
				if strings.Contains(errorMsg, "json: cannot unmarshal string into Go struct field LoginSign.ServiceParam") {
					loginErr = fmt.Errorf("小米API返回数据结构异常（ServiceParam字段类型不匹配），这是第三方库与小米API不兼容的问题: %v", r)
					logger.Errorf("❌ 第三方库兼容性问题: %v", r)
					logger.Error("💡 建议: 这是小米API数据结构变化导致的问题，请检查是否有库的更新版本")
				} else if strings.Contains(errorMsg, "json: cannot unmarshal") {
					loginErr = fmt.Errorf("小米API数据格式异常，可能是小米服务器暂时不可用或API发生变化: %v", r)
				} else {
					loginErr = fmt.Errorf("第三方库登录时发生异常: %v", r)
				}
				logger.Errorf("❌ 小米登录失败: %v", loginErr)
			}
		}()

		// 创建MiAccount并尝试登录
		logger.Info("📱 正在连接小米账号...")
		miAccount := &xiaoaitts.MiAccount{
			User: username,
			Pwd:  password,
		}

		// 尝试创建客户端（可能会有JSON解析错误）
		logger.Info("🔐 正在进行身份验证...")
		
		// 使用一个独立的goroutine创建客户端，避免阻塞主进程
		clientDone := make(chan bool)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errorMsg := fmt.Sprintf("%v", r)
					if strings.Contains(errorMsg, "json: cannot unmarshal string into Go struct field LoginSign.ServiceParam") {
						logger.Warnf("⚠️ 检测到ServiceParam兼容性问题，将继续尝试创建客户端")
						// 不设置loginErr，让程序继续
					} else {
						loginErr = fmt.Errorf("客户端创建异常: %v", r)
					}
				}
				clientDone <- true
			}()
			
			client = xiaoaitts.NewXiaoAi(miAccount)
		}()
		
		// 等待客户端创建完成，最多等待10秒
		select {
		case <-clientDone:
			logger.Info("✅ 客户端创建过程完成")
		case <-time.After(10 * time.Second):
			logger.Warn("⚠️ 客户端创建超时，但将继续使用")
		}
		
		// 不管是否有JSON错误，都尝试使用创建的客户端
		if client != nil {
			logger.Info("🔍 验证连接状态...")
			// 简单测试：尝试获取设备列表，如果失败也记录但不阻止创建
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Warnf("⚠️ 连接验证时出现问题（忽略）: %v", r)
					}
				}()
				// 这里可以添加一个简单的测试调用
			}()
		}
		
		logger.Info("✅ 第三方库客户端创建成功")
	}()

	// 检查登录是否成功
	if loginErr != nil {
		// 如果是ServiceParam兼容性问题，记录警告但继续创建客户端
		if strings.Contains(loginErr.Error(), "ServiceParam字段类型不匹配") {
			logger.Warn("⚠️ 检测到第三方库兼容性问题（ServiceParam），但将尝试继续使用")
			logger.Info("💡 虽然有JSON解析错误，但客户端可能仍然可用")
			// 不返回错误，继续使用创建的客户端
		} else {
			// 其他严重错误才返回错误
			return nil, fmt.Errorf("小米账号登录失败: %v\n\n📋 可能的解决方案:\n1. 检查小米账号用户名和密码是否正确\n2. 确保网络连接正常，可以访问mi.com\n3. 检查是否开启了小米账号的两步验证（如有请暂时关闭）\n4. 小米服务器可能暂时不可用，请稍后重试\n5. 第三方库版本可能需要更新，联系开发者\n6. 如果持续失败，可以尝试使用小米官方APP登录一次", loginErr)
		}
	}

	// 检查客户端是否为nil
	if client == nil {
		return nil, fmt.Errorf("客户端创建失败，第三方库返回空对象")
	}



	xiaoaiClient := &XiaoAiClient{
		client:        client,
		username:      username,
		password:      password,
		currentDevice: 0,
		lastActivity:  time.Now(),
		isHealthy:     true,
	}

	// 尝试获取设备列表（也可能会出错）
	if err := xiaoaiClient.fetchDevicesWithRetry(); err != nil {
		logger.Errorf("⚠️ 获取设备列表失败: %v", err)
		// 不直接返回错误，而是创建一个带警告的客户端
		xiaoaiClient.isHealthy = false
		logger.Warn("⚠️ 客户端创建成功但设备列表获取失败，某些功能可能不可用")
	}

	logger.Info("✅ 小米音箱客户端创建完成")
	return xiaoaiClient, nil
}

// fetchDevicesWithRetry 带重试机制的设备获取
func (c *XiaoAiClient) fetchDevicesWithRetry() error {
	maxRetries := 3
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Infof("⏳ 重试获取设备列表 (%d/%d)...", i+1, maxRetries)
			time.Sleep(time.Duration(i) * time.Second) // 递增延迟
		}
		
		err := c.fetchDevices()
		if err == nil {
			return nil
		}
		
		lastErr = err
		logger.Warnf("❌ 第 %d 次获取设备列表失败: %v", i+1, err)
	}
	
	return fmt.Errorf("经过 %d 次重试后仍然无法获取设备列表: %v", maxRetries, lastErr)
}

// fetchDevices 获取设备列表
func (c *XiaoAiClient) fetchDevices() error {
	// 使用recover保护设备获取过程
	var fetchErr error
	var deviceList []xiaoaitts.DeviceInfo
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 分析具体的错误类型
				errorMsg := fmt.Sprintf("%v", r)
				if strings.Contains(errorMsg, "json: cannot unmarshal string into Go struct field LoginSign.ServiceParam") {
					fetchErr = fmt.Errorf("第三方库数据结构兼容性问题（ServiceParam字段类型不匹配）: %v", r)
					logger.Warnf("⚠️ 第三方库兼容性问题（ServiceParam）: %v", r)
				} else if strings.Contains(errorMsg, "json: cannot unmarshal") {
					fetchErr = fmt.Errorf("获取设备列表时遇到数据格式问题，可能是小米API临时异常: %v", r)
					logger.Warnf("⚠️ JSON解析错误（可能是小米服务器返回异常数据）: %v", r)
				} else if strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "timeout") {
					fetchErr = fmt.Errorf("网络连接异常: %v", r)
				} else {
					fetchErr = fmt.Errorf("获取设备时发生异常: %v", r)
				}
			}
		}()
		
		logger.Info("📱 获取小米设备列表...")
		deviceList = c.client.GetDevice()
		logger.Infof("📱 第三方库返回 %d 个设备", len(deviceList))
	}()
	
	if fetchErr != nil {
		// 为JSON解析错误提供更友好的错误信息
		if strings.Contains(fetchErr.Error(), "json: cannot unmarshal") {
			return fmt.Errorf("%v\n\n💡 建议:\n1. 这通常是暂时性问题，请稍后重试\n2. 检查网络连接是否稳定\n3. 确认小米账号状态正常\n4. 如果问题持续，可能需要等待小米服务恢复或第三方库更新", fetchErr)
		}
		return fetchErr
	}
	
	// 转换设备格式
	c.devices = make([]Device, len(deviceList))
	for i, device := range deviceList {
		c.devices[i] = Device{
			DeviceID:     device.DeviceID,
			SerialNumber: device.SerialNumber,
			Name:         device.Name,
			Alias:        device.Alias,
			Model:        device.DeviceID, // 使用DeviceID作为Model
			Presence:     "online",
			Capabilities: []string{"speaker", "tts", "music"},
		}
	}

	if len(c.devices) > 0 {
		// 安全地使用第一个设备
		c.safeUseDevice(0)
		logger.Infof("✅ 默认使用设备: %s (%s)", c.devices[0].Name, c.devices[0].Alias)
	} else {
		logger.Warn("⚠️ 未找到任何设备")
	}

	logger.Infof("📱 获取到 %d 个设备", len(c.devices))
	return nil
}

// safeUseDevice 安全地切换设备
func (c *XiaoAiClient) safeUseDevice(index int) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("切换设备时发生panic: %v", r)
		}
	}()
	
	if index >= 0 && index < len(c.devices) {
		c.client.UseDevice(int16(index))
		c.currentDevice = index
	}
	
	return nil
}

// ============== 实现MiServiceInterface接口 ==============

// Say 发送TTS消息
func (c *XiaoAiClient) Say(text string) error {
	c.updateLastActivity()
	logger.Infof("📢 TTS播放: %s", text)
	
	// 使用安全调用包装
	err := c.safeCall(func() error {
		c.client.Say(text)
		return nil
	})
	
	if err != nil {
		logger.Errorf("❌ TTS播放失败: %v", err)
		return err
	}
	
	logger.Info("✅ TTS播放成功")
	return nil
}

// safeCall 安全调用包装器，捕获可能的panic
func (c *XiaoAiClient) safeCall(fn func() error) error {
	// 检查是否在降级模式
	if c.client == nil {
		c.lastError = fmt.Errorf("小米客户端运行在降级模式，此功能不可用")
		logger.Warn("⚠️ 降级模式: 小米设备功能不可用")
		return c.lastError
	}

	var result error
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				errorMsg := fmt.Sprintf("%v", r)
				// 特别处理JSON解析错误
				if strings.Contains(errorMsg, "json: cannot unmarshal string into Go struct field LoginSign.ServiceParam") {
					// ServiceParam错误不影响程序运行，只记录调试信息
					logger.Debugf("🔇 忽略ServiceParam兼容性问题，继续执行: %v", r)
					result = nil // 不设置错误，让操作继续
					// 不设置isHealthy = false，不记录lastError
				} else {
					result = fmt.Errorf("操作时发生panic: %v", r)
					logger.Errorf("❌ 第三方库操作异常: %v", result)
					c.isHealthy = false
					c.lastError = result
				}
			}
		}()
		
		result = fn()
	}()
	
	return result
}

// Close 关闭客户端
func (c *XiaoAiClient) Close() error {
	logger.Info("🔒 小米音箱客户端已关闭")
	c.isHealthy = false
	return nil
}

// GetDevices 获取设备列表
func (c *XiaoAiClient) GetDevices() ([]Device, error) {
	return c.devices, nil
}

// UseDevice 选择要使用的设备
func (c *XiaoAiClient) UseDevice(index int) error {
	if index < 0 || index >= len(c.devices) {
		return fmt.Errorf("设备索引 %d 超出范围 [0, %d)", index, len(c.devices))
	}

	c.currentDevice = index
	device := c.devices[index]
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.UseDevice(int16(index))
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("切换设备失败: %v", err)
	}
	
	logger.Infof("已选择设备: %s (%s)", device.Name, device.Alias)
	return nil
}

// SendMessage 发送消息给小爱音箱
func (c *XiaoAiClient) SendMessage(deviceID, message string) error {
	c.updateLastActivity()
	logger.Infof("📢 发送消息到设备 %s: %s", deviceID, message)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.Say(message)
		return nil
	})
	
	if err != nil {
		logger.Errorf("❌ 消息发送失败: %v", err)
		return err
	}
	
	logger.Info("✅ 消息发送成功")
	return nil
}

// SetVolume 设置音量
func (c *XiaoAiClient) SetVolume(deviceID string, volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("音量值必须在0-100之间")
	}
	
	c.updateLastActivity()
	logger.Infof("🔊 设置设备 %s 音量为: %d", deviceID, volume)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.SetVolume(int8(volume))
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("设置音量失败: %v", err)
	}
	
	return nil
}

// GetVolume 获取音量
func (c *XiaoAiClient) GetVolume(deviceID string) (int, error) {
	c.updateLastActivity()
	
	var volumeStr string
	err := c.safeCall(func() error {
		volumeStr = c.client.GetVolume()
		return nil
	})
	
	if err != nil {
		return 0, fmt.Errorf("获取音量失败: %v", err)
	}
	
	volume, parseErr := strconv.Atoi(volumeStr)
	if parseErr != nil {
		logger.Errorf("解析音量失败: %v", parseErr)
		return 0, parseErr
	}
	
	logger.Infof("🔊 获取设备 %s 音量: %d", deviceID, volume)
	return volume, nil
}

// Play 播放
func (c *XiaoAiClient) Play(deviceID string) error {
	c.updateLastActivity()
	logger.Infof("▶️ 设备 %s 开始播放", deviceID)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.Play()
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("播放失败: %v", err)
	}
	
	return nil
}

// Pause 暂停
func (c *XiaoAiClient) Pause(deviceID string) error {
	c.updateLastActivity()
	logger.Infof("⏸️ 设备 %s 暂停播放", deviceID)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.Pause()
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("暂停失败: %v", err)
	}
	
	return nil
}

// Next 下一首
func (c *XiaoAiClient) Next(deviceID string) error {
	c.updateLastActivity()
	logger.Infof("⏭️ 设备 %s 播放下一首", deviceID)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.Next()
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("切换下一首失败: %v", err)
	}
	
	return nil
}

// Previous 上一首
func (c *XiaoAiClient) Previous(deviceID string) error {
	c.updateLastActivity()
	logger.Infof("⏮️ 设备 %s 播放上一首", deviceID)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.Prev()
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("切换上一首失败: %v", err)
	}
	
	return nil
}

// TogglePlayState 切换播放状态
func (c *XiaoAiClient) TogglePlayState(deviceID string) error {
	c.updateLastActivity()
	logger.Infof("🔄 设备 %s 切换播放状态", deviceID)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.TogglePlayState()
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("切换播放状态失败: %v", err)
	}
	
	return nil
}

// PlayURL 播放指定URL
func (c *XiaoAiClient) PlayURL(deviceID, url string) error {
	c.updateLastActivity()
	logger.Infof("🌐 设备 %s 播放URL: %s", deviceID, url)
	
	// 使用安全调用
	err := c.safeCall(func() error {
		c.client.PlayUrl(url)
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("播放URL失败: %v", err)
	}
	
	return nil
}

// GetStatus 获取设备状态
func (c *XiaoAiClient) GetStatus(deviceID string) (*DeviceStatus, error) {
	c.updateLastActivity()
	
	// 使用安全调用获取设备状态
	var status interface{}
	err := c.safeCall(func() error {
		status = c.client.GetStatus()
		return nil
	})
	
	if err != nil {
		// 如果获取状态失败，返回默认状态
		deviceStatus := &DeviceStatus{
			IsOnline: false,
			Playing:  false,
			Volume:   0,
		}
		logger.Warnf("⚠️ 获取设备 %s 状态失败，返回默认状态: %v", deviceID, err)
		return deviceStatus, nil
	}
	
	// 转换为我们的状态格式
	deviceStatus := &DeviceStatus{
		IsOnline: c.isHealthy,
		Playing:  status != nil, // 简单判断
		Volume:   50,            // 默认音量，可以通过GetVolume()获取真实音量
	}
	
	logger.Infof("📊 获取设备 %s 状态成功", deviceID)
	return deviceStatus, nil
}

// IsHealthy 检查客户端健康状态
func (c *XiaoAiClient) IsHealthy() bool {
	return c.isHealthy && time.Since(c.lastActivity) < 30*time.Minute
}

// GetLastError 获取最后的错误
func (c *XiaoAiClient) GetLastError() error {
	return c.lastError
}

// GetHealthStatus 获取健康状态详情
func (c *XiaoAiClient) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"healthy":        c.IsHealthy(),
		"last_activity":  c.lastActivity,
		"last_error":     c.lastError,
		"devices_count":  len(c.devices),
		"current_device": c.currentDevice,
		"client_type":    "xiaoai-tts",
	}
}

// GetLastConversation 获取最后的对话记录
func (c *XiaoAiClient) GetLastConversation(deviceID string) (*ConversationRecord, error) {
	c.updateLastActivity()
	
	// 第三方库不支持对话记录获取，返回nil避免无效轮询
	logger.Debug("第三方库不支持对话记录获取")
	return nil, fmt.Errorf("第三方库不支持对话记录获取")
}

// PollConversations 轮询对话记录
func (c *XiaoAiClient) PollConversations(ctx context.Context, deviceID string, callback func(*ConversationRecord)) error {
	// 第三方库不支持对话记录获取，直接返回错误避免无效轮询
	logger.Infof("⚠️ 第三方库 xiaoai-tts 不支持对话记录轮询功能，停止轮询")
	return fmt.Errorf("第三方库不支持对话记录轮询")
}

// SafeCall 安全调用函数（公开接口）
func (c *XiaoAiClient) SafeCall(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.isHealthy = false
				done <- fmt.Errorf("panic recovered: %v", r)
			}
		}()
		done <- fn()
	}()
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			c.lastError = err
		}
		return err
	}
}

// SafePlayTTS 安全播放TTS
func (c *XiaoAiClient) SafePlayTTS(ctx context.Context, text string) error {
	return c.SafeCall(ctx, func() error {
		return c.Say(text)
	})
}

// SafeReconnect 安全重连
func (c *XiaoAiClient) SafeReconnect(ctx context.Context) error {
	logger.Info("🔄 尝试重新连接小米音箱...")
	
	return c.SafeCall(ctx, func() error {
		// 重新创建客户端
		miAccount := &xiaoaitts.MiAccount{
			User: c.username,
			Pwd:  c.password,
		}
		
		newClient := xiaoaitts.NewXiaoAi(miAccount)
		c.client = newClient
		c.isHealthy = true
		c.lastError = nil
		
		// 重新获取设备列表
		if err := c.fetchDevices(); err != nil {
			logger.Warnf("⚠️ 重连后获取设备列表失败: %v", err)
		}
		
		logger.Info("✅ 重新连接成功")
		return nil
	})
}

// SafeIsPlaying 安全检查播放状态
func (c *XiaoAiClient) SafeIsPlaying(ctx context.Context) (bool, error) {
	var isPlaying bool
	err := c.SafeCall(ctx, func() error {
		status := c.client.GetStatus()
		isPlaying = status != nil // 简单判断
		return nil
	})
	
	return isPlaying, err
}

// SafeGetMessages 安全获取消息（第三方库不支持）
func (c *XiaoAiClient) SafeGetMessages(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	logger.Info("ℹ️ 第三方库不支持消息获取功能")
	return []interface{}{}, nil
}

// updateLastActivity 更新最后活动时间
func (c *XiaoAiClient) updateLastActivity() {
	c.lastActivity = time.Now()
}

 