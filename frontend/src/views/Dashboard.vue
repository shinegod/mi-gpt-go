<template>
  <div class="dashboard">
    <!-- 系统状态卡片 -->
    <el-row :gutter="20" class="status-cards">
      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#67C23A">
              <CircleCheck v-if="systemStore.status.isRunning" />
              <CircleClose v-else />
            </el-icon>
            <div class="status-info">
              <h3>系统状态</h3>
              <p>{{ systemStore.status.isRunning ? '运行中' : '已停止' }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#409EFF">
              <Microphone />
            </el-icon>
            <div class="status-info">
              <h3>音箱状态</h3>
              <p>{{ speakerStore.status.connected ? '已连接' : '未连接' }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#E6A23C">
              <Operation />
            </el-icon>
            <div class="status-info">
              <h3>并发模式</h3>
              <p>{{ concurrentStore.status.enabled ? '已启用' : '已禁用' }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#F56C6C">
              <TrendCharts />
            </el-icon>
            <div class="status-info">
              <h3>处理任务</h3>
              <p>{{ concurrentStore.status.processedTasks }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统信息 -->
    <el-row :gutter="20" class="info-section">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
              <el-button size="small" @click="refreshData">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="版本">{{ systemStore.status.version }}</el-descriptions-item>
            <el-descriptions-item label="Go版本">{{ systemStore.status.goVersion }}</el-descriptions-item>
            <el-descriptions-item label="Web端口">{{ systemStore.status.webServerPort }}</el-descriptions-item>
            <el-descriptions-item label="数据库路径">{{ systemStore.status.databasePath }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">
              {{ systemStore.status.startTime ? new Date(systemStore.status.startTime).toLocaleString() : '未知' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>音箱信息</span>
            </div>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="设备名称">{{ speakerStore.status.name }}</el-descriptions-item>
            <el-descriptions-item label="设备标识">{{ speakerStore.status.deviceID }}</el-descriptions-item>
            <el-descriptions-item label="连接状态">
              <el-tag :type="speakerStore.status.connected ? 'success' : 'danger'">
                {{ speakerStore.status.connected ? '已连接' : '未连接' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="播放状态">
              <el-tag :type="speakerStore.status.isPlaying ? 'warning' : 'info'">
                {{ speakerStore.status.isPlaying ? '播放中' : '空闲' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最后消息">{{ speakerStore.status.lastMessage }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- 并发处理统计 -->
    <el-row :gutter="20" class="concurrent-section" v-if="concurrentStore.status.enabled">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>并发处理统计</span>
              <el-switch
                v-model="autoRefresh"
                active-text="自动刷新"
                @change="toggleAutoRefresh"
              />
            </div>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="8">
              <div class="stat-item">
                <h3>{{ concurrentStore.status.workerCount }}</h3>
                <p>工作线程</p>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-item">
                <h3>{{ concurrentStore.status.currentQueueSize }}/{{ concurrentStore.status.queueSize }}</h3>
                <p>队列使用率</p>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-item">
                <h3>{{ concurrentStore.status.averageProcessTime }}</h3>
                <p>平均处理时间</p>
              </div>
            </el-col>
          </el-row>
          
          <el-progress
            :percentage="Math.round((concurrentStore.status.currentQueueSize / concurrentStore.status.queueSize) * 100)"
            :stroke-width="20"
            :text-inside="true"
            class="queue-progress"
          />
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷操作 -->
    <el-row :gutter="20" class="quick-actions">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span>快捷操作</span>
          </template>
          
          <div class="action-buttons">
            <el-button type="primary" @click="testTTS">
              <el-icon><Microphone /></el-icon>
              测试TTS
            </el-button>
            
            <el-button 
              type="success" 
              @click="startSpeaker" 
              :loading="startingService"
              v-if="!speakerStore.status?.connected"
            >
              <el-icon><Operation /></el-icon>
              启动音箱服务
            </el-button>
            
            <el-button type="success" @click="reloadConfig">
              <el-icon><RefreshRight /></el-icon>
              重新加载配置
            </el-button>
            
            <el-button type="warning" @click="restartSystem">
              <el-icon><Refresh /></el-icon>
              重启系统
            </el-button>
            
            <el-button @click="$router.push('/config')">
              <el-icon><Setting /></el-icon>
              配置管理
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSystemStore, useSpeakerStore, useConcurrentStore, useConfigStore } from '../stores'
import {
  CircleCheck,
  CircleClose,
  Microphone,
  Operation,
  TrendCharts,
  Refresh,
  RefreshRight,
  Setting
} from '@element-plus/icons-vue'

const systemStore = useSystemStore()
const speakerStore = useSpeakerStore()
const concurrentStore = useConcurrentStore()
const configStore = useConfigStore()

const autoRefresh = ref(true)
const startingService = ref(false)
let refreshTimer = null

// 刷新所有数据
const refreshData = async () => {
  await Promise.all([
    systemStore.fetchStatus(),
    speakerStore.fetchStatus(),
    concurrentStore.fetchStatus()
  ])
}

// 自动刷新开关
const toggleAutoRefresh = (value) => {
  if (value) {
    refreshTimer = setInterval(refreshData, 5000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

// 测试TTS
const testTTS = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入要播放的文本', '测试TTS', {
      confirmButtonText: '播放',
      cancelButtonText: '取消',
      inputValue: '你好，我是小爱同学！'
    })
    
    if (value) {
      const success = await speakerStore.playTTS(value)
      if (success) {
        ElMessage.success('TTS播放指令已发送')
      }
    }
  } catch {
    // 用户取消
  }
}

// 重新加载配置
const reloadConfig = async () => {
  try {
    await ElMessageBox.confirm('确定要重新加载配置吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const success = await configStore.reloadConfig()
    if (success) {
      ElMessage.success('配置重新加载成功')
      await refreshData()
    }
  } catch {
    // 用户取消
  }
}

// 启动音箱服务
const startSpeaker = async () => {
  startingService.value = true
  try {
    const { configAPI } = await import('../api')
    const result = await configAPI.startSpeaker()
    
    if (result.success) {
      ElMessage.success('音箱服务启动成功')
      await refreshData()
    }
  } catch (error) {
    ElMessage.error('启动音箱服务失败，请检查配置')
  } finally {
    startingService.value = false
  }
}

// 重启系统
const restartSystem = async () => {
  try {
    await ElMessageBox.confirm('确定要重启系统吗？这可能需要几秒钟时间。', '警告', {
      confirmButtonText: '确定重启',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const success = await systemStore.restart()
    if (success) {
      ElMessage.success('系统重启指令已发送')
    }
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  refreshData()
  // 启动自动刷新（默认开启）
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
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.status-cards {
  margin-bottom: 20px;
}

.status-card {
  height: 120px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 15px;
  height: 100%;
}

.status-info h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.status-info p {
  margin: 5px 0 0 0;
  font-size: 14px;
  color: #909399;
}

.info-section,
.concurrent-section,
.quick-actions {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-item {
  text-align: center;
  padding: 20px 0;
}

.stat-item h3 {
  margin: 0;
  font-size: 24px;
  color: #409EFF;
  font-weight: 600;
}

.stat-item p {
  margin: 5px 0 0 0;
  color: #909399;
  font-size: 14px;
}

.queue-progress {
  margin-top: 20px;
}

.action-buttons {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

:deep(.el-card__body) {
  padding: 20px;
}
</style> 