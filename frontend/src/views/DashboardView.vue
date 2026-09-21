<template>
  <div v-loading="loading">
    <el-row :gutter="16" class="stats">
      <el-col :span="4" :xs="12" v-for="item in statCards" :key="item.label">
        <el-card shadow="never">
          <div class="stat-label">{{ item.label }}</div>
          <div class="stat-value">{{ item.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top:16px">
      <template #header>最近更新</template>
      <el-table :data="recentIssues" stripe>
        <el-table-column prop="id" label="#" width="60" />
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'bug' ? 'danger' : 'primary'" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="240" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="priority" label="优先级" width="100" />
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { getDashboard } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const loading = ref(false)
const data = ref({ stats: {}, recent_issues: [] })

const statCards = computed(() => {
  const s = data.value.stats || {}
  return [
    { label: '需求', value: s.requirements ?? 0 },
    { label: '任务', value: s.tasks ?? 0 },
    { label: 'Bug', value: s.bugs ?? 0 },
    { label: '发布', value: s.releases ?? 0 },
    { label: '我的待办', value: s.my_tasks ?? 0 },
  ]
})

const recentIssues = computed(() => data.value.recent_issues || [])

function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString('zh-CN')
}

async function load() {
  loading.value = true
  try {
    const params = {}
    if (projectStore.currentIdOrDefault) {
      params.project_id = projectStore.currentIdOrDefault
    }
    data.value = await getDashboard(params)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => projectStore.currentId, load)
</script>

<style scoped>
.stat-label { color: #909399; font-size: 13px; }
.stat-value { font-size: 28px; font-weight: 700; margin-top: 8px; color: #303133; }
</style>
