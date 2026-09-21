import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/DashboardView.vue'), meta: { title: '工作台' } },
      { path: 'projects', name: 'Projects', component: () => import('../views/ProjectsView.vue'), meta: { title: '项目' } },
      { path: 'requirements', name: 'Requirements', component: () => import('../views/RequirementsView.vue'), meta: { title: '需求' } },
      { path: 'kanban', name: 'Kanban', component: () => import('../views/KanbanView.vue'), meta: { title: '迭代看板' } },
      { path: 'tasks', name: 'Tasks', component: () => import('../views/TasksView.vue'), meta: { title: '任务' } },
      { path: 'members', name: 'Members', component: () => import('../views/MembersView.vue'), meta: { title: '成员权限', adminOnly: true } },
      { path: 'activities', name: 'Activities', component: () => import('../views/ActivitiesView.vue'), meta: { title: '操作日志', adminOnly: true } },
      { path: 'config', name: 'ProjectConfig', component: () => import('../views/ProjectConfigView.vue'), meta: { title: '项目配置' } },
      { path: 'bugs', name: 'Bugs', component: () => import('../views/BugsView.vue'), meta: { title: 'Bug' } },
      { path: 'releases', name: 'Releases', component: () => import('../views/ReleasesView.vue'), meta: { title: '发布' } },
      { path: 'store-assets', redirect: { path: '/config', query: { tab: 'store' } } },
      { path: 'builds', name: 'Builds', component: () => import('../views/BuildsView.vue'), meta: { title: '打包构建' } },
      { path: 'progress', name: 'Progress', component: () => import('../views/ProgressView.vue'), meta: { title: '项目进度' } },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.token) {
    return '/login'
  }
  if (to.path === '/login' && auth.token) {
    return '/dashboard'
  }
  if (auth.token && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      auth.logout()
      return '/login'
    }
  }
})

export default router
