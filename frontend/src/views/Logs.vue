<template>
  <div class="logs-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>系统日志</span>
          <div class="header-actions">
            <el-switch
              v-model="autoRefresh"
              active-text="自动刷新"
              @change="toggleAutoRefresh"
            />
            <el-button size="small" @click="refreshLogs">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button size="small" @click="clearLogs">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 日志过滤器 -->
      <div class="log-filters">
        <el-row :gutter="15">
          <el-col :span="6">
            <el-select v-model="logLevel" placeholder="过滤级别" clearable @change="filterLogs">
              <el-option label="全部" value="" />
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warn" />
              <el-option label="信息" value="info" />
              <el-option label="调试" value="debug" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="currentLogLevel" placeholder="系统日志级别" @change="setLogLevel">
              <template #prefix>
                <span style="color: #909399; font-size: 12px;">系统级别:</span>
              </template>
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warn" />
              <el-option label="信息" value="info" />
              <el-option label="调试" value="debug" />
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索日志内容..."
              clearable
              @input="filterLogs"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="4">
            <el-select v-model="maxLines" placeholder="显示行数" @change="filterLogs">
              <el-option label="最近100行" :value="100" />
              <el-option label="最近500行" :value="500" />
              <el-option label="最近1000行" :value="1000" />
              <el-option label="全部" :value="0" />
            </el-select>
          </el-col>
        </el-row>
        <el-row :gutter="15" style="margin-top: 10px;">
          <el-col :span="4">
            <el-checkbox v-model="wordWrap" @change="toggleWordWrap">自动换行</el-checkbox>
          </el-col>
          <el-col :span="20">
            <div class="level-info">
              <el-tag size="small" type="info">当前系统日志级别: {{ currentLogLevel.toUpperCase() }}</el-tag>
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                调试级别可显示更详细的日志信息
              </span>
            </div>
          </el-col>
        </el-row>
      </div>
      
      <!-- 日志内容 -->
      <div class="log-container">
        <div 
          ref="logContent"
          class="log-content"
          :class="{ 'word-wrap': wordWrap }"
        >
          <div
            v-for="(log, index) in filteredLogs"
            :key="index"
            class="log-line"
            :class="getLogLevelClass(log)"
          >
            <span class="log-time">{{ formatTime(log.time) }}</span>
            <span class="log-level">{{ log.level }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
          
          <div v-if="filteredLogs.length === 0" class="no-logs">
            <el-empty description="暂无日志数据" />
          </div>
        </div>
      </div>
      
      <!-- 日志统计 -->
      <div class="log-stats">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-item">
              <span class="stat-label">总日志数:</span>
              <span class="stat-value">{{ logStats.total || allLogs.length }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <span class="stat-label">错误:</span>
              <span class="stat-value error">{{ logStats.error || getLogCount('error') }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <span class="stat-label">警告:</span>
              <span class="stat-value warning">{{ logStats.warn || getLogCount('warn') }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <span class="stat-label">信息:</span>
              <span class="stat-value info">{{ logStats.info || getLogCount('info') }}</span>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSystemStore } from '../stores'
import { Refresh, Delete, Search } from '@element-plus/icons-vue'

const systemStore = useSystemStore()
const logContent = ref(null)
const autoRefresh = ref(true) // 默认开启自动刷新
const logLevel = ref('')
const searchKeyword = ref('')
const maxLines = ref(500)
const wordWrap = ref(false)
const currentLogLevel = ref('info')
const logStats = ref({
  total: 0,
  error: 0,
  warn: 0,
  info: 0,
  debug: 0
})
let refreshTimer = null

// 从store获取真实日志数据
const allLogs = computed(() => systemStore.logs || [])

// 过滤后的日志
const filteredLogs = computed(() => {
  let logs = [...allLogs.value]
  
  // 按级别过滤
  if (logLevel.value) {
    logs = logs.filter(log => {
      const level = log.level.toLowerCase()
      const filterLevel = logLevel.value.toLowerCase()
      return level === filterLevel || 
             (filterLevel === 'warn' && (level === 'warning' || level === 'warn'))
    })
  }
  
  // 按关键词过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    logs = logs.filter(log => 
      log.message.toLowerCase().includes(keyword) ||
      log.level.toLowerCase().includes(keyword)
    )
  }
  
  // 限制行数
  if (maxLines.value > 0) {
    logs = logs.slice(-maxLines.value)
  }
  
  return logs.reverse() // 最新的在前面
})

// 获取日志级别样式类
const getLogLevelClass = (log) => {
  return `log-${log.level.toLowerCase()}`
}

// 格式化时间
const formatTime = (timeStr) => {
  return new Date(timeStr).toLocaleString()
}

// 获取指定级别的日志数量
const getLogCount = (level) => {
  if (systemStore.logStats && systemStore.logStats[level] !== undefined) {
    return systemStore.logStats[level]
  }
  
  // 如果没有统计数据，从日志中计算
  return allLogs.value.filter(log => {
    const logLevel = log.level.toLowerCase()
    if (level === 'warn') {
      return logLevel === 'warn' || logLevel === 'warning'
    }
    return logLevel === level.toLowerCase()
  }).length
}

// 刷新日志
const refreshLogs = async () => {
  try {
    const params = {}
    if (logLevel.value) {
      params.level = logLevel.value
    }
    if (maxLines.value > 0) {
      params.limit = maxLines.value
    }
    
    await systemStore.fetchLogs(params)
    
    // 更新统计信息
    logStats.value = systemStore.logStats || { total: 0, error: 0, warn: 0, info: 0, debug: 0 }
    
    ElMessage.success('日志刷新成功')
    await nextTick()
    scrollToBottom()
  } catch (error) {
    ElMessage.error('日志刷新失败')
    console.error('刷新日志失败:', error)
  }
}

// 清空日志
const clearLogs = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有日志吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const success = await systemStore.clearLogs()
    if (success) {
      logStats.value = { total: 0, error: 0, warn: 0, info: 0, debug: 0 }
      ElMessage.success('日志已清空')
      await nextTick()
      scrollToBottom()
    }
  } catch {
    // 用户取消
  }
}

