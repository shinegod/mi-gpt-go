import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../components/Layout.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'Monitor' }
      },
      {
        path: '/config',
        name: 'Config',
        component: () => import('../views/Config.vue'),
        meta: { title: '配置管理', icon: 'Setting' }
      },
      {
        path: '/speaker',
        name: 'Speaker',
        component: () => import('../views/Speaker.vue'),
        meta: { title: '音箱控制', icon: 'Microphone' }
      },
      {
        path: '/concurrent',
        name: 'Concurrent',
        component: () => import('../views/Concurrent.vue'),
        meta: { title: '并发处理', icon: 'Operation' }
      },
      {
        path: '/logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { title: '系统日志', icon: 'Document' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router 