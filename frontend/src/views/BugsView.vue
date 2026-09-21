<template>
  <div>
    <div class="page-header">
      <h2>Bug 管理 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <el-button v-if="canWrite" type="primary" @click="openCreate">提交 Bug</el-button>
    </div>
    <el-alert v-if="isViewer" type="info" show-icon :closable="false" style="margin-bottom:12px">
      当前为只读权限，可查看 Bug 详情，不可提交或修改
    </el-alert>
    <el-table v-loading="loading" :data="items" stripe @row-click="openDetail">
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="title" label="标题" min-width="180" />
      <el-table-column label="端" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.bug_side" size="small" :type="row.bug_side === 'frontend' ? 'primary' : 'warning'">
            {{ bugSideLabel(row.bug_side) }}
          </el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="severity" label="严重度" width="90" />
      <el-table-column label="指派人" width="100">
        <template #default="{ row }">{{ memberName(row.assignee_id) }}</template>
      </el-table-column>
      <el-table-column label="关联版本" width="100">
        <template #default="{ row }">{{ displayVersion(row) }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="140">
        <template #default="{ row }">
          <el-select
            v-model="row.status"
            size="small"
            :disabled="!canWrite"
            @click.stop
            @change="(v) => onStatusChange(row, v)"
          >
            <el-option v-for="s in bugStatuses" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="环境" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.environment" size="small" :type="row.environment === 'prod' ? 'danger' : 'info'">
            {{ bugEnvironmentLabel(row.environment) }}
          </el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="170">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="detailVisible" title="Bug 详情" size="560px" class="bug-detail-drawer">
      <template v-if="current">
        <div class="bug-detail-body">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="标题">{{ current.title }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-select
              v-model="current.status"
              style="width: 160px"
              :disabled="!canWrite"
              @change="(v) => onStatusChange(current, v)"
            >
              <el-option v-for="s in bugStatuses" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
          </el-descriptions-item>
          <el-descriptions-item label="端">
            <el-select
              v-if="canWrite"
              v-model="current.bug_side"
              style="width: 160px"
              placeholder="选择前端/后端（必填）"
              @change="saveDetailMeta"
            >
              <el-option v-for="s in bugSideOptions" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
            <span v-else>{{ bugSideLabel(current.bug_side) || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="指派人">
            <el-select
              v-if="canWrite"
              v-model="current.assignee_id"
              style="width: 200px"
              clearable
              filterable
              placeholder="选择指派人"
              @change="saveDetailMeta"
            >
              <el-option
                v-for="m in members"
                :key="m.user_id"
                :label="`${m.name} (${m.username})`"
                :value="m.user_id"
              />
            </el-select>
            <span v-else>{{ memberName(current.assignee_id) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="关联版本">
            <el-autocomplete
              v-if="canWrite"
              v-model="current.app_version"
              style="width: 200px"
              clearable
              :fetch-suggestions="queryVersionSuggestions"
              placeholder="输入或选择版本号"
              @change="saveDetailMeta"
              @select="saveDetailMeta"
            />
            <span v-else>{{ displayVersion(current) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="环境">
            <el-select
              v-if="canWrite"
              v-model="current.environment"
              style="width: 160px"
              placeholder="选择正式 / 测试"
              @change="saveDetailMeta"
            >
              <el-option v-for="e in bugEnvironmentOptions" :key="e.value" :label="e.label" :value="e.value" />
            </el-select>
            <span v-else>{{ bugEnvironmentLabel(current.environment) || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="严重度">{{ current.severity }}</el-descriptions-item>
          <el-descriptions-item label="复现步骤">{{ current.steps_to_reproduce || '-' }}</el-descriptions-item>
          <el-descriptions-item label="实际结果">{{ current.actual_result || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">截图 / 录屏</div>
        <AttachmentGallery ref="galleryRef" target-type="issue" :target-id="current.id" />
        <MediaUpload
          v-if="canWrite"
          target-type="issue"
          :target-id="current.id"
          auto-upload
          @uploaded="galleryRef?.reload()"
        />

        <CommentActivityPanel target-type="issue" :target-id="current.id" />
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="dialogVisible" title="提交 Bug" width="600px" @closed="onDialogClosed">
      <el-form ref="createFormRef" :model="form" :rules="createRules" label-width="90px">
        <el-form-item label="所属项目">
          <el-input :model-value="projectStore.current?.name" disabled />
        </el-form-item>
        <el-form-item label="标题" prop="title"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="端" prop="bug_side" required>
          <el-select v-model="form.bug_side" placeholder="请选择前端 / 后端" style="width: 160px">
            <el-option v-for="s in bugSideOptions" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派人">
          <el-select v-model="form.assignee_id" clearable filterable placeholder="可选">
            <el-option
              v-for="m in members"
              :key="m.user_id"
              :label="`${m.name} (${m.username})`"
              :value="m.user_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关联版本">
          <el-autocomplete
            v-model="form.app_version"
            clearable
            :fetch-suggestions="queryVersionSuggestions"
            placeholder="输入或选择版本号，如 1.2.0"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="环境" prop="environment" required>
          <el-select v-model="form.environment" placeholder="请选择正式 / 测试" style="width: 160px">
            <el-option v-for="e in bugEnvironmentOptions" :key="e.value" :label="e.label" :value="e.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="复现步骤"><el-input v-model="form.steps_to_reproduce" type="textarea" /></el-form-item>
        <el-form-item label="实际结果"><el-input v-model="form.actual_result" type="textarea" /></el-form-item>
        <el-form-item label="严重度">
          <el-select v-model="form.severity">
            <el-option label="Critical" value="critical" />
            <el-option label="Major" value="major" />
            <el-option label="Minor" value="minor" />
          </el-select>
        </el-form-item>
        <el-form-item label="附件">
          <MediaUpload ref="uploadRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import AttachmentGallery from '../components/AttachmentGallery.vue'
import CommentActivityPanel from '../components/CommentActivityPanel.vue'
import MediaUpload from '../components/MediaUpload.vue'
import { useProjectPermission } from '../composables/useProjectPermission'
import { createIssue, listIssues, listMembers, listReleases, patchIssueStatus, updateIssue } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const { canWrite, isViewer, refresh: refreshPermission } = useProjectPermission()
const loading = ref(false)
const submitting = ref(false)
const items = ref([])
const members = ref([])
const releases = ref([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const current = ref(null)
const uploadRef = ref(null)
const galleryRef = ref(null)
const createFormRef = ref(null)
const createRules = {
  title: [{ required: true, message: '请填写标题', trigger: 'blur' }],
  bug_side: [{ required: true, message: '请选择前端或后端', trigger: 'change' }],
  environment: [{ required: true, message: '请选择正式或测试环境', trigger: 'change' }],
}
const bugStatuses = [
  { value: 'open', label: '待处理' },
  { value: 'confirmed', label: '已确认' },
  { value: 'in_progress', label: '修复中' },
  { value: 'resolved', label: '已解决' },
  { value: 'closed', label: '已关闭' },
]
const bugSideOptions = [
  { value: 'frontend', label: '前端' },
  { value: 'backend', label: '后端' },
]
const bugEnvironmentOptions = [
  { value: 'prod', label: '正式' },
  { value: 'test', label: '测试' },
]
const form = reactive({
  type: 'bug',
  title: '',
  steps_to_reproduce: '',
  actual_result: '',
  environment: '',
  severity: 'major',
  bug_side: '',
  assignee_id: null,
  app_version: '',
})

/** 已有发布版本 + 已填过的 Bug 版本，仅作建议 */
const versionOptions = computed(() => {
  const set = new Set()
  for (const r of releases.value) {
    if (r.version) set.add(String(r.version).trim())
  }
  for (const i of items.value) {
    if (i.app_version) set.add(String(i.app_version).trim())
  }
  return [...set].filter(Boolean)
})

function queryVersionSuggestions(queryString, cb) {
  const q = (queryString || '').trim().toLowerCase()
  const list = versionOptions.value
    .filter((v) => !q || v.toLowerCase().includes(q))
    .map((value) => ({ value }))
  cb(list)
}

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function bugSideLabel(side) {
  return bugSideOptions.find((s) => s.value === side)?.label || side || '-'
}

function bugEnvironmentLabel(env) {
  return bugEnvironmentOptions.find((e) => e.value === env)?.label || env || '-'
}

function memberName(userId) {
  if (!userId) return '-'
  const m = members.value.find((x) => x.user_id === userId)
  return m ? m.name : `#${userId}`
}

function displayVersion(row) {
  if (!row) return '-'
  if (row.app_version) return row.app_version
  if (row.release_id) {
    const r = releases.value.find((x) => x.id === row.release_id)
    return r ? r.version : `#${row.release_id}`
  }
  return '-'
}

async function loadMeta() {
  const projectId = projectStore.currentIdOrDefault
  if (!projectId) {
    members.value = []
    releases.value = []
    return
  }
  ;[members.value, releases.value] = await Promise.all([
    listMembers(projectId),
    listReleases({ project_id: projectId }),
  ])
}

async function load() {
  loading.value = true
  try {
    await loadMeta()
    items.value = await listIssues({ type: 'bug', project_id: projectStore.currentIdOrDefault })
  } finally {
    loading.value = false
  }
}

function openDetail(row) {
  current.value = {
    ...row,
    app_version: row.app_version || (row.release_id
      ? (releases.value.find((x) => x.id === row.release_id)?.version || '')
      : ''),
  }
  detailVisible.value = true
}

function openCreate() {
  Object.assign(form, {
    type: 'bug',
    title: '',
    steps_to_reproduce: '',
    actual_result: '',
    environment: '',
    severity: 'major',
    bug_side: '',
    assignee_id: null,
    app_version: '',
  })
  dialogVisible.value = true
}

function onDialogClosed() {
  uploadRef.value?.reset()
}

async function onStatusChange(row, status) {
  if (!canWrite.value) return
  try {
    const updated = await patchIssueStatus(row.id, status)
    Object.assign(row, updated)
    if (current.value?.id === row.id) {
      current.value = { ...current.value, ...updated }
    }
    const listRow = items.value.find((i) => i.id === row.id)
    if (listRow) Object.assign(listRow, updated)
  } catch (err) {
    ElMessage.error(err.message || '状态更新失败')
    await load()
  }
}

function ensureBugSide(side) {
  if (side === 'frontend' || side === 'backend') return true
  ElMessage.warning('请选择前端或后端')
  return false
}

async function saveDetailMeta() {
  if (!current.value || !canWrite.value) return
  if (!ensureBugSide(current.value.bug_side)) {
    await load()
    return
  }
  try {
    const payload = {
      ...current.value,
      assignee_id: current.value.assignee_id || null,
      app_version: (current.value.app_version || '').trim(),
      bug_side: current.value.bug_side,
      environment: current.value.environment || '',
    }
    const updated = await updateIssue(current.value.id, payload)
    Object.assign(current.value, updated)
    const listRow = items.value.find((i) => i.id === current.value.id)
    if (listRow) Object.assign(listRow, updated)
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
    await load()
  }
}

async function submitCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const issue = await createIssue({
      ...form,
      project_id: projectStore.currentIdOrDefault,
      assignee_id: form.assignee_id || null,
      app_version: (form.app_version || '').trim(),
    })
    await uploadRef.value?.uploadPending('issue', issue.id)
    dialogVisible.value = false
    await load()
  } catch (err) {
    ElMessage.error(err.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await refreshPermission()
  await load()
})
watch(() => projectStore.currentId, async () => {
  await refreshPermission()
  await load()
})
</script>

<style scoped>
.section-title { font-weight: 600; margin: 20px 0 10px; font-size: 16px; color: #303133; }
.bug-detail-body :deep(.el-descriptions__label) {
  width: 100px;
  font-size: 15px;
  font-weight: 500;
}
.bug-detail-body :deep(.el-descriptions__content) {
  font-size: 15px;
  line-height: 1.6;
  color: #303133;
}
.bug-detail-body :deep(.el-descriptions__cell) {
  padding: 14px 16px;
}
</style>
