import { defineStore } from 'pinia'
import { configAPI, systemAPI, speakerAPI, concurrentAPI } from '../api'

// 配置管理Store
export const useConfigStore = defineStore('config', {
  state: () => ({
    config: {
      ai: {},
      bot: {},
      speaker: {},
      database: {},
      mi: {},
      concurrent: {}
    },
    loading: false
  }),

  actions: {
    async fetchConfig() {
      this.loading = true
      try {
        const res = await configAPI.getConfig()
        this.config = res.data
      } catch (error) {
        console.error('获取配置失败:', error)
      } finally {
        this.loading = false
      }
    },

    async updateConfig(newConfig) {
      this.loading = true
      try {
        await configAPI.updateConfig(newConfig)
        this.config = { ...this.config, ...newConfig }
        return true
      } catch (error) {
        console.error('更新配置失败:', error)
        return false
      } finally {
        this.loading = false
      }
    },

    async reloadConfig() {
      try {
        await configAPI.reloadConfig()
        await this.fetchConfig()
        return true
      } catch (error) {
        console.error('重新加载配置失败:', error)
        return false
      }
    }
  }
})

// 系统状态Store
export const useSystemStore = defineStore('system', {
  state: () => ({
    status: {
      isRunning: false,
      startTime: null,
      version: '1.0.0',
      goVersion: '1.21',
      webServerPort: 8080,
      databasePath: '',
      concurrentMode: false
    },
    logs: [],
    logLevel: 'info',
    logStats: {
      total: 0,
      error: 0,
      warn: 0,
      info: 0,
      debug: 0
    },
    loading: false
  }),

  actions: {
    async fetchStatus() {
      this.loading = true
      try {
        const res = await systemAPI.getStatus()
        this.status = res.data
      } catch (error) {
        console.error('获取系统状态失败:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchLogs(params = {}) {
      try {
        const res = await systemAPI.getLogs(params)
        const data = res.data
        
        // 如果返回的是新格式的数据
        if (data.entries && Array.isArray(data.entries)) {
          this.logs = data.entries
          this.logLevel = data.level || 'info'
          
          // 计算日志统计
          this.logStats = {
            total: data.total || data.entries.length,
            error: data.entries.filter(log => log.level === 'error').length,
            warn: data.entries.filter(log => log.level === 'warning' || log.level === 'warn').length,
            info: data.entries.filter(log => log.level === 'info').length,
            debug: data.entries.filter(log => log.level === 'debug').length
          }
        } else if (data.logs && Array.isArray(data.logs)) {
          // 兼容旧格式
          this.logs = data.logs.map(logStr => ({
            time: new Date().toISOString(),
            level: this.extractLogLevel(logStr),
            message: logStr
          }))
        } else if (Array.isArray(data)) {
          // 直接是日志数组
          this.logs = data.map(logStr => ({
            time: new Date().toISOString(),
            level: this.extractLogLevel(logStr),
            message: logStr
          }))
        }
      } catch (error) {
        console.error('获取系统日志失败:', error)
      }
    },

    async clearLogs() {
      try {
        await systemAPI.clearLogs()
        this.logs = []
        this.logStats = { total: 0, error: 0, warn: 0, info: 0, debug: 0 }
        return true
      } catch (error) {
        console.error('清空日志失败:', error)
        return false
      }
    },

    async setLogLevel(level) {
      try {
        const res = await systemAPI.setLogLevel(level)
        this.logLevel = res.data.level
        return true
      } catch (error) {
        console.error('设置日志级别失败:', error)
        return false
      }
    },

    async fetchLogLevel() {
      try {
        const res = await systemAPI.getLogLevel()
        this.logLevel = res.data.level
      } catch (error) {
        console.error('获取日志级别失败:', error)
      }
    },

    // 从日志字符串中提取日志级别
    extractLogLevel(logStr) {
      if (logStr.includes('[ERROR]') || logStr.includes('error')) return 'error'
      if (logStr.includes('[WARN]') || logStr.includes('warn')) return 'warn'
      if (logStr.includes('[DEBUG]') || logStr.includes('debug')) return 'debug'
      return 'info'
    },

    async restart() {
      try {
        await systemAPI.restart()
        return true
      } catch (error) {
        console.error('重启系统失败:', error)
        return false
      }
    }
  }
})

// 音箱状态Store
export const useSpeakerStore = defineStore('speaker', {
  state: () => ({
    status: {
      connected: false,
      deviceID: '',
      name: '',
      isPlaying: false,
      volume: 50,
      lastMessage: ''
    },
    loading: false
  }),

  actions: {
    async fetchStatus() {
      this.loading = true
      try {
        const res = await speakerAPI.getStatus()
        this.status = res.data
      } catch (error) {
        console.error('获取音箱状态失败:', error)
      } finally {
        this.loading = false
      }
    },

    async playTTS(text) {
      try {
        await speakerAPI.playTTS(text)
        return true
      } catch (error) {
        console.error('播放TTS失败:', error)
        return false
      }
    },

    async stop() {
      try {
        await speakerAPI.stop()
        await this.fetchStatus()
        return true
      } catch (error) {
        console.error('停止音箱失败:', error)
        return false
      }
    }
  }
})

// 并发处理Store
export const useConcurrentStore = defineStore('concurrent', {
  state: () => ({
    status: {
      enabled: false,
      workerCount: 4,
      queueSize: 100,
      currentQueueSize: 0,
      processedTasks: 0,
      activeWorkers: 0,
      averageProcessTime: '0ms'
    },
    loading: false
  }),

  actions: {
    async fetchStatus() {
      this.loading = true
      try {
        const res = await concurrentAPI.getStatus()
        this.status = res.data
      } catch (error) {
        console.error('获取并发状态失败:', error)
      } finally {
        this.loading = false
      }
    },

    async updateConfig(config) {
      try {
        await concurrentAPI.updateConfig(config)
        await this.fetchStatus()
        return true
      } catch (error) {
        console.error('更新并发配置失败:', error)
        return false
      }
    }
  }
}) 