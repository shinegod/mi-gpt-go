<template>
  <div class="config-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>配置管理</span>
          <div class="header-actions">
            <el-button size="small" @click="loadConfig">
              <el-icon><Refresh /></el-icon>
              重新加载
            </el-button>
            <el-button size="small" @click="validateConfig" :loading="validating">
              验证配置
            </el-button>
            <el-button size="small" @click="testConnections" :loading="testing">
              测试连接
            </el-button>
            <el-button type="primary" size="small" @click="saveConfig" :loading="saving">
              <el-icon><Check /></el-icon>
              保存配置
            </el-button>
            <el-button type="success" size="small" @click="applyConfig" :loading="applying">
              🔄 应用配置
            </el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="config-tabs">
        <!-- AI服务配置 -->
        <el-tab-pane label="AI服务" name="ai">
          <el-form :model="configForm.ai" label-width="140px" class="config-form">
            <el-alert
              title="AI服务配置"
              type="info"
              :closable="false"
              style="margin-bottom: 20px;"
            >
              请根据选择的AI服务提供商填写相应的配置信息
            </el-alert>
            
            <el-form-item label="服务提供商" required>
              <el-select v-model="configForm.ai.provider" placeholder="请选择AI服务提供商">
                <el-option label="OpenAI" value="openai" />
                <el-option label="Azure OpenAI" value="azure" />
                <el-option label="DeepSeek" value="deepseek" />
              </el-select>
            </el-form-item>
            
            <!-- OpenAI 配置 -->
            <template v-if="configForm.ai.provider === 'openai'">
              <el-form-item label="API Key" required>
                <el-input 
                  v-model="configForm.ai.apiKey" 
                  type="password" 
                  placeholder="sk-..." 
                  show-password
                  clearable
                />
                <div class="form-tip">OpenAI API密钥</div>
              </el-form-item>
              
              <el-form-item label="模型名称" required>
                <el-select v-model="configForm.ai.model" placeholder="请选择模型">
                  <el-option label="gpt-4o (推荐)" value="gpt-4o" />
                  <el-option label="gpt-4o-mini" value="gpt-4o-mini" />
                  <el-option label="gpt-4-turbo" value="gpt-4-turbo" />
                  <el-option label="gpt-3.5-turbo" value="gpt-3.5-turbo" />
                </el-select>
              </el-form-item>
            </template>
            
            <!-- Azure OpenAI 配置 -->
            <template v-if="configForm.ai.provider === 'azure'">
              <el-form-item label="Azure API Key" required>
                <el-input 
                  v-model="configForm.ai.azureAPIKey" 
                  type="password" 
                  placeholder="Azure OpenAI API密钥" 
                  show-password
                  clearable
                />
              </el-form-item>
              
              <el-form-item label="Azure端点" required>
                <el-input 
                  v-model="configForm.ai.azureEndpoint" 
                  placeholder="https://your-resource.openai.azure.com" 
                  clearable
                />
              </el-form-item>
              
              <el-form-item label="模型名称" required>
                <el-select v-model="configForm.ai.model" placeholder="请选择模型">
                  <el-option label="gpt-4 (推荐)" value="gpt-4" />
                  <el-option label="gpt-4o" value="gpt-4o" />
                  <el-option label="gpt-35-turbo" value="gpt-35-turbo" />
                </el-select>
              </el-form-item>
              
              <el-form-item label="部署名称" required>
                <el-input 
                  v-model="configForm.ai.azureDeployment" 
                  placeholder="gpt-4" 
                  clearable
                />
                <div class="form-tip">Azure OpenAI 中的部署名称</div>
              </el-form-item>
            </template>
            
            <!-- DeepSeek 配置 -->
            <template v-if="configForm.ai.provider === 'deepseek'">
              <el-form-item label="DeepSeek API Key" required>
                <el-input 
                  v-model="configForm.ai.deepSeekAPIKey" 
                  placeholder="请输入DeepSeek API Key"
                  show-password
                  clearable
                />
              </el-form-item>
              
              <el-form-item label="模型名称" required>
                <el-select v-model="configForm.ai.model" placeholder="请选择模型">
                  <el-option label="deepseek-chat (推荐)" value="deepseek-chat" />
                  <el-option label="deepseek-coder" value="deepseek-coder" />
                  <el-option label="deepseek-reasoner" value="deepseek-reasoner" />
                </el-select>
              </el-form-item>
              
              <el-form-item label="API基础URL">
                <el-input 
                  v-model="configForm.ai.deepSeekBaseURL" 
                  placeholder="留空使用默认URL"
                  clearable
                />
                <div class="form-help">自定义API基础地址（可选）</div>
              </el-form-item>
            </template>
            
            <!-- 通用配置（所有AI服务商都可用） -->
            <el-form-item label="API基础URL" v-if="configForm.ai.provider !== 'deepseek'">
              <el-input v-model="configForm.ai.baseURL" placeholder="留空使用默认URL" />
              <div class="form-tip">自定义API基础地址（可选）</div>
            </el-form-item>
            
            <el-form-item label="代理URL">
              <el-input v-model="configForm.ai.proxyURL" placeholder="可选，留空不使用代理" />
              <div class="form-tip">网络代理地址（可选）</div>
            </el-form-item>
            
            <el-form-item>
              <el-button 
                type="success" 
                @click="testAIConnection"
                :loading="testing"
                size="large"
                style="width: 100%;"
              >
                <template #icon>
                  <el-icon><Check /></el-icon>
                </template>
                测试AI服务连接
              </el-button>
              <div class="form-tip">验证AI服务配置是否正确</div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 机器人配置 -->
        <el-tab-pane label="机器人人设" name="bot">
          <el-form :model="configForm.bot" label-width="120px" class="config-form">
            <el-form-item label="机器人名称">
              <el-input v-model="configForm.bot.name" placeholder="例如: 小爱同学, 傻妞" />
            </el-form-item>
            
            <el-form-item label="机器人人设">
              <el-input
                v-model="configForm.bot.profile"
                type="textarea"
                :rows="4"
                placeholder="描述机器人的性格、特点等"
              />
            </el-form-item>
            
            <el-form-item label="主人名称">
              <el-input v-model="configForm.bot.masterName" placeholder="例如: 主人, 小明" />
            </el-form-item>
            
            <el-form-item label="主人描述">
              <el-input
                v-model="configForm.bot.masterProfile"
                type="textarea"
                :rows="3"
                placeholder="描述主人的特点"
              />
            </el-form-item>
            
            <el-form-item label="房间名称">
              <el-input v-model="configForm.bot.roomName" placeholder="例如: 客厅, 卧室" />
            </el-form-item>
            
            <el-form-item label="房间描述">
              <el-input
                v-model="configForm.bot.roomDescription"
                type="textarea"
                :rows="3"
                placeholder="描述房间的环境"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 音箱配置 -->
        <el-tab-pane label="音箱设置" name="speaker">
          <el-form :model="configForm.speaker" label-width="140px" class="config-form">
            <el-form-item label="音箱名称">
              <el-input v-model="configForm.speaker.name" placeholder="音箱显示名称" />
            </el-form-item>
            
            <el-form-item label="召唤关键词">
              <el-input
                v-model="configForm.speaker.callAIKeywords"
                placeholder="用逗号分隔，例如: 小爱同学,你好小爱"
              />
              <div class="form-tip">说出这些关键词会触发AI对话</div>
            </el-form-item>
            
            <el-form-item label="唤醒关键词">
              <el-input
                v-model="configForm.speaker.wakeupKeywords"
                placeholder="用逗号分隔，例如: 进入连续对话,开始聊天"
              />
              <div class="form-tip">进入连续对话模式的关键词</div>
            </el-form-item>
            
            <el-form-item label="退出关键词">
              <el-input
                v-model="configForm.speaker.exitKeywords"
                placeholder="用逗号分隔，例如: 退出连续对话,再见"
              />
              <div class="form-tip">退出连续对话模式的关键词</div>
            </el-form-item>
            
            <el-form-item label="进入AI提示语">
              <el-input
                v-model="configForm.speaker.onEnterAI"
                placeholder="用逗号分隔多个提示语，随机播放"
              />
            </el-form-item>
            
            <el-form-item label="退出AI提示语">
              <el-input
                v-model="configForm.speaker.onExitAI"
                placeholder="用逗号分隔多个提示语，随机播放"
              />
            </el-form-item>
            
            <el-form-item label="AI思考提示语">
              <el-input
                v-model="configForm.speaker.onAIAsking"
                placeholder="AI思考时播放的提示语"
              />
            </el-form-item>
            
            <el-form-item label="AI回复完成提示语">
              <el-input
                v-model="configForm.speaker.onAIReplied"
                placeholder="AI回复完成后的提示语"
              />
            </el-form-item>
            
            <el-form-item label="AI错误提示语">
              <el-input
                v-model="configForm.speaker.onAIError"
                placeholder="AI出错时的提示语"
              />
            </el-form-item>
            
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item label="流式响应">
                  <el-switch v-model="configForm.speaker.streamResponse" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="音频日志">
                  <el-switch v-model="configForm.speaker.enableAudioLog" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="调试模式">
                  <el-switch v-model="configForm.speaker.debugMode" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-tab-pane>

        <!-- 小米设备配置 -->
        <el-tab-pane label="小米设备" name="mi">
          <el-form :model="configForm.mi" label-width="140px" class="config-form">
            <el-alert
              title="小米设备连接配置"
              type="error"
              :closable="false"
              style="margin-bottom: 20px;"
            >
              <div><strong>⚠️ 常见登录失败原因（状态码405）：</strong></div>
              <div>1. 账号格式错误：必须是手机号或邮箱，不能是QQ号、用户名</div>
              <div>2. 设备标识错误：建议使用设备名称（如"小爱音箱Pro"）或数字设备ID</div>
              <div>3. 开启了二步验证：需要先关闭小米账号的二步验证</div>
              <div>4. 密码错误：请确认是小米账号的登录密码</div>
            </el-alert>
            
            <el-form-item label="小米账号" required>
              <el-input 
                v-model="configForm.mi.userID" 
                placeholder="13812345678 或 user@example.com" 
                clearable
                type="password"
                show-password
              />
              <div class="form-tip">
                <strong style="color: #67C23A;">✅ 正确示例：</strong> 13812345678（手机号）、user@163.com（邮箱）、12345678（数字ID）
              </div>
            </el-form-item>
            
            <el-form-item label="账号密码" required>
              <el-input 
                v-model="configForm.mi.password" 
                type="password" 
                placeholder="小米账号登录密码" 
                show-password
                clearable
              />
              <div class="form-tip">
                <strong style="color: #F56C6C;">🔑 重要：</strong>
                <br>• 必须是小米账号的登录密码（不是设备密码）
                <br>• 如果账号开启了二步验证，请先到小米官网关闭
                <br>• 建议先在小米官网确认能正常登录
              </div>
            </el-form-item>
            
            <el-form-item label="设备ID/名称" required>
              <el-input 
                v-model="configForm.mi.deviceID" 
                placeholder="小爱音箱Pro 或 123456789" 
                clearable
              />
              <div class="form-tip">
                <strong style="color: #409EFF;">📱 设备标识支持两种方式：</strong>
                <br><strong style="color: #67C23A;">✅ 方式1（推荐）：</strong>使用设备名称，如 "小爱音箱Pro"、"小爱音箱Play"
                <br>• 在米家APP中查看设备名称
                <br>• 在小爱音箱APP中查看设备名称
                <br><strong style="color: #67C23A;">✅ 方式2：</strong>使用设备ID，如 "123456789"（纯数字）
                <br>• 在小爱音箱APP → 设备设置 → 设备信息中查看
                <br><strong style="color: #E6A23C;">💡 提示：</strong>设备名称更简单易用，推荐优先使用
              </div>
            </el-form-item>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="检查间隔(毫秒)">
                  <el-input-number
                    v-model="configForm.mi.checkInterval"
                    :min="100"
                    :max="10000"
                    :step="100"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="超时时间(毫秒)">
                  <el-input-number
                    v-model="configForm.mi.timeout"
                    :min="1000"
                    :max="30000"
                    :step="1000"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-form-item label="启用跟踪日志">
              <el-switch v-model="configForm.mi.enableTrace" />
              <div class="form-tip">启用后会记录详细的设备通信日志</div>
            </el-form-item>
            
            <el-form-item>
              <el-button 
                type="primary" 
                @click="testMiConnection"
                :loading="testing"
                size="large"
                style="width: 100%;"
              >
                <template #icon>
                  <el-icon><Connection /></el-icon>
                </template>
                测试小米设备连接
              </el-button>
              <div class="form-tip">验证小米账号和设备ID是否正确</div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 并发处理配置 -->
        <el-tab-pane label="并发处理" name="concurrent">
          <el-form :model="configForm.concurrent" label-width="140px" class="config-form">
            <el-form-item label="启用并发处理">
              <el-switch v-model="configForm.concurrent.enable" />
              <div class="form-tip">开启后可显著提升处理性能</div>
            </el-form-item>
            
            <template v-if="configForm.concurrent.enable">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="工作线程数">
                    <el-input-number
                      v-model="configForm.concurrent.workerCount"
                      :min="1"
                      :max="16"
                    />
                    <div class="form-tip">推荐设置为CPU核心数</div>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="队列大小">
                    <el-input-number
                      v-model="configForm.concurrent.queueSize"
                      :min="10"
                      :max="1000"
                      :step="10"
                    />
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="消息缓冲区大小">
                    <el-input-number
                      v-model="configForm.concurrent.messageBufferSize"
                      :min="10"
                      :max="200"
                      :step="10"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="速率限制(每秒)">
                    <el-input-number
                      v-model="configForm.concurrent.rateLimit"
                      :min="1"
                      :max="100"
                    />
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="批处理大小">
                    <el-input-number
                      v-model="configForm.concurrent.batchSize"
                      :min="1"
                      :max="20"
                    />
                    <div class="form-tip">1表示禁用批处理</div>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="批处理超时(秒)">
                    <el-input-number
                      v-model="configForm.concurrent.batchTimeoutSeconds"
                      :min="1"
                      :max="30"
                    />
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-form-item label="启用性能指标">
                <el-switch v-model="configForm.concurrent.enableMetrics" />
                <div class="form-tip">记录详细的性能统计信息</div>
              </el-form-item>
            </template>
          </el-form>
        </el-tab-pane>

        <!-- 数据库配置 -->
        <el-tab-pane label="数据库" name="database">
          <el-form :model="configForm.database" label-width="120px" class="config-form">
            <el-form-item label="数据库路径">
              <el-input v-model="configForm.database.path" placeholder="数据库文件路径" />
            </el-form-item>
            
            <el-form-item label="调试模式">
              <el-switch v-model="configForm.database.debug" />
              <div class="form-tip">启用后会记录所有SQL查询</div>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useConfigStore } from '../stores'
