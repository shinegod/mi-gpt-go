import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const res = response.data
    if (res.success) {
      return res
    } else {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
  },
  error => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

// 配置相关API
export const configAPI = {
  // 获取配置
  getConfig() {
    return api.get('/config')
  },
  
  // 更新配置
  updateConfig(data) {
    return api.put('/config', data)
  },
  
  // 重新加载配置
  reloadConfig() {
    return api.post('/config/reload')
  },
  
  // 获取配置模板
  getConfigTemplate() {
    return api.get('/config/template')
  },
  
  // 验证配置
  validateConfig() {
    return api.post('/config/validate')
  },
  
  // 测试AI连接
  testAIConnection() {
    return api.post('/config/test-ai')
  },
  
  // 测试小米设备连接
  testMiConnection() {
    return api.post('/config/test-mi')
  },
  
  // 启动音箱服务
  startSpeaker() {
    return api.post('/config/start-speaker')
  }
}

// 系统相关API
export const systemAPI = {
  // 获取系统状态
  getStatus() {
    return api.get('/system/status')
  },
  
  // 重启系统
  restart() {
    return api.post('/system/restart')
  },
  
  // 获取系统日志
  getLogs(params = {}) {
    return api.get('/system/logs', { params })
  },
  
  // 清空系统日志
  clearLogs() {
    return api.delete('/system/logs')
  },
  
  // 设置日志级别
  setLogLevel(level) {
    return api.put('/system/log-level', { level })
  },
  
  // 获取当前日志级别
  getLogLevel() {
    return api.get('/system/log-level')
  },
  

}

// 音箱相关API
export const speakerAPI = {
  // 获取音箱状态
  getStatus() {
    return api.get('/speaker/status')
  },
  
  // 播放TTS
  playTTS(text) {
    return api.post('/speaker/play', { text })
  },
  
  // 停止音箱
  stop() {
    return api.post('/speaker/stop')
  },
  
  // 重启音箱服务（配置热重载）
  restartSpeaker() {
    return api.post('/speaker/restart')
  },
  
  // 执行语音命令
  executeVoiceCommand(command, needResponse = true) {
    return api.post('/speaker/execute', { 
      command, 
      needResponse 
    })
  }
}

// 并发处理相关API
export const concurrentAPI = {
  // 获取并发状态
  getStatus() {
    return api.get('/concurrent/status')
  },
  
  // 更新并发配置
  updateConfig(data) {
    return api.put('/concurrent/config', data)
  }
}

export default api 