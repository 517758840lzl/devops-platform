import { onMounted, onUnmounted, watch } from 'vue'
import { apiBaseForSSE } from '../utils/sse'
import { useAuthStore } from '../stores/auth'
import { useLocalNotificationStore } from '../stores/localNotifications'
import { useProjectStore } from '../stores/project'

export function useLocalNotificationWatcher() {
  const auth = useAuthStore()
  const projectStore = useProjectStore()
  const notif = useLocalNotificationStore()
  let eventSource = null

  function eventsURL(projectId, token) {
    const qs = new URLSearchParams({
      project_id: String(projectId),
      token,
    })
    return `${apiBaseForSSE()}/api/notifications/events?${qs}`
  }

  function disconnect() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  async function connect() {
    disconnect()
    if (!auth.user) {
      try {
        await auth.fetchMe()
      } catch {
        return
      }
    }
    const projectId = projectStore.currentIdOrDefault
    const token = localStorage.getItem('token')
    if (!projectId || !token) {
      notif.items = []
      return
    }
    await notif.load(projectId)
    const es = new EventSource(eventsURL(projectId, token))
    eventSource = es
    es.addEventListener('notification', (e) => {
      try {
        notif.pushRemote(JSON.parse(e.data))
      } catch {
        /* ignore */
      }
    })
    es.onerror = () => {
      // 无项目权限时后端 403，停掉避免空转重连
      if (es.readyState === EventSource.CLOSED) {
        disconnect()
      }
    }
  }

  onMounted(connect)
  onUnmounted(disconnect)
  watch(() => projectStore.currentId, () => connect())

  return { reconnect: connect }
}
