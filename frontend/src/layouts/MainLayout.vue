<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">
        <img :src="appLogoUrl" alt="" class="logo-mark" width="28" height="28" />
        <span class="logo-text">{{ APP_NAME }}</span>
      </div>
      <el-menu :default-active="route.path" router background-color="#1d2b3a" text-color="#bfcbd9" active-text-color="#409eff">
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <span>{{ route.meta.title || '工作台' }}</span>
          <el-select
            v-if="projectStore.projects.length"
            v-model="projectStore.currentId"
            class="project-select"
            placeholder="选择项目"
            @change="onProjectChange"
          >
            <el-option
              v-for="p in projectStore.projects"
              :key="p.id"
              :label="`${p.name} (${p.code})`"
              :value="p.id"
            />
          </el-select>
        </div>
        <div class="user">
          <LocalNotificationBell />
          <span>{{ auth.user?.name || auth.user?.username }}</span>
          <el-button link type="primary" @click="onLogout">退出</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view :key="projectStore.currentId" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LocalNotificationBell from '../components/LocalNotificationBell.vue'
import { useLocalNotificationWatcher } from '../composables/useLocalNotificationWatcher'
import { useProjectPermission } from '../composables/useProjectPermission'
import { useAuthStore } from '../stores/auth'
import { useProjectStore } from '../stores/project'
import { APP_NAME, appLogoUrl } from '../config/brand'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const projectStore = useProjectStore()
const { canManage, loaded: permLoaded, refresh: refreshPermission } = useProjectPermission()

useLocalNotificationWatcher()

const allMenus = [
  { path: '/dashboard', title: '工作台', icon: 'Odometer' },
  { path: '/projects', title: '项目', icon: 'Folder' },
  { path: '/requirements', title: '需求', icon: 'Document' },
  { path: '/bugs', title: 'Bug', icon: 'Warning' },
  { path: '/releases', title: '发布', icon: 'Upload' },
  { path: '/builds', title: '打包构建', icon: 'Box' },
  { path: '/progress', title: '项目进度', icon: 'TrendCharts' },
  { path: '/members', title: '成员权限', icon: 'User', adminOnly: true },
  { path: '/activities', title: '操作日志', icon: 'Notebook', adminOnly: true },
  { path: '/config', title: '项目配置', icon: 'Setting' },
]

const menus = computed(() =>
  allMenus.filter((item) => !item.adminOnly || canManage.value),
)

function guardAdminRoutes() {
  if (!permLoaded.value) return
  if ((route.path === '/members' || route.path === '/activities') && !canManage.value) {
    router.replace('/dashboard')
  }
}

onMounted(async () => {
  await auth.fetchMe()
  await projectStore.fetchProjects()
  await refreshPermission()
  guardAdminRoutes()
})

function onProjectChange(id) {
  projectStore.setCurrent(id)
}

watch(() => projectStore.currentId, async () => {
  await refreshPermission()
  guardAdminRoutes()
})

watch(() => route.path, guardAdminRoutes)
watch(canManage, guardAdminRoutes)

function onLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout { height: 100vh; }
.aside { background: #1d2b3a; }
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 0 12px;
  color: #fff;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.logo-mark {
  flex-shrink: 0;
  border-radius: 6px;
}
.logo-text {
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.02em;
  line-height: 1.2;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
}
.header-left { display: flex; align-items: center; gap: 16px; }
.project-select { width: 220px; }
.user { display: flex; align-items: center; gap: 12px; }
.main { padding: 20px; }
</style>
