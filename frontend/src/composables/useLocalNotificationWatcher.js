import { onMounted, onUnmounted, watch } from 'vue'
import { listBuilds, listIssues, listReleases } from '../api'
import { useProjectPermission } from './useProjectPermission'
import { useAuthStore } from '../stores/auth'
import { useLocalNotificationStore } from '../stores/localNotifications'
import { useProjectStore } from '../stores/project'

const POLL_MS = 5 * 60 * 1000

export function useLocalNotificationWatcher() {
  const auth = useAuthStore()
  const projectStore = useProjectStore()
  const notif = useLocalNotificationStore()
  const { canApprove, refresh: refreshPermission } = useProjectPermission()
  let timer = null

  function bootKey(projectId) {
    return `_boot_${projectId}`
  }

  async function snapshotProject(projectId) {
    if (!projectId) return
    const userId = auth.user?.id
    const state = { ...notif.watchState }

    const builds = await listBuilds({ project_id: projectId })
    for (const b of builds.slice(0, 20)) {
      state[`build:${b.id}`] = buildSig(b)
    }

    const bugs = await listIssues({ type: 'bug', project_id: projectId })
    for (const bug of bugs) {
      if (bug.assignee_id === userId) {
        state[`bug:${bug.id}`] = `${bug.assignee_id}:${bug.updated_at}:${bug.status}`
      }
    }

    const releases = await listReleases({ project_id: projectId })
    for (const r of releases) {
      state[`release:${r.id}`] = r.status
    }

    state[bootKey(projectId)] = true
    notif.watchState = state
    notif.persistWatch()
  }

  function buildSig(b) {
    return `${b.status}:${b.artifacts_deleted ? 1 : 0}`
  }

  async function checkBuilds(projectId) {
    const builds = await listBuilds({ project_id: projectId })
    for (const b of builds.slice(0, 20)) {
      const watchKey = `build:${b.id}`
      const sig = buildSig(b)
      const prev = notif.watchState[watchKey]
      if (prev && prev !== sig) {
        const [prevStatus = '', prevDeleted = '0'] = String(prev).split(':')
        // 仅当构建状态真正变化为成功/失败时通知，避免删除产物触发「构建成功」
        if (prevStatus !== b.status && (b.status === 'success' || b.status === 'failed')) {
          notif.add({
            key: `build:${b.id}:${b.status}`,
            type: 'build',
            title: b.status === 'success' ? '构建成功' : '构建失败',
            message: `构建 #${b.build_number || b.id} · ${b.branch}${b.commit_sha ? ` · ${b.commit_sha.slice(0, 8)}` : ''}`,
            link: '/builds',
          })
        }
        if (prevDeleted !== '1' && b.artifacts_deleted) {
          notif.add({
            key: `build:${b.id}:artifacts_deleted`,
            type: 'build',
            title: '产物已删除',
            message: `构建 #${b.build_number || b.id} · ${b.branch}${b.commit_sha ? ` · ${b.commit_sha.slice(0, 8)}` : ''}`,
            link: '/builds',
          })
        }
      }
      notif.watchState[watchKey] = sig
    }
  }

  async function checkBugs(projectId) {
    const userId = auth.user?.id
    if (!userId) return

    const bugs = await listIssues({ type: 'bug', project_id: projectId })
    for (const bug of bugs) {
      if (bug.assignee_id !== userId) continue
      const watchKey = `bug:${bug.id}`
      const sig = `${bug.assignee_id}:${bug.updated_at}:${bug.status}`
      const prev = notif.watchState[watchKey]
      if (prev && prev !== sig) {
        notif.add({
          key: `bug:${bug.id}:${bug.updated_at}`,
          type: 'bug',
          title: '指派给你的 Bug 有更新',
          message: `#${bug.id} ${bug.title}`,
          link: '/bugs',
        })
      }
      notif.watchState[watchKey] = sig
    }
  }

  async function checkReleases(projectId) {
    if (!canApprove.value) return

    const releases = await listReleases({ project_id: projectId })
    for (const r of releases) {
      const watchKey = `release:${r.id}`
      const prev = notif.watchState[watchKey]
      if (r.status === 'pending_approval' && prev && prev !== 'pending_approval') {
        notif.add({
          key: `release:${r.id}:pending`,
          type: 'release',
          title: '发布待审批',
          message: `${r.version} ${r.title}`,
          link: '/releases',
        })
      }
      notif.watchState[watchKey] = r.status
    }
  }

  async function poll() {
    if (!auth.user) await auth.fetchMe()
    if (!auth.user) return

    const projectId = projectStore.currentIdOrDefault
    if (!projectId) return

    await refreshPermission()

    if (!notif.watchState[bootKey(projectId)]) {
      await snapshotProject(projectId)
      return
    }

    await checkBuilds(projectId)
    await checkBugs(projectId)
    await checkReleases(projectId)
    notif.persistWatch()
  }

  function start() {
    stop()
    poll()
    timer = setInterval(poll, POLL_MS)
  }

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(start)
  onUnmounted(stop)

  watch(() => projectStore.currentId, async (id) => {
    if (id && !notif.watchState[bootKey(id)]) {
      await snapshotProject(id)
    }
  })

  return { poll }
}
