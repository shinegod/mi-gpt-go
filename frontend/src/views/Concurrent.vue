<template>
  <div class="concurrent-page">
    <el-row :gutter="20">
      <!-- 并发状态概览 -->
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>并发处理状态</span>
              <el-switch
                v-model="autoRefresh"
                active-text="自动刷新"
                @change="toggleAutoRefresh"
              />
            </div>
          </template>
          
          <el-row :gutter="20" class="status-overview">
            <el-col :span="6">
              <div class="status-item">
                <h3>{{ concurrentStore.status.enabled ? '已启用' : '已禁用' }}</h3>
                <p>并发处理</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="status-item">
                <h3>{{ concurrentStore.status.workerCount }}</h3>
                <p>工作线程</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="status-item">
                <h3>{{ concurrentStore.status.currentQueueSize }}/{{ concurrentStore.status.queueSize }}</h3>
                <p>队列使用</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="status-item">
                <h3>{{ concurrentStore.status.processedTasks }}</h3>
                <p>已处理任务</p>
              </div>
            </el-col>
          </el-row>
          
          <el-row :gutter="20" class="performance-metrics">
            <el-col :span="12">
              <div class="metric-card">
                <h4>队列使用率</h4>
                <el-progress
                  :percentage="queueUsagePercentage"
                  :stroke-width="20"
                  :text-inside="true"
                  :color="queueUsageColor"
                />
              </div>
            </el-col>
            <el-col :span="12">
              <div class="metric-card">
                <h4>平均处理时间</h4>
                <div class="process-time">{{ concurrentStore.status.averageProcessTime }}</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <!-- 并发配置管理 -->
    <el-row :gutter="20" class="config-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>并发配置管理</span>
              <el-button type="primary" size="small" @click="saveConfig" :loading="saving">
                <el-icon><Check /></el-icon>
                保存配置
              </el-button>
            </div>
          </template>
          
          <el-form :model="configForm" label-width="140px" class="config-form">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="启用并发处理">
                  <el-switch v-model="configForm.enabled" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="启用性能指标">
                  <el-switch v-model="configForm.enableMetrics" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="工作线程数">
                  <el-input-number
                    v-model="configForm.workerCount"
                    :min="1"
                    :max="16"
                    style="width: 100%"
                  />
                  <div class="form-tip">推荐设置为CPU核心数</div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="队列大小">
                  <el-input-number
                    v-model="configForm.queueSize"
                    :min="10"
                    :max="1000"
                    :step="10"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="消息缓冲区大小">
                  <el-input-number
                    v-model="configForm.messageBufferSize"
                    :min="10"
                    :max="200"
                    :step="10"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="速率限制(每秒)">
                  <el-input-number
                    v-model="configForm.rateLimit"
                    :min="1"
                    :max="100"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="批处理大小">
                  <el-input-number
                    v-model="configForm.batchSize"
                    :min="1"
                    :max="20"
                    style="width: 100%"
                  />
                  <div class="form-tip">1表示禁用批处理</div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="批处理超时(秒)">
                  <el-input-number
                    v-model="configForm.batchTimeoutSeconds"
                    :min="1"
                    :max="30"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <!-- 性能建议 -->
    <el-row :gutter="20" class="suggestions-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span>性能优化建议</span>
          </template>
          
          <el-alert
            v-for="suggestion in performanceSuggestions"
            :key="suggestion.type"
            :title="suggestion.title"
            :description="suggestion.description"
            :type="suggestion.type"
            :closable="false"
            class="suggestion-alert"
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useConcurrentStore } from '../stores'
import { Check } from '@element-plus/icons-vue'

const concurrentStore = useConcurrentStore()
const autoRefresh = ref(true)
const saving = ref(false)
let refreshTimer = null

// 配置表单
const configForm = reactive({
  enabled: true,
  workerCount: 4,
  queueSize: 100,
  messageBufferSize: 50,
  rateLimit: 10,
  batchSize: 5,
  batchTimeoutSeconds: 5,
  enableMetrics: true
})

// 队列使用率
const queueUsagePercentage = computed(() => {
  if (concurrentStore.status.queueSize === 0) return 0
  return Math.round((concurrentStore.status.currentQueueSize / concurrentStore.status.queueSize) * 100)
})

// 队列使用率颜色
const queueUsageColor = computed(() => {
  const percentage = queueUsagePercentage.value
  if (percentage < 50) return '#67C23A'
  if (percentage < 80) return '#E6A23C'
  return '#F56C6C'
})

// 性能建议
const performanceSuggestions = computed(() => {
  const suggestions = []
  
  if (!concurrentStore.status.enabled) {
    suggestions.push({
      type: 'warning',
      title: '并发处理未启用',
      description: '建议启用并发处理以提升系统性能，特别是在高负载场景下。'
    })
  }
  
  if (queueUsagePercentage.value > 80) {
    suggestions.push({
      type: 'error',
      title: '队列使用率过高',
      description: '当前队列使用率超过80%，建议增加队列大小或优化处理逻辑。'
    })
  }
  
  if (concurrentStore.status.workerCount < 2) {
    suggestions.push({
      type: 'info',
      title: '工作线程数较少',
      description: '建议根据CPU核心数适当增加工作线程数以提升并发处理能力。'
    })
  }
  
  if (suggestions.length === 0) {
    suggestions.push({
      type: 'success',
      title: '配置良好',
      description: '当前并发处理配置运行良好，系统性能表现正常。'
    })
  }
  
  return suggestions
})

// 刷新数据
const refreshData = async () => {
  await concurrentStore.fetchStatus()
}

// 自动刷新开关
const toggleAutoRefresh = (value) => {
  if (value) {
    refreshTimer = setInterval(refreshData, 3000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

// 保存配置
const saveConfig = async () => {
  saving.value = true
  try {
    const success = await concurrentStore.updateConfig(configForm)
    if (success) {
      ElMessage.success('并发配置保存成功')
    }
  } finally {
    saving.value = false
  }
}

// 初始化配置表单
const initConfigForm = () => {
  Object.assign(configForm, {
    enabled: concurrentStore.status.enabled,
    workerCount: concurrentStore.status.workerCount,
    queueSize: concurrentStore.status.queueSize
  })
}

onMounted(() => {
  refreshData().then(() => {
    initConfigForm()
    // 启动自动刷新（默认开启）
    if (autoRefresh.value) {
      toggleAutoRefresh(true)
    }
  })
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.concurrent-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-overview {
  margin-bottom: 20px;
}

.status-item {
  text-align: center;
  padding: 20px 0;
}

.status-item h3 {
  margin: 0;
  font-size: 24px;
  color: #409EFF;
  font-weight: 600;
}

.status-item p {
  margin: 5px 0 0 0;
  color: #909399;
  font-size: 14px;
}

.performance-metrics {
  margin-top: 20px;
}

.metric-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.metric-card h4 {
  margin: 0 0 15px 0;
  color: #303133;
  font-size: 16px;
}

.process-time {
  font-size: 24px;
  font-weight: 600;
  color: #409EFF;
}

.config-section,
.suggestions-section {
  margin-top: 20px;
}

.config-form {
  max-width: 800px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.suggestion-alert {
  margin-bottom: 10px;
}

.suggestion-alert:last-child {
  margin-bottom: 0;
}
</style> 