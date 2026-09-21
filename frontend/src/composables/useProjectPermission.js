import { computed, ref } from 'vue'
import { listMembers } from '../api'
import { useAuthStore } from '../stores/auth'
import { useProjectStore } from '../stores/project'

const WRITE_ROLES = ['owner', 'super_admin', 'admin', 'developer', 'tester']

export function useProjectPermission() {
  const authStore = useAuthStore()
  const projectStore = useProjectStore()
  const role = ref('')
  const loaded = ref(false)

  async function refresh() {
    const projectId = projectStore.currentIdOrDefault
    if (!projectId) {
      role.value = ''
      loaded.value = true
      return
    }
    if (!authStore.user) await authStore.fetchMe()
    if (authStore.user?.role === 'admin') {
      role.value = 'owner'
      loaded.value = true
      return
    }

    // 优先用项目列表里的 my_role，避免成员列表 id 类型不一致导致误判为 viewer
    const fromList = projectStore.projects.find((p) => Number(p.id) === Number(projectId))
    if (fromList?.my_role) {
      role.value = fromList.my_role
      loaded.value = true
      return
    }

    try {
      const members = await listMembers(projectId)
      const uid = Number(authStore.user?.id)
      const me = members.find((m) => Number(m.user_id) === uid)
      role.value = me?.role || 'viewer'
    } catch {
      role.value = 'viewer'
    }
    loaded.value = true
  }

  const isViewer = computed(() => role.value === 'viewer')
  const isTester = computed(() => role.value === 'tester')
  // tester 可写 Bug；仅 viewer 只读
  const canWrite = computed(() => WRITE_ROLES.includes(role.value))
  const canEditProjectConfig = computed(() => ['owner', 'super_admin', 'admin', 'developer'].includes(role.value))
  const canManage = computed(() => ['owner', 'super_admin', 'admin'].includes(role.value))
  const canApprove = computed(() => canManage.value)
  const canDeleteStoreAsset = computed(() => canManage.value)

  return {
    role,
    loaded,
    refresh,
    isViewer,
    isTester,
    canWrite,
    canEditProjectConfig,
    canManage,
    canApprove,
    canDeleteStoreAsset,
  }
}
