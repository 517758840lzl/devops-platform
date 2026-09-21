<template>
  <div>
    <div class="page-header">
      <h2>任务列表 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <el-button type="primary" @click="openCreate">新建任务</el-button>
    </div>
    <el-table v-loading="loading" :data="items" stripe @row-click="openDetail">
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="title" label="标题" min-width="220" />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-select v-model="row.status" size="small" @click.stop @change="(v) => onStatusChange(row, v)">
            <el-option v-for="s in taskStatuses" :key="s" :label="s" :value="s" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="priority" label="优先级" width="100" />
      <el-table-column prop="updated_at" label="更新时间" width="180">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="detailVisible" title="任务详情" size="480px">
      <template v-if="current">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="标题">{{ current.title }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ current.status }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ current.description || '-' }}</el-descriptions-item>
        </el-descriptions>
        <CommentActivityPanel target-type="issue" :target-id="current.id" />
      </template>
    </el-drawer>

    <el-dialog v-model="dialogVisible" title="新建任务" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="所属项目">
          <el-input :model-value="projectStore.current?.name" disabled />
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import CommentActivityPanel from '../components/CommentActivityPanel.vue'
import { createIssue, listIssues, updateIssue } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const loading = ref(false)
const items = ref([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const current = ref(null)
const taskStatuses = ['todo', 'in_progress', 'review', 'done']
const form = reactive({ type: 'task', title: '', description: '', priority: 'medium' })

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

async function load() {
  loading.value = true
  try {
    items.value = await listIssues({ type: 'task', project_id: projectStore.currentIdOrDefault })
  } finally {
    loading.value = false
  }
}

function openDetail(row) {
  current.value = row
  detailVisible.value = true
}

function openCreate() {
  Object.assign(form, { type: 'task', title: '', description: '', priority: 'medium' })
  dialogVisible.value = true
}

async function submitCreate() {
  await createIssue({ ...form, project_id: projectStore.currentIdOrDefault })
  dialogVisible.value = false
  await load()
}

async function onStatusChange(row, status) {
  await updateIssue(row.id, { ...row, status })
}

onMounted(load)
</script>