import { Refresh, Check, Connection, Loading } from '@element-plus/icons-vue'

const configStore = useConfigStore()
const activeTab = ref('ai')
const saving = ref(false)
const validating = ref(false)
const testing = ref(false)
const applying = ref(false)

// 配置表单
const configForm = reactive({
  ai: {
    provider: 'deepseek',
    model: 'deepseek-chat',
    baseURL: '',
    proxyURL: '',
    apiKey: '',
    azureAPIKey: '',
    azureEndpoint: '',
    azureDeployment: '',
    deepSeekAPIKey: '',
    deepSeekBaseURL: ''
  },
  bot: {
    name: '小爱同学',
    profile: '',
    masterName: '主人',
    masterProfile: '',
    roomName: '客厅',
    roomDescription: ''
  },
  speaker: {
    name: '小爱同学',
    callAIKeywords: '',
    wakeupKeywords: '',
    exitKeywords: '',
    onEnterAI: '',
    onExitAI: '',
    onAIAsking: '',
    onAIReplied: '',
    onAIError: '',
    streamResponse: true,
    enableAudioLog: false,
    debugMode: false
  },
  mi: {
    userID: '',
    password: '',
    deviceID: '',
    checkInterval: 1000,
    timeout: 5000,
    enableTrace: false
  },
  concurrent: {
    enable: true,
    workerCount: 4,
    queueSize: 100,
    messageBufferSize: 50,
    rateLimit: 10,
    batchSize: 5,
    batchTimeoutSeconds: 5,
    enableMetrics: true
  },
  database: {
    path: './data/app.db',
    debug: false
  }
})

