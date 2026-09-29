import { defineStore } from 'pinia'
import { listNotifications, markAllNotificationsRead, markNotificationRead } from '../api'

function normalize(row) {
  return {
    id: row.id,
    key: row.event_key || `n:${row.id}`,
    type: row.type,
    title: row.title,
    message: row.message,
    link: row.link,
    read: !!row.read,
    createdAt: row.created_at,
    projectId: row.project_id,
  }
}

export const useLocalNotificationStore = defineStore('localNotifications', {
  state: () => ({
    items: [],
    projectId: null,
    loading: false,
  }),
  getters: {
    unreadCount: (s) => s.items.filter((i) => !i.read).length,
    recentItems: (s) => s.items.slice(0, 30),
  },
  actions: {
    async load(projectId) {
      this.projectId = projectId || null
      if (!projectId) {
        this.items = []
        return
      }
      this.loading = true
      try {
        const rows = await listNotifications(projectId)
        this.items = (rows || []).map(normalize)
      } catch {
        this.items = []
      } finally {
        this.loading = false
      }
    },
    pushRemote(row) {
      const item = normalize(row)
      if (this.projectId && item.projectId && Number(item.projectId) !== Number(this.projectId)) {
        return
      }
      if (this.items.some((i) => i.id === item.id || i.key === item.key)) return
      this.items.unshift(item)
    },
    async markRead(id) {
      const item = this.items.find((i) => i.id === id)
      if (item) item.read = true
      try {
        await markNotificationRead(id)
      } catch {
        /* keep optimistic */
      }
    },
    async markAllRead() {
      this.items.forEach((i) => { i.read = true })
      if (!this.projectId) return
      try {
        await markAllNotificationsRead(this.projectId)
      } catch {
        /* keep optimistic */
      }
    },
    async clearAll() {
      await this.markAllRead()
    },
  },
})
