<template>
  <div class="speaker-page">
    <el-row :gutter="20">
      <!-- 音箱状态 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>音箱状态</span>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="设备名称">{{ speakerStore.status.name || '未设置' }}</el-descriptions-item>
            <el-descriptions-item label="设备标识">{{ speakerStore.status.deviceID || '未设置' }}</el-descriptions-item>
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
            <el-descriptions-item label="音量">{{ speakerStore.status.volume || 0 }}%</el-descriptions-item>
            <el-descriptions-item label="最后消息">{{ speakerStore.status.lastMessage || '无' }}</el-descriptions-item>
          </el-descriptions>
          
          <div class="status-actions">
            <el-button @click="refreshStatus" :loading="speakerStore.loading">
              <el-icon><Refresh /></el-icon>
              刷新状态
            </el-button>
          </div>
        </el-card>
      </el-col>
      
      <!-- TTS测试 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>TTS语音测试</span>
          </template>
          
          <el-form @submit.prevent="playTTS">
            <el-form-item label="测试文本">
              <el-input
                v-model="ttsText"
                type="textarea"
                :rows="4"
                placeholder="请输入要播放的文本内容..."
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="playTTS" :loading="playing">
                <el-icon><VideoPlay /></el-icon>
                播放语音
              </el-button>
              <el-button @click="stopSpeaker" :loading="stopping">
                <el-icon><VideoPause /></el-icon>
                停止播放
              </el-button>
            </el-form-item>
          </el-form>
          
          <!-- 预设文本快捷按钮 -->
          <div class="preset-texts">
            <h4>快捷测试文本:</h4>
            <div class="preset-buttons">
              <el-button
                v-for="preset in presetTexts"
                :key="preset"
                size="small"
                @click="ttsText = preset"
              >
                {{ preset }}
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 语音命令执行 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span>语音命令执行</span>
            <span style="color: #909399; margin-left: 10px; font-size: 12px;">
              基于xiaoai-tts第三方库，模拟小爱音箱语音交互
            </span>
          </template>
          
          <el-form @submit.prevent="executeCommand">
            <el-row :gutter="16">
              <el-col :span="16">
                <el-form-item label="语音命令">
                  <el-input
                    v-model="voiceCommand"
                    placeholder="例如：查询天气、播放音乐、设置闹钟..."
                    clearable
                  />
                </el-form-item>
              </el-col>
              <el-col :span="4">
                <el-form-item label="需要回应">
                  <el-switch v-model="needResponse" />
                </el-form-item>
              </el-col>
              <el-col :span="4">
                <el-form-item label=" ">
                  <el-button type="primary" @click="executeCommand" :loading="executing">
                    <el-icon><Microphone /></el-icon>
                    执行命令
                  </el-button>
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
          
          <!-- 预设语音命令 -->
          <div class="preset-commands">
            <h4>常用语音命令:</h4>
            <div class="preset-buttons">
              <el-button
                v-for="command in voiceCommands"
                :key="command"
                size="small"
                @click="voiceCommand = command"
              >
                {{ command }}
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useSpeakerStore } from '../stores'
import { speakerAPI } from '../api'
import { Refresh, VideoPlay, VideoPause, Microphone } from '@element-plus/icons-vue'

const speakerStore = useSpeakerStore()
const ttsText = ref('你好，我是小爱同学！')
const playing = ref(false)
const stopping = ref(false)

// 语音命令相关
const voiceCommand = ref('')
const needResponse = ref(true)
const executing = ref(false)

// 预设测试文本
const presetTexts = [
  '你好，我是小爱同学！',
  '今天天气怎么样？',
  '播放一首好听的音乐',
  '现在几点了？',
  '帮我设个闹钟',
  '讲个笑话给我听',
  '今天是几月几号？',
  '明天的天气预报'
]

// 预设语音命令
const voiceCommands = [
  '今天天气怎么样',
  '播放音乐',
  '现在几点了',
  '设置10分钟后的闹钟',
  '音量调到50',
  '暂停播放',
  '继续播放',
  '下一首',
  '讲个笑话',
  '查询明天天气',
  '开启静音模式',
  '关闭静音模式'
]

// 刷新音箱状态
const refreshStatus = async () => {
  try {
    await speakerStore.fetchStatus()
    ElMessage.success('状态刷新成功')
  } catch (error) {
    ElMessage.error('状态刷新失败: ' + error.message)
  }
}

// 播放TTS
const playTTS = async () => {
  if (!ttsText.value.trim()) {
    ElMessage.warning('请输入要播放的文本')
    return
  }
  
  playing.value = true
  try {
    const success = await speakerStore.playTTS(ttsText.value)
    if (success) {
      ElMessage.success('TTS播放指令已发送')
    }
  } finally {
    playing.value = false
  }
}

// 停止音箱
const stopSpeaker = async () => {
  stopping.value = true
  try {
    const success = await speakerStore.stop()
    if (success) {
      ElMessage.success('音箱停止指令已发送')
    }
  } finally {
    stopping.value = false
  }
}

// 执行语音命令
const executeCommand = async () => {
  if (!voiceCommand.value.trim()) {
    ElMessage.warning('请输入要执行的语音命令')
    return
  }
  
  executing.value = true
  try {
    const response = await speakerAPI.executeVoiceCommand(voiceCommand.value, needResponse.value)
    if (response.success) {
      ElMessage.success('语音命令执行成功')
    }
  } catch (error) {
    ElMessage.error('语音命令执行失败: ' + error.message)
  } finally {
    executing.value = false
  }
}

onMounted(() => {
  refreshStatus()
})
</script>

<style scoped>
.speaker-page {
  max-width: 1200px;
  margin: 0 auto;
}

.status-actions {
  margin-top: 20px;
  text-align: center;
}

.preset-texts,
.preset-commands {
  margin-top: 20px;
}

.preset-texts h4,
.preset-commands h4 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 14px;
}

.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-buttons .el-button {
  margin: 0;
}
</style> 