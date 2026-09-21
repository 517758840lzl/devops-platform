import { defineStore } from 'pinia'

const STORAGE_KEY = 'devops_local_notifications'
const WATCH_KEY = 'devops_local_notification_watch'
const MAX_ITEMS = 80

function loadJSON(key, fallback) {
  try {
    return JSON.parse(localStorage.getItem(key) || '')
  } catch {
    return fallback
  }
}

export const useLocalNotificationStore = defineStore('localNotifications', {
  state: () => ({
    items: loadJSON(STORAGE_KEY, []),
    watchState: loadJSON(WATCH_KEY, {}),
  }),
  getters: {
    unreadCount: (s) => s.items.filter((i) => !i.read).length,
    recentItems: (s) => s.items.slice(0, 30),
  },
  actions: {
    persistItems() {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(this.items.slice(0, MAX_ITEMS)))
    },
    persistWatch() {
      localStorage.setItem(WATCH_KEY, JSON.stringify(this.watchState))
    },
    add({ key, type, title, message, link }) {
      if (this.items.some((i) => i.key === key)) return
      this.items.unshift({
        id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
        key,
        type,
        title,
        message,
        link,
        read: false,
        createdAt: new Date().toISOString(),
      })
      this.persistItems()
    },
    markRead(id) {
      const item = this.items.find((i) => i.id === id)
      if (item) item.read = true
      this.persistItems()
    },
    markAllRead() {
      this.items.forEach((i) => { i.read = true })
      this.persistItems()
    },
    remove(id) {
      this.items = this.items.filter((i) => i.id !== id)
      this.persistItems()
    },
    clearAll() {
      this.items = []
      this.persistItems()
    },
    resetWatch() {
      this.watchState = {}
      this.persistWatch()
    },
  },
})