// 设置日志级别
const setLogLevel = async (level) => {
  try {
    const success = await systemStore.setLogLevel(level)
    if (success) {
      ElMessage.success(`日志级别已设置为: ${level.toUpperCase()}`)
      // 刷新日志
      await refreshLogs()
    }
  } catch (error) {
    ElMessage.error('设置日志级别失败')
  }
}

// 监听日志级别变化
watch(() => systemStore.logLevel, (newLevel) => {
  currentLogLevel.value = newLevel
}, { immediate: true })

// 过滤日志
const filterLogs = () => {
  nextTick(() => {
    scrollToBottom()
  })
}

// 切换自动换行
const toggleWordWrap = () => {
  // 样式已在CSS中处理
}

// 滚动到底部
const scrollToBottom = () => {
  if (logContent.value) {
    logContent.value.scrollTop = logContent.value.scrollHeight
  }
}

// 自动刷新开关
const toggleAutoRefresh = (value) => {
  if (value) {
    refreshTimer = setInterval(refreshLogs, 5000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

// 监听自动刷新开关
watch(autoRefresh, (newValue) => {
  toggleAutoRefresh(newValue)
}, { immediate: true })

onMounted(async () => {
  await systemStore.fetchLogLevel()
  await refreshLogs()
  // 启动自动刷新（如果开启）
  if (autoRefresh.value) {
    toggleAutoRefresh(true)
  }
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.logs-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}

.log-filters {
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.log-container {
  height: 500px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}

.log-content {
  height: 100%;
  overflow-y: auto;
  padding: 10px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  background: #2d3748;
  color: #e2e8f0;
}

.log-content.word-wrap {
  white-space: pre-wrap;
  word-break: break-all;
}

.log-line {
  margin-bottom: 2px;
  padding: 2px 5px;
  border-radius: 2px;
}

.log-time {
  color: #a0aec0;
  margin-right: 10px;
}

.log-level {
  font-weight: bold;
  margin-right: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 10px;
}

.log-message {
  color: #e2e8f0;
}

.log-error {
  background-color: rgba(245, 108, 108, 0.1);
}

.log-error .log-level {
  background-color: #f56565;
  color: white;
}

.log-warn {
  background-color: rgba(237, 137, 54, 0.1);
}

.log-warn .log-level {
  background-color: #ed8936;
  color: white;
}

.log-info .log-level {
  background-color: #4299e1;
  color: white;
}

.log-debug .log-level {
  background-color: #9f7aea;
  color: white;
}

.no-logs {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #a0aec0;
}

.log-stats {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-item {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-right: 5px;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.stat-value.error {
  color: #f56565;
}

.stat-value.warning {
  color: #ed8936;
}

.stat-value.info {
  color: #4299e1;
}

/* 滚动条样式 */
.log-content::-webkit-scrollbar {
  width: 6px;
}

.log-content::-webkit-scrollbar-track {
  background: #4a5568;
}

.log-content::-webkit-scrollbar-thumb {
  background: #718096;
  border-radius: 3px;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: #a0aec0;
}
</style> 