<template>
  <div>
    <div class="page-header">
      <h2>项目列表</h2>
      <el-button v-if="canCreateProject" type="primary" @click="openCreate">新建项目</el-button>
    </div>
    <el-table
      v-loading="loading"
      :data="items"
      stripe
      :row-class-name="rowClassName"
      @row-click="onRowClick"
    >
      <el-table-column prop="code" label="代号" width="100" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="git_url" label="Git 仓库" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">{{ row.git_url || '-' }}</template>
      </el-table-column>
      <el-table-column label="构建" width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.build_profile" class="build-cell">{{ row.build_profile }} / {{ row.build_platform || 'android' }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status || 'active' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="170">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '项目设置' : '新建项目'" width="640px" destroy-on-close>
      <el-form :model="form" label-width="130px" class="project-form">
        <el-form-item label="代号"><el-input v-model="form.code" :disabled="editing" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" @input="onNameInput" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>

        <template v-if="editing">
          <el-divider content-position="left">Git</el-divider>
          <el-form-item label="Git 仓库">
            <el-input
              v-model="form.git_url"
              placeholder="https://github.com/org/repo.git 或 git@github.com:org/repo.git"
            />
            <div class="hint">分支在「打包构建」时选择；支持 HTTPS / SSH</div>
          </el-form-item>

          <el-divider content-position="left">应用配置</el-divider>
          <el-form-item label="应用显示名">
            <el-input
              v-model="form.app_display_name"
              placeholder="AppStrings.appTitle"
              @input="nameSyncedDisplay = false"
            />
            <div class="hint">默认跟随项目名称；进度/推送/产物命名会用到</div>
          </el-form-item>
          <el-form-item label="App ID">
            <el-input v-model="form.app_id" placeholder="environment_config → acqChannel / 请求头 BridgeKey" />
            <div class="hint">如 Easy Money：<code>PrimeCreditLoan</code></div>
          </el-form-item>
          <el-form-item label="渠道序号">
            <el-input v-model="form.channel_index" placeholder="acqChannelIndex / 请求头 Position" />
            <div class="hint">通常为 <code>0</code>，对应 lib/core/config/environment_config.dart</div>
          </el-form-item>
          <el-form-item label="Android 包名">
            <el-input v-model="form.package_android" placeholder="android/app/build.gradle.kts → applicationId" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createProject,
  getProjectSettings,
  listProjects,
  saveProjectSettings,
  updateProject,
} from '../api'
import { useAuthStore } from '../stores/auth'
import { useProjectStore } from '../stores/project'
import { defaultArtifactPath } from '../utils/buildPresets'

const EDIT_ROLES = ['owner', 'super_admin', 'admin', 'developer']

const auth = useAuthStore()
const projectStore = useProjectStore()
const loading = ref(false)
const saving = ref(false)
const items = ref([])
const dialogVisible = ref(false)
const editing = ref(false)
const editId = ref(null)
const editingRow = ref(null)
const settingsSnapshot = ref(null)
const nameSyncedDisplay = ref(true)

const form = reactive({
  code: '',
  name: '',
  description: '',
  git_url: '',
  app_display_name: '',
  app_id: '',
  channel_index: '0',
  package_android: '',
})

const canCreateProject = computed(() => {
  if (auth.user?.role === 'admin') return true
  return items.value.some((p) => EDIT_ROLES.includes(p.my_role))
})

function canEditProject(row) {
  return EDIT_ROLES.includes(row?.my_role)
}

function rowClassName({ row }) {
  return canEditProject(row) ? 'row-editable' : 'row-readonly'
}

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function onNameInput() {
  if (nameSyncedDisplay.value || !form.app_display_name) {
    form.app_display_name = form.name
    nameSyncedDisplay.value = true
  }
}

function resetForm() {
  Object.assign(form, {
    code: '',
    name: '',
    description: '',
    git_url: '',
    app_display_name: '',
    app_id: '',
    channel_index: '0',
    package_android: '',
  })
}

async function load() {
  loading.value = true
  try {
    items.value = await listProjects()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = false
  editId.value = null
  editingRow.value = null
  settingsSnapshot.value = null
  nameSyncedDisplay.value = true
  resetForm()
  dialogVisible.value = true
}

function onRowClick(row) {
  if (!canEditProject(row)) return
  openEdit(row)
}

async function openEdit(row) {
  editing.value = true
  editId.value = row.id
  editingRow.value = row
  Object.assign(form, {
    code: row.code,
    name: row.name,
    description: row.description,
    git_url: row.git_url || '',
    app_display_name: '',
    app_id: '',
    channel_index: '0',
    package_android: '',
  })

  try {
    const settings = await getProjectSettings(row.id)
    settingsSnapshot.value = settings || {}
    const display = settings?.app_display_name || ''
    nameSyncedDisplay.value = !display || display === row.name
    Object.assign(form, {
      app_display_name: display || row.name || '',
      app_id: settings?.app_id || '',
      channel_index: settings?.channel_index || '0',
      package_android: settings?.package_android || '',
    })
  } catch {
    settingsSnapshot.value = {}
    form.app_display_name = row.name || ''
    nameSyncedDisplay.value = true
  }

  dialogVisible.value = true
}

async function submit() {
  saving.value = true
  try {
    let projectId = editId.value
    if (editing.value) {
      // 只改对话框里的字段，构建相关仍走「打包构建」
      await updateProject(projectId, {
        ...editingRow.value,
        code: form.code,
        name: form.name,
        description: form.description,
        git_url: form.git_url?.trim() || '',
      })
      const base = { ...(settingsSnapshot.value || {}), project_id: projectId }
      await saveProjectSettings(projectId, {
        ...base,
        app_display_name: form.app_display_name?.trim() || form.name,
        app_id: form.app_id?.trim() || '',
        channel_index: form.channel_index?.trim() || '0',
        package_android: form.package_android?.trim() || '',
      })
    } else {
      const created = await createProject({
        code: form.code,
        name: form.name,
        description: form.description,
        git_url: form.git_url?.trim() || '',
        artifact_output_dir: defaultArtifactPath(form.code, null),
        status: 'active',
      })
      projectId = created?.id
    }

    await projectStore.fetchProjects()
    dialogVisible.value = false
    await load()
    ElMessage.success('已保存')
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.project-form :deep(.el-divider) { margin: 8px 0 18px; }
.hint { font-size: 12px; color: #909399; margin-top: 6px; line-height: 1.4; }
.hint code {
  background: #f0f2f5;
  padding: 0 4px;
  border-radius: 3px;
}
:deep(.row-editable) { cursor: pointer; }
:deep(.row-readonly) { cursor: default; }
.build-cell { white-space: nowrap; }
</style>
