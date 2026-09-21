import { defineStore } from 'pinia'
import { listProjects } from '../api'

export const useProjectStore = defineStore('project', {
  state: () => ({
    projects: [],
    currentId: Number(localStorage.getItem('currentProjectId')) || 0,
  }),
  getters: {
    current: (s) => s.projects.find((p) => p.id === s.currentId) || s.projects[0] || null,
    currentIdOrDefault: (s) => s.currentId || s.projects[0]?.id || 0,
  },
  actions: {
    async fetchProjects() {
      this.projects = await listProjects()
      if (!this.currentId && this.projects.length) {
        this.setCurrent(this.projects[0].id)
      } else if (this.currentId && !this.projects.some((p) => p.id === this.currentId)) {
        this.setCurrent(this.projects[0]?.id || 0)
      }
    },
    setCurrent(id) {
      this.currentId = id
      localStorage.setItem('currentProjectId', String(id))
    },
  },
})
