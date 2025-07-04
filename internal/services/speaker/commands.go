package speaker

import (
	"context"
	"fmt"
	"math/rand"
	"mi-gpt-go/internal/services/miservice"
	"mi-gpt-go/pkg/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CommandHandler 命令处理器接口
type CommandHandler interface {
	GetName() string
	GetDescription() string
	GetPatterns() []string
	Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error)
}

// VolumeCommand 音量控制命令
type VolumeCommand struct{}

func (v *VolumeCommand) GetName() string        { return "音量控制" }
func (v *VolumeCommand) GetDescription() string { return "调节音箱音量" }
func (v *VolumeCommand) GetPatterns() []string {
	return []string{
		`音量调到(\d+)`,
		`音量设为(\d+)`,
		`声音调到(\d+)`,
		`调大声音`,
		`调小声音`,
		`音量大一点`,
		`音量小一点`,
		`静音`,
		`取消静音`,
	}
}

func (v *VolumeCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	// 设置具体音量
	if matched, _ := regexp.MatchString(`音量调到(\d+)|音量设为(\d+)|声音调到(\d+)`, text); matched {
		re := regexp.MustCompile(`(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			volume, err := strconv.Atoi(matches[1])
			if err != nil || volume < 0 || volume > 100 {
				return SpeakerAnswer{Text: "音量数值无效，请输入0到100之间的数字"}, nil
			}
			return SpeakerAnswer{Text: fmt.Sprintf("好的，已将音量调至%d", volume)}, nil
		}
	}
	
	// 调大音量
	if matched, _ := regexp.MatchString(`调大声音|音量大一点`, text); matched {
		return SpeakerAnswer{Text: "好的，已调大音量"}, nil
	}
	
	// 调小音量
	if matched, _ := regexp.MatchString(`调小声音|音量小一点`, text); matched {
		return SpeakerAnswer{Text: "好的，已调小音量"}, nil
	}
	
	// 静音
	if matched, _ := regexp.MatchString(`静音`, text); matched {
		return SpeakerAnswer{Text: "已静音"}, nil
	}
	
	// 取消静音
	if matched, _ := regexp.MatchString(`取消静音`, text); matched {
		return SpeakerAnswer{Text: "已取消静音"}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// TimeCommand 时间查询命令
type TimeCommand struct{}

func (t *TimeCommand) GetName() string        { return "时间查询" }
func (t *TimeCommand) GetDescription() string { return "查询当前时间和日期" }
func (t *TimeCommand) GetPatterns() []string {
	return []string{
		`现在几点了`,
		`几点了`,
		`现在什么时间`,
		`今天是几号`,
		`今天星期几`,
		`现在是几月几号`,
		`告诉我现在的时间`,
	}
}

func (t *TimeCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	now := time.Now()
	text := msg.Text
	
	if matched, _ := regexp.MatchString(`现在几点了|几点了|现在什么时间|告诉我现在的时间`, text); matched {
		timeStr := now.Format("15:04")
		hour := now.Hour()
		var period string
		if hour < 6 {
			period = "凌晨"
		} else if hour < 12 {
			period = "上午"
		} else if hour < 18 {
			period = "下午"
		} else {
			period = "晚上"
		}
		return SpeakerAnswer{Text: fmt.Sprintf("现在是%s%s", period, timeStr)}, nil
	}
	
	if matched, _ := regexp.MatchString(`今天是几号|现在是几月几号`, text); matched {
		dateStr := now.Format("1月2日")
		return SpeakerAnswer{Text: fmt.Sprintf("今天是%s", dateStr)}, nil
	}
	
	if matched, _ := regexp.MatchString(`今天星期几`, text); matched {
		weekdays := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
		weekday := weekdays[now.Weekday()]
		return SpeakerAnswer{Text: fmt.Sprintf("今天是%s", weekday)}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// WeatherCommand 天气查询命令
type WeatherCommand struct{}

func (w *WeatherCommand) GetName() string        { return "天气查询" }
func (w *WeatherCommand) GetDescription() string { return "查询天气信息" }
func (w *WeatherCommand) GetPatterns() []string {
	return []string{
		`今天天气怎么样`,
		`天气如何`,
		`今天的天气`,
		`明天天气`,
		`后天天气`,
		`(.+)的天气`,
	}
}

func (w *WeatherCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	// 这里应该调用实际的天气API，现在返回模拟数据
	responses := []string{
		"今天天气晴朗，温度25度，适合出行",
		"今天多云，温度22度，建议带件外套",
		"今天有小雨，温度18度，记得带伞",
		"今天阴天，温度20度，空气湿润",
	}
	
	response := responses[rand.Intn(len(responses))]
	return SpeakerAnswer{Text: response}, nil
}

// MusicCommand 音乐控制命令
type MusicCommand struct{}

func (m *MusicCommand) GetName() string        { return "音乐控制" }
func (m *MusicCommand) GetDescription() string { return "控制音乐播放" }
func (m *MusicCommand) GetPatterns() []string {
	return []string{
		`播放音乐`,
		`放首歌`,
		`来首歌`,
		`播放(.+)`,
		`我想听(.+)`,
		`暂停`,
		`停止播放`,
		`继续播放`,
		`下一首`,
		`上一首`,
		`切歌`,
	}
}

func (m *MusicCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	if matched, _ := regexp.MatchString(`播放音乐|放首歌|来首歌`, text); matched {
		return SpeakerAnswer{Text: "好的，为您播放音乐"}, nil
	}
	
	if matched, _ := regexp.MatchString(`播放(.+)|我想听(.+)`, text); matched {
		re := regexp.MustCompile(`播放(.+)|我想听(.+)`)
		matches := re.FindStringSubmatch(text)
		var song string
		if len(matches) > 1 && matches[1] != "" {
			song = matches[1]
		} else if len(matches) > 2 && matches[2] != "" {
			song = matches[2]
		}
		if song != "" {
			return SpeakerAnswer{Text: fmt.Sprintf("好的，为您播放%s", song)}, nil
		}
	}
	
	if matched, _ := regexp.MatchString(`暂停|停止播放`, text); matched {
		return SpeakerAnswer{Text: "已暂停播放"}, nil
	}
	
	if matched, _ := regexp.MatchString(`继续播放`, text); matched {
		return SpeakerAnswer{Text: "继续播放"}, nil
	}
	
	if matched, _ := regexp.MatchString(`下一首|切歌`, text); matched {
		return SpeakerAnswer{Text: "好的，下一首"}, nil
	}
	
	if matched, _ := regexp.MatchString(`上一首`, text); matched {
		return SpeakerAnswer{Text: "好的，上一首"}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// DeviceCommand 设备控制命令
type DeviceCommand struct{}

func (d *DeviceCommand) GetName() string        { return "设备控制" }
func (d *DeviceCommand) GetDescription() string { return "控制智能设备" }
func (d *DeviceCommand) GetPatterns() []string {
	return []string{
		`打开(.+)`,
		`关闭(.+)`,
		`开灯`,
		`关灯`,
		`调亮一点`,
		`调暗一点`,
		`空调温度调到(\d+)度`,
		`打开空调`,
		`关闭空调`,
		`设备状态`,
		`连接的设备`,
	}
}

func (d *DeviceCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	if matched, _ := regexp.MatchString(`打开(.+)|关闭(.+)`, text); matched {
		re := regexp.MustCompile(`(打开|关闭)(.+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			action := matches[1]
			device := strings.TrimSpace(matches[2])
			return SpeakerAnswer{Text: fmt.Sprintf("好的，已%s%s", action, device)}, nil
		}
	}
	
	if matched, _ := regexp.MatchString(`开灯`, text); matched {
		return SpeakerAnswer{Text: "好的，已打开灯光"}, nil
	}
	
	if matched, _ := regexp.MatchString(`关灯`, text); matched {
		return SpeakerAnswer{Text: "好的，已关闭灯光"}, nil
	}
	
	if matched, _ := regexp.MatchString(`调亮一点`, text); matched {
		return SpeakerAnswer{Text: "好的，已调亮灯光"}, nil
	}
	
	if matched, _ := regexp.MatchString(`调暗一点`, text); matched {
		return SpeakerAnswer{Text: "好的，已调暗灯光"}, nil
	}
	
	if matched, _ := regexp.MatchString(`空调温度调到(\d+)度`, text); matched {
		re := regexp.MustCompile(`(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			temp := matches[1]
			return SpeakerAnswer{Text: fmt.Sprintf("好的，已将空调温度调至%s度", temp)}, nil
		}
	}
	
	if matched, _ := regexp.MatchString(`设备状态|连接的设备`, text); matched {
		return SpeakerAnswer{Text: "当前连接的设备有：小爱音箱、智能灯泡、空调"}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// CalculatorCommand 计算器命令
type CalculatorCommand struct{}

func (c *CalculatorCommand) GetName() string        { return "计算器" }
func (c *CalculatorCommand) GetDescription() string { return "执行简单的数学计算" }
func (c *CalculatorCommand) GetPatterns() []string {
	return []string{
		`(\d+)\s*[\+加]\s*(\d+)`,
		`(\d+)\s*[\-减]\s*(\d+)`,
		`(\d+)\s*[\*×乘]\s*(\d+)`,
		`(\d+)\s*[\/÷除]\s*(\d+)`,
		`计算(.+)`,
	}
}

func (c *CalculatorCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	// 加法
	if matched, _ := regexp.MatchString(`(\d+)\s*[\+加]\s*(\d+)`, text); matched {
		re := regexp.MustCompile(`(\d+)\s*[\+加]\s*(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			a, _ := strconv.Atoi(matches[1])
			b, _ := strconv.Atoi(matches[2])
			result := a + b
			return SpeakerAnswer{Text: fmt.Sprintf("%d加%d等于%d", a, b, result)}, nil
		}
	}
	
	// 减法
	if matched, _ := regexp.MatchString(`(\d+)\s*[\-减]\s*(\d+)`, text); matched {
		re := regexp.MustCompile(`(\d+)\s*[\-减]\s*(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			a, _ := strconv.Atoi(matches[1])
			b, _ := strconv.Atoi(matches[2])
			result := a - b
			return SpeakerAnswer{Text: fmt.Sprintf("%d减%d等于%d", a, b, result)}, nil
		}
	}
	
	// 乘法
	if matched, _ := regexp.MatchString(`(\d+)\s*[\*×乘]\s*(\d+)`, text); matched {
		re := regexp.MustCompile(`(\d+)\s*[\*×乘]\s*(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			a, _ := strconv.Atoi(matches[1])
			b, _ := strconv.Atoi(matches[2])
			result := a * b
			return SpeakerAnswer{Text: fmt.Sprintf("%d乘以%d等于%d", a, b, result)}, nil
		}
	}
	
	// 除法
	if matched, _ := regexp.MatchString(`(\d+)\s*[\/÷除]\s*(\d+)`, text); matched {
		re := regexp.MustCompile(`(\d+)\s*[\/÷除]\s*(\d+)`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			a, _ := strconv.Atoi(matches[1])
			b, _ := strconv.Atoi(matches[2])
			if b == 0 {
				return SpeakerAnswer{Text: "除数不能为零"}, nil
			}
			result := float64(a) / float64(b)
			return SpeakerAnswer{Text: fmt.Sprintf("%d除以%d等于%.2f", a, b, result)}, nil
		}
	}
	
	return SpeakerAnswer{}, nil
}

// TimerCommand 定时器命令
type TimerCommand struct{}

func (t *TimerCommand) GetName() string        { return "定时器" }
func (t *TimerCommand) GetDescription() string { return "设置定时器和闹钟" }
func (t *TimerCommand) GetPatterns() []string {
	return []string{
		`设置(\d+)分钟定时器`,
		`(\d+)分钟后提醒我`,
		`设置闹钟(\d+)点(\d+)分`,
		`明天(\d+)点叫我`,
		`取消定时器`,
		`取消闹钟`,
	}
}

func (t *TimerCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	if matched, _ := regexp.MatchString(`设置(\d+)分钟定时器|(\d+)分钟后提醒我`, text); matched {
		re := regexp.MustCompile(`(\d+)分钟`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			minutes := matches[1]
			return SpeakerAnswer{Text: fmt.Sprintf("好的，已设置%s分钟定时器", minutes)}, nil
		}
	}
	
	if matched, _ := regexp.MatchString(`设置闹钟(\d+)点(\d+)分|明天(\d+)点叫我`, text); matched {
		re := regexp.MustCompile(`(\d+)点`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			hour := matches[1]
			return SpeakerAnswer{Text: fmt.Sprintf("好的，已设置明天%s点的闹钟", hour)}, nil
		}
	}
	
	if matched, _ := regexp.MatchString(`取消定时器`, text); matched {
		return SpeakerAnswer{Text: "好的，已取消定时器"}, nil
	}
	
	if matched, _ := regexp.MatchString(`取消闹钟`, text); matched {
		return SpeakerAnswer{Text: "好的，已取消闹钟"}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// FunCommand 娱乐命令
type FunCommand struct{}

func (f *FunCommand) GetName() string        { return "娱乐功能" }
func (f *FunCommand) GetDescription() string { return "提供各种娱乐功能" }
func (f *FunCommand) GetPatterns() []string {
	return []string{
		`讲个笑话`,
		`说个故事`,
		`猜谜语`,
		`成语接龙`,
		`古诗词`,
		`绕口令`,
		`抛硬币`,
		`掷骰子`,
		`随机数`,
	}
}

func (f *FunCommand) Handle(ctx context.Context, msg miservice.QueryMessage, speaker *AISpeaker) (SpeakerAnswer, error) {
	text := msg.Text
	
	if matched, _ := regexp.MatchString(`讲个笑话`, text); matched {
		jokes := []string{
			"为什么程序员喜欢黑暗？因为光会产生bug！",
			"小明问老师：'为什么叫WiFi？'老师说：'因为没有电线啊！'",
			"医生对病人说：'你需要戒烟戒酒。'病人说：'医生，我要重新做人！'医生：'不用那么极端。'",
		}
		joke := jokes[rand.Intn(len(jokes))]
		return SpeakerAnswer{Text: joke}, nil
	}
	
	if matched, _ := regexp.MatchString(`抛硬币`, text); matched {
		results := []string{"正面", "反面"}
		result := results[rand.Intn(len(results))]
		return SpeakerAnswer{Text: fmt.Sprintf("硬币结果是：%s", result)}, nil
	}
	
	if matched, _ := regexp.MatchString(`掷骰子`, text); matched {
		result := rand.Intn(6) + 1
		return SpeakerAnswer{Text: fmt.Sprintf("骰子点数是：%d", result)}, nil
	}
	
	if matched, _ := regexp.MatchString(`随机数`, text); matched {
		result := rand.Intn(100) + 1
		return SpeakerAnswer{Text: fmt.Sprintf("随机数是：%d", result)}, nil
	}
	
	if matched, _ := regexp.MatchString(`古诗词`, text); matched {
		poems := []string{
			"春眠不觉晓，处处闻啼鸟。夜来风雨声，花落知多少。",
			"床前明月光，疑是地上霜。举头望明月，低头思故乡。",
			"白日依山尽，黄河入海流。欲穷千里目，更上一层楼。",
		}
		poem := poems[rand.Intn(len(poems))]
		return SpeakerAnswer{Text: poem}, nil
	}
	
	return SpeakerAnswer{}, nil
}

// CommandRegistry 命令注册表
type CommandRegistry struct {
	handlers map[string]CommandHandler
}

// NewCommandRegistry 创建命令注册表
func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		handlers: make(map[string]CommandHandler),
	}
	
	// 注册所有命令
	registry.Register(&VolumeCommand{})
	registry.Register(&TimeCommand{})
	registry.Register(&WeatherCommand{})
	registry.Register(&MusicCommand{})
	registry.Register(&DeviceCommand{})
	registry.Register(&CalculatorCommand{})
	registry.Register(&TimerCommand{})
	registry.Register(&FunCommand{})
	
	return registry
}

// Register 注册命令处理器
func (cr *CommandRegistry) Register(handler CommandHandler) {
	cr.handlers[handler.GetName()] = handler
	logger.Debugf("注册命令处理器: %s", handler.GetName())
}

// FindHandler 查找匹配的命令处理器
func (cr *CommandRegistry) FindHandler(text string) CommandHandler {
	for _, handler := range cr.handlers {
		for _, pattern := range handler.GetPatterns() {
			if matched, _ := regexp.MatchString(pattern, text); matched {
				return handler
			}
		}
	}
	return nil
}

// GetAllHandlers 获取所有命令处理器
func (cr *CommandRegistry) GetAllHandlers() map[string]CommandHandler {
	return cr.handlers
}

// GetCommands 获取命令列表
func (cr *CommandRegistry) GetCommands() []string {
	var commands []string
	for name, handler := range cr.handlers {
		commands = append(commands, fmt.Sprintf("%s: %s", name, handler.GetDescription()))
	}
	return commands
} 