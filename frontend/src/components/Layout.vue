<template>
  <div class="layout-container">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <el-icon size="32" color="#409EFF">
            <Monitor />
          </el-icon>
          <h3>MiGPT Go</h3>
        </div>
        
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <el-menu-item
            v-for="route in menuRoutes"
            :key="route.path"
            :index="route.path"
          >
            <el-icon>
              <component :is="iconMap[route.meta.icon] || Monitor" />
            </el-icon>
            <span>{{ route.meta.title }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主要内容区域 -->
      <el-container>
        <!-- 顶部导航 -->
        <el-header class="header">
          <div class="header-left">
            <h2>{{ currentTitle }}</h2>
          </div>
          
          <div class="header-right">
            <el-badge :value="systemStore.status.isRunning ? '运行中' : '已停止'" :type="systemStore.status.isRunning ? 'success' : 'danger'">
              <el-button size="small" @click="refreshStatus">
                <el-icon><Refresh /></el-icon>
                刷新状态
              </el-button>
            </el-badge>
          </div>
        </el-header>

        <!-- 内容区域 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSystemStore } from '../stores'
import { 
  Monitor, 
  Refresh, 
  Setting, 
  Microphone, 
  Operation, 
  Document 
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const systemStore = useSystemStore()

// 图标映射
const iconMap = {
  Monitor,
  Setting,
  Microphone,
  Operation,
  Document
}

// 菜单路由
const menuRoutes = computed(() => {
  const mainRoute = router.getRoutes().find(r => r.path === '/')
  return mainRoute?.children?.filter(child => child.meta?.title) || []
})

// 当前页面标题
const currentTitle = computed(() => {
  const currentRoute = menuRoutes.value.find(r => r.path === route.path)
  return currentRoute?.meta?.title || 'MiGPT Go 管理面板'
})

// 刷新系统状态
const refreshStatus = async () => {
  await systemStore.fetchStatus()
}

onMounted(() => {
  refreshStatus()
  // 每30秒自动刷新状态
  setInterval(refreshStatus, 30000)
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: white;
  border-bottom: 1px solid #434a50;
}

.logo h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.sidebar-menu {
  border-right: none;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-left h2 {
  margin: 0;
  color: #303133;
  font-size: 20px;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.main-content {
  background-color: #f5f5f5;
  padding: 20px;
  overflow-y: auto;
}

:deep(.el-badge__content) {
  font-size: 10px;
  padding: 0 4px;
}
</style> 