// 定义各服务提供商的默认模型
const providerDefaultModels = {
  openai: 'gpt-4o',
  azure: 'gpt-4',
  deepseek: 'deepseek-chat'
}

// 监听服务提供商变化，自动更新模型
watch(() => configForm.ai.provider, (newProvider, oldProvider) => {
  if (newProvider !== oldProvider && providerDefaultModels[newProvider]) {
    configForm.ai.model = providerDefaultModels[newProvider]
    ElMessage.info(`已自动切换到 ${newProvider.toUpperCase()} 的推荐模型：${providerDefaultModels[newProvider]}`)
  }
})

// 加载配置
const loadConfig = async () => {
  await configStore.fetchConfig()
  
  // 更新表单数据
  if (configStore.config) {
    Object.assign(configForm.ai, configStore.config.ai)
    Object.assign(configForm.bot, configStore.config.bot)
    Object.assign(configForm.speaker, configStore.config.speaker)
    Object.assign(configForm.mi, configStore.config.mi)
    Object.assign(configForm.concurrent, configStore.config.concurrent)
    Object.assign(configForm.database, configStore.config.database)
  }
}

// 保存配置
const saveConfig = async () => {
  saving.value = true
  try {
    const success = await configStore.updateConfig(configForm)
    if (success) {
      ElMessage.success('配置保存成功！点击"应用配置"按钮来重启音箱服务使配置生效。')
    }
  } catch (error) {
    ElMessage.error('配置保存失败')
  } finally {
    saving.value = false
  }
}

// 应用配置（仅重启音箱服务，不保存）
const applyConfig = async () => {
  applying.value = true
  try {
    const { speakerAPI } = await import('../api')
    const restartResult = await speakerAPI.restartSpeaker()
    if (restartResult.success) {
      ElMessage.success('配置应用成功，音箱服务已重启')
    }
  } catch (error) {
    ElMessage.error('应用配置失败，请检查配置是否完整')
  } finally {
    applying.value = false
  }
}

// 验证配置
const validateConfig = async () => {
  validating.value = true
  try {
    const { configAPI } = await import('../api')
    const result = await configAPI.validateConfig()
    
    if (result.success) {
      const { validation, messages, isComplete } = result.data
      
      if (isComplete) {
        ElMessage.success('配置验证通过！所有必需项都已正确配置。')
      } else {
        let errorMessages = []
        if (!validation.basic && messages.basic) {
          errorMessages.push(`基础配置: ${messages.basic}`)
        }
        if (!validation.ai && messages.ai) {
          errorMessages.push(`AI服务: ${messages.ai}`)
        }
        if (!validation.mi && messages.mi) {
          errorMessages.push(`小米设备: ${messages.mi}`)
        }
        
        ElMessage.warning(`配置验证失败:\n${errorMessages.join('\n')}`)
      }
    }
  } catch (error) {
    ElMessage.error('配置验证失败')
  } finally {
    validating.value = false
  }
}

// 测试AI连接
const testAIConnection = async () => {
  testing.value = true
  try {
    const { configAPI } = await import('../api')
    
    // 首先保存当前配置
    await configStore.updateConfig(configForm)
    
    // 显示测试进度
    ElMessage.info(`正在测试 ${configForm.ai.provider.toUpperCase()} 服务连接...`)
    
    // 测试AI连接
    const aiResult = await configAPI.testAIConnection()
    
    if (aiResult.success) {
      // 显示详细的成功信息
      const { provider, model, testMessage, response } = aiResult.data
      
      await ElMessageBox.alert(
        `🎉 AI服务连接测试成功！\n\n` +
        `📋 服务信息:\n` +
        `• 服务商: ${provider.toUpperCase()}\n` +
        `• 模型: ${model}\n\n` +
        `💬 测试对话:\n` +
        `• 测试消息: ${testMessage}\n` +
        `• AI回复: ${response}\n\n` +
        `✅ 连接状态: 正常\n` +
        `⏱️ 详细日志请查看系统日志页面`,
        'AI服务连接测试结果',
        {
          confirmButtonText: '确定',
          type: 'success',
          dangerouslyUseHTMLString: false
        }
      )
      
      ElMessage.success(`${provider.toUpperCase()} 服务连接测试成功`)
    } else {
      // 显示详细的错误信息
      throw new Error(aiResult.message || '测试失败')
    }
  } catch (error) {
    // 显示详细的错误信息
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    const errorData = error.response?.data?.data
    
    let errorDetails = `❌ AI服务连接测试失败\n\n`
    errorDetails += `🔧 服务信息:\n`
    errorDetails += `• 服务商: ${configForm.ai.provider.toUpperCase()}\n`
    errorDetails += `• 模型: ${configForm.ai.model}\n\n`
    errorDetails += `❌ 错误信息:\n${errorMsg}\n\n`
    
    if (errorData) {
      if (errorData.testMessage) {
        errorDetails += `💬 测试消息: ${errorData.testMessage}\n`
      }
      if (errorData.error) {
        errorDetails += `🐛 详细错误: ${errorData.error}\n`
      }
    }
    
    errorDetails += `\n💡 解决建议:\n`
    
    if (errorMsg.includes('401') || errorMsg.includes('Authentication')) {
      errorDetails += `• 检查API密钥是否正确\n`
      errorDetails += `• 确认API密钥未过期\n`
      errorDetails += `• 检查账户余额是否充足\n`
    } else if (errorMsg.includes('network') || errorMsg.includes('timeout')) {
      errorDetails += `• 检查网络连接\n`
      errorDetails += `• 确认防火墙设置\n`
      errorDetails += `• 尝试配置代理URL\n`
    } else {
      errorDetails += `• 检查所有配置项是否填写正确\n`
      errorDetails += `• 查看系统日志获取详细信息\n`
      errorDetails += `• 确认服务商API是否正常\n`
    }
    
    await ElMessageBox.alert(
      errorDetails,
      'AI服务连接测试失败',
      {
        confirmButtonText: '确定',
        type: 'error',
        dangerouslyUseHTMLString: false
      }
    )
    
    ElMessage.error(`${configForm.ai.provider.toUpperCase()} 服务连接测试失败`)
  } finally {
    testing.value = false
  }
}

// 测试小米设备连接
const testMiConnection = async () => {
  testing.value = true
  try {
    const { configAPI } = await import('../api')
    
    // 首先保存当前配置
    await configStore.updateConfig(configForm)
    
    // 显示测试进度
    ElMessage.info('正在测试小米设备连接...')
    
    // 测试小米设备连接
    const miResult = await configAPI.testMiConnection()
    
    if (miResult.success) {
      // 显示详细的成功信息
      const { userID, deviceID, autoStart, speakerStatus } = miResult.data
      
      let successDetails = `🎉 小米设备连接测试成功！\n\n`
      successDetails += `📱 设备信息:\n`
      successDetails += `• 用户账号: ${userID}\n`
      successDetails += `• 设备ID: ${deviceID}\n`
      successDetails += `• 连接状态: 正常\n\n`
      
      if (autoStart) {
        successDetails += `🚀 音箱服务状态:\n`
        successDetails += `• 自动启动: 成功\n`
        successDetails += `• 服务状态: 运行中\n\n`
      }
      
      successDetails += `✅ 第三方库 xiaoai-tts 工作正常\n`
      successDetails += `🔧 可以在音箱控制页面进行TTS测试`
      
      await ElMessageBox.alert(
        successDetails,
        '小米设备连接测试结果',
        {
          confirmButtonText: '确定',
          type: 'success',
          dangerouslyUseHTMLString: false
        }
      )
      
      ElMessage.success('小米设备连接测试成功')
    } else {
      throw new Error(miResult.message || '测试失败')
    }
  } catch (error) {
    // 显示详细的错误信息
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    const errorData = error.response?.data?.data
    
    let errorDetails = `❌ 小米设备连接测试失败\n\n`
    errorDetails += `📱 设备信息:\n`
    errorDetails += `• 用户账号: ${configForm.mi.userID}\n`
    errorDetails += `• 设备ID: ${configForm.mi.deviceID}\n\n`
    errorDetails += `❌ 错误信息:\n${errorMsg}\n\n`
    
    if (errorData?.error) {
      errorDetails += `🐛 详细错误: ${errorData.error}\n\n`
    }
    
    errorDetails += `💡 解决建议:\n`
    
    if (errorMsg.includes('登录失败') || errorMsg.includes('Authentication')) {
      errorDetails += `• 检查小米账号用户名和密码是否正确\n`
      errorDetails += `• 确认没有开启二步验证\n`
      errorDetails += `• 尝试在小米官方APP中先登录一次\n`
    } else if (errorMsg.includes('json') || errorMsg.includes('unmarshal')) {
      errorDetails += `• 这可能是小米服务器临时问题\n`
      errorDetails += `• 请稍后重试\n`
      errorDetails += `• 检查网络连接是否稳定\n`
    } else if (errorMsg.includes('设备')) {
      errorDetails += `• 检查设备ID是否正确\n`
      errorDetails += `• 确认设备在同一网络中\n`
      errorDetails += `• 在小米官方APP中确认设备状态\n`
    } else {
      errorDetails += `• 检查网络连接\n`
      errorDetails += `• 确认小米账号状态正常\n`
      errorDetails += `• 查看系统日志获取详细信息\n`
    }
    
    await ElMessageBox.alert(
      errorDetails,
      '小米设备连接测试失败',
      {
        confirmButtonText: '确定',
        type: 'error',
        dangerouslyUseHTMLString: false
      }
    )
    
    ElMessage.error('小米设备连接测试失败')
  } finally {
    testing.value = false
  }
}

// 测试所有连接
const testConnections = async () => {
  testing.value = true
  try {
    const { configAPI } = await import('../api')
    
    // 首先保存当前配置
    await configStore.updateConfig(configForm)
    
    let successCount = 0
    let totalTests = 2
    let testResults = []
    
    ElMessage.info('正在测试所有服务连接...')
    
    // 测试AI连接
    try {
      const aiResult = await configAPI.testAIConnection()
      if (aiResult.success) {
        testResults.push({
          service: 'AI服务',
          status: 'success',
          provider: aiResult.data.provider.toUpperCase(),
          details: `模型: ${aiResult.data.model}`
        })
        successCount++
      }
    } catch (error) {
      testResults.push({
        service: 'AI服务',
        status: 'failed',
        provider: configForm.ai.provider.toUpperCase(),
        details: error.response?.data?.message || error.message || '连接失败'
      })
    }
    
    // 测试小米设备连接
    try {
      const miResult = await configAPI.testMiConnection()
      if (miResult.success) {
        testResults.push({
          service: '小米设备',
          status: 'success',
          provider: 'xiaoai-tts',
          details: `设备: ${miResult.data.deviceID}`
        })
        successCount++
      }
    } catch (error) {
      testResults.push({
        service: '小米设备',
        status: 'failed',
        provider: 'xiaoai-tts',
        details: error.response?.data?.message || error.message || '连接失败'
      })
    }
    
    // 显示详细的测试结果
    let resultDetails = `🧪 连接测试完成\n\n`
    resultDetails += `📊 测试结果: ${successCount}/${totalTests} 服务连接成功\n\n`
    
    testResults.forEach((result, index) => {
      const icon = result.status === 'success' ? '✅' : '❌'
      const status = result.status === 'success' ? '成功' : '失败'
      
      resultDetails += `${icon} ${result.service} (${result.provider})\n`
      resultDetails += `   状态: ${status}\n`
      resultDetails += `   详情: ${result.details}\n\n`
    })
    
    if (successCount === totalTests) {
      resultDetails += `🎉 所有服务都已就绪，可以启动音箱服务！`
      
      await ElMessageBox.alert(
        resultDetails,
        '全部连接测试结果',
        {
          confirmButtonText: '确定',
          type: 'success',
          dangerouslyUseHTMLString: false
        }
      )
      
      ElMessage.success(`所有服务连接测试成功 (${successCount}/${totalTests})`)
    } else {
      resultDetails += `⚠️ 请解决失败的服务连接问题后再启动音箱服务`
      
      await ElMessageBox.alert(
        resultDetails,
        '连接测试结果',
        {
          confirmButtonText: '确定',
          type: 'warning',
          dangerouslyUseHTMLString: false
        }
      )
      
      ElMessage.warning(`部分服务连接成功 (${successCount}/${totalTests})`)
    }
    
  } catch (error) {
    ElMessage.error('连接测试失败')
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.config-page {
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.text-center {
  text-align: center;
}

.text-muted {
  color: #909399;
}

.config-tabs {
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

.form-help {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

:deep(.el-tab-pane) {
  padding: 0 20px;
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-input-number) {
  width: 100%;
}
</style>