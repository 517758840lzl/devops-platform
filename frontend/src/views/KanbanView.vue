<template>
  <div>
    <div class="page-header">
      <h2>迭代看板 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <div class="toolbar">
        <el-select v-model="iterationId" placeholder="选择迭代" style="width:220px" @change="loadKanban">
          <el-option v-for="it in iterations" :key="it.id" :label="`${it.name} (${it.version})`" :value="it.id" />
        </el-select>
        <el-button type="primary" @click="openCreate">新建任务</el-button>
      </div>
    </div>

    <div v-loading="loading" class="kanban">
      <div
        v-for="col in columns"
        :key="col.key"
        class="column"
        @dragover.prevent
        @drop="onDrop(col.key)"
      >
        <div class="column-header">
          <span>{{ col.title }}</span>
          <el-tag size="small" type="info">{{ col.issues?.length || 0 }}</el-tag>
        </div>
        <div
          v-for="issue in col.issues"
          :key="issue.id"
          class="card"
          draggable="true"
          @dragstart="onDragStart(issue)"
          @click="openDetail(issue)"
        >
          <div class="card-title">#{{ issue.id }} {{ issue.title }}</div>
          <div class="card-meta">
            <el-tag size="small" :type="priorityType(issue.priority)">{{ issue.priority }}</el-tag>
          </div>
        </div>
      </div>
    </div>

    <el-drawer v-model="detailVisible" title="任务详情" size="480px">
      <template v-if="currentIssue">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="标题">{{ currentIssue.title }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentIssue.status }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ currentIssue.priority }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ currentIssue.description || '-' }}</el-descriptions-item>
        </el-descriptions>
        <CommentActivityPanel target-type="issue" :target-id="currentIssue.id" />
      </template>
    </el-drawer>

    <el-dialog v-model="createVisible" title="新建任务" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="所属项目">
          <el-input :model-value="projectStore.current?.name" disabled />
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import CommentActivityPanel from '../components/CommentActivityPanel.vue'
import { createIssue, getIterationKanban, listIterations, patchIssueStatus } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const loading = ref(false)
const iterations = ref([])
const iterationId = ref(null)
const columns = ref([])
const dragging = ref(null)
const detailVisible = ref(false)
const currentIssue = ref(null)
const createVisible = ref(false)
const form = reactive({ type: 'task', title: '', description: '', status: 'todo' })

function priorityType(p) {
  if (p === 'high') return 'danger'
  if (p === 'low') return 'info'
  return 'warning'
}

async function loadIterations() {
  iterations.value = await listIterations({ project_id: projectStore.currentIdOrDefault })
  if (iterations.value.length && !iterationId.value) {
    iterationId.value = iterations.value[0].id
    await loadKanban()
  }
}

async function loadKanban() {
  if (!iterationId.value) return
  loading.value = true
  try {
    const data = await getIterationKanban(iterationId.value)
    columns.value = data.columns || []
  } finally {
    loading.value = false
  }
}

function onDragStart(issue) {
  dragging.value = issue
}

async function onDrop(status) {
  if (!dragging.value || dragging.value.status === status) return
  await patchIssueStatus(dragging.value.id, status)
  dragging.value = null
  await loadKanban()
}

function openDetail(issue) {
  currentIssue.value = issue
  detailVisible.value = true
}

function openCreate() {
  Object.assign(form, { type: 'task', title: '', description: '', status: 'todo', iteration_id: iterationId.value })
  createVisible.value = true
}

async function submitCreate() {
  await createIssue({ ...form, project_id: projectStore.currentIdOrDefault, iteration_id: iterationId.value })
  createVisible.value = false
  await loadKanban()
}

onMounted(loadIterations)
</script>

<style scoped>
.toolbar { display: flex; gap: 12px; align-items: center; }
.kanban { display: flex; gap: 16px; overflow-x: auto; min-height: 480px; }
.column {
  flex: 1; min-width: 240px; background: #f0f2f5; border-radius: 8px; padding: 12px;
}
.column-header {
  display: flex; justify-content: space-between; align-items: center;
  font-weight: 600; margin-bottom: 12px; color: #303133;
}
.card {
  background: #fff; border-radius: 6px; padding: 10px 12px; margin-bottom: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08); cursor: grab;
}
.card-title { font-size: 14px; margin-bottom: 8px; }
.card-meta { display: flex; gap: 6px; }
</style>
