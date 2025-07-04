package main

import (
	"context"
	"mi-gpt-go/internal/config"
	"mi-gpt-go/internal/database"
	configService "mi-gpt-go/internal/services/config"
	"mi-gpt-go/internal/services/speaker"
	"mi-gpt-go/internal/web"
	"mi-gpt-go/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化日志
	logger.Init()
	logger.Info("=== MiGPT Go 增强版 (支持多线程处理 + Web管理面板) ===")
	logger.Info("项目地址: https://github.com/shinegod/mi-gpt-go")
	logger.Info("基于原版 MiGPT 项目: https://github.com/idootop/mi-gpt")

	// 初始化数据库（先初始化数据库，后续需要用来存储配置）
	_, err := database.Init("app.db") // 使用固定的数据库路径
	if err != nil {
		logger.Error("初始化数据库失败:", err)
		os.Exit(1)
	}
	logger.Info("数据库初始化成功")

	// 创建数据库配置服务
	dbConfigService := configService.NewDBConfigService()
	
	// 设置全局配置服务
	config.SetDBConfigService(dbConfigService)

	// 从数据库加载配置
	cfg, err := config.LoadWithDefaults()
	if err != nil {
		logger.Error("加载配置失败:", err)
		os.Exit(1)
	}

	// 打印配置信息
	logger.Infof("当前AI服务提供商: %s", cfg.OpenAI.Provider)
	if cfg.Speaker.EnableConcurrent {
		logger.Infof("并发处理已启用 - 工作协程: %d, 队列大小: %d", 
			cfg.Speaker.WorkerCount, cfg.Speaker.QueueSize)
	} else {
		logger.Info("使用串行处理模式")
	}

	// 创建增强版AI音箱服务（仅在配置完整时启动）
	var aiSpeaker *speaker.EnhancedAISpeaker
	if cfg.IsConfigured() {
		var err error
		aiSpeaker, err = speaker.NewEnhancedAISpeaker(cfg)
		if err != nil {
			logger.Warn("创建AI音箱服务失败，将在前端配置完成后重新创建:", err)
		} else {
			logger.Info("AI音箱服务初始化成功")
		}
	} else {
		logger.Warn("配置不完整，AI音箱服务将在前端配置完成后启动")
	}

	// 启动Web管理面板（传入数据库配置服务）
	webServer := web.NewWebServer(cfg, aiSpeaker, dbConfigService)
	if err := webServer.Start(8080); err != nil {
		logger.Error("Web服务器启动失败:", err)
		// Web服务器启动失败不应该导致整个程序退出
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动AI音箱服务（如果已创建）
	if aiSpeaker != nil {
		logger.Info("正在启动AI音箱服务...")
		if err := aiSpeaker.Start(); err != nil {
			logger.Error("启动AI音箱服务失败:", err)
		} else {
			// 等待服务完全启动
			time.Sleep(2 * time.Second)
			
			// 输出服务状态
			status := aiSpeaker.GetStatus()
			logger.Infof("AI音箱服务状态: %+v", status)
		}
	}

	logger.Info("🎉 MiGPT Go 增强版启动成功!")
	logger.Info("🌐 Web管理面板: http://0.0.0.0:8080")
	logger.Info("🌍 外网访问地址: http://[你的IP]:8080")
	logger.Info("💾 配置数据存储在数据库中，无需配置文件")
	if aiSpeaker != nil {
		logger.Info("📱 现在可以对小爱音箱说话了")
		if cfg.Speaker.EnableConcurrent {
			logger.Info("⚡ 并发处理模式已启用，支持多任务并行处理")
		}
	} else {
		logger.Info("⚙️  请在Web管理面板中完成配置后启动AI音箱服务")
	}
	logger.Info("🔧 按 Ctrl+C 可安全退出程序")

	// 启动状态监控
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if aiSpeaker != nil && aiSpeaker.IsRunning() {
					status := aiSpeaker.GetStatus()
					logger.Debugf("服务运行状态: %+v", status)
				}
			}
		}
	}()

	// 等待退出信号
	select {
	case sig := <-sigChan:
		logger.Infof("收到信号 %v，正在优雅关闭...", sig)
	case <-ctx.Done():
		logger.Info("上下文取消，正在关闭...")
	}

	// 取消上下文
	cancel()

	// 停止Web服务器
	logger.Info("正在停止Web服务器...")
	if err := webServer.Stop(); err != nil {
		logger.Error("停止Web服务器失败:", err)
	} else {
		logger.Info("Web服务器已停止")
	}

	// 停止AI音箱服务
	if aiSpeaker != nil {
		logger.Info("正在停止AI音箱服务...")
		if err := aiSpeaker.Stop(); err != nil {
			logger.Error("停止AI音箱服务失败:", err)
		} else {
			logger.Info("AI音箱服务已停止")
		}
	}

	// 等待所有goroutine结束
	time.Sleep(2 * time.Second)

	logger.Info("🎯 MiGPT Go 增强版已安全退出")
	logger.Info("感谢使用! 再见! 👋")
} 