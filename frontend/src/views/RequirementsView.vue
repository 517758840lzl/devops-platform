<template>
  <div>
    <div class="page-header">
      <h2>需求管理 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <el-button type="primary" @click="openCreate">新建需求</el-button>
    </div>
    <el-table v-loading="loading" :data="items" stripe>
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="title" label="标题" min-width="220" />
      <el-table-column prop="source" label="来源" width="100" />
      <el-table-column prop="priority" label="优先级" width="100" />
      <el-table-column prop="status" label="状态" width="120" />
      <el-table-column prop="updated_at" label="更新时间" width="180">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="新建需求" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="所属项目">
          <el-input :model-value="projectStore.current?.name" disabled />
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="来源"><el-input v-model="form.source" /></el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="form.priority">
            <el-option label="高" value="high" /><el-option label="中" value="medium" /><el-option label="低" value="low" />
          </el-select>
        </el-form-item>
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
import { createRequirement, listRequirements } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const loading = ref(false)
const items = ref([])
const dialogVisible = ref(false)
const form = reactive({ title: '', description: '', source: '产品', priority: 'medium' })

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

async function load() {
  loading.value = true
  try {
    items.value = await listRequirements({ project_id: projectStore.currentIdOrDefault })
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { title: '', description: '', source: '产品', priority: 'medium' })
  dialogVisible.value = true
}

async function submitCreate() {
  await createRequirement({ ...form, project_id: projectStore.currentIdOrDefault })
  dialogVisible.value = false
  await load()
}

onMounted(load)
</script>
