import { defineStore } from 'pinia'
import { getMe, login as loginApi } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null,
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
  },
  actions: {
    async login(username, password) {
      const data = await loginApi({ username, password })
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
    },
    async fetchMe() {
      if (!this.token) return
      this.user = await getMe()
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('devops_local_notification_watch')
    },
  },
})
