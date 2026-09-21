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
      <el-table-column label="构建" width="120">
        <template #default="{ row }">
          <span v-if="row.build_profile">{{ row.build_profile }} / {{ row.build_platform || 'android' }}</span>
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

    <el-dialog v-model="dialogVisible" :title="editing ? '项目设置' : '新建项目'" width="620px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="代号"><el-input v-model="form.code" :disabled="editing" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <template v-if="editing">
          <el-divider>Git 与构建</el-divider>
          <el-form-item label="主仓库">
            <el-input
              v-model="form.git_url"
              placeholder="https://gitlab.com/org/repo.git 或 git@git.xxx.com:org/repo.git"
            />
            <div class="hint">支持 GitHub / GitLab / Gitee / 自建 Git，HTTPS 或 SSH 均可</div>
          </el-form-item>
          <el-form-item label="主仓库分支">
            <GitBranchSelect v-model="form.git_branch" :git-url="form.git_url" @update:model-value="onBranchChange" />
          </el-form-item>

          <div v-for="(repo, idx) in extraGitRepos" :key="idx" class="extra-repo-block">
            <el-form-item :label="`仓库 ${idx + 2}`">
              <div class="repo-fields">
                <el-input v-model="repo.name" placeholder="名称，如：配置仓 / Android 仓" class="repo-name" />
                <el-input v-model="repo.git_url" placeholder="Git 地址" />
                <GitBranchSelect v-model="repo.git_branch" :git-url="repo.git_url" />
                <el-button link type="danger" @click="extraGitRepos.splice(idx, 1)">删除</el-button>
              </div>
            </el-form-item>
          </div>
          <el-form-item label=" ">
            <el-button size="small" @click="extraGitRepos.push({ name: '', git_url: '', git_branch: 'main' })">
              添加其他仓库
            </el-button>
          </el-form-item>

          <el-form-item label="构建环境">
            <el-radio-group v-model="form.build_profile" @change="onBuildPrefChange">
              <el-radio-button value="develop">Develop (Debug)</el-radio-button>
              <el-radio-button value="release">Release</el-radio-button>
              <el-radio-button value="custom">自定义</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="目标平台">
            <el-radio-group v-model="form.build_platform" :disabled="form.build_profile === 'custom'" @change="onBuildPrefChange">
              <el-radio-button value="android">Android (APK，默认)</el-radio-button>
              <el-radio-button value="ios">iOS</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="构建命令">
            <el-input
              v-model="form.build_command"
              type="textarea"
              :rows="3"
              :disabled="form.build_profile !== 'custom'"
              :placeholder="commandPreview"
            />
            <div v-if="form.build_profile !== 'custom'" class="hint">当前将执行：{{ commandPreview }}</div>
          </el-form-item>

          <el-form-item label="产物输出目录">
            <el-input v-model="form.artifact_output_dir" :placeholder="defaultArtifactHint">
              <template #append>
                <el-button @click="pathPickerVisible = true">选择路径</el-button>
              </template>
            </el-input>
            <div class="hint">
              初始默认：<code>{{ defaultArtifactHint }}</code>。仅支持 <code>data/artifacts/...</code> 相对路径，每次构建产物会存到该目录下的 <code>build_编号/</code> 子目录。
            </div>
          </el-form-item>
          <PathPickerDialog
            v-model="pathPickerVisible"
            :initial-path="form.artifact_output_dir || defaultArtifactHint"
            @select="onPathSelected"
          />
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import GitBranchSelect from '../components/GitBranchSelect.vue'
import PathPickerDialog from '../components/PathPickerDialog.vue'
import { createProject, listProjects, updateProject } from '../api'
import { useAuthStore } from '../stores/auth'
import { useProjectStore } from '../stores/project'
import {
  defaultArtifactPath,
  inferPlatformFromCommand,
  inferProfileFromCommand,
  resolveBuildCommand,
} from '../utils/buildPresets'
import { parseGitRepos } from '../utils/gitRepos'
import { defaultBuildPrefs, getProjectPrefs, saveProjectPrefs } from '../utils/projectPrefs'

const EDIT_ROLES = ['owner', 'super_admin', 'admin', 'developer']

const auth = useAuthStore()
const projectStore = useProjectStore()
const loading = ref(false)
const items = ref([])
const dialogVisible = ref(false)
const editing = ref(false)
const editId = ref(null)
const pathPickerVisible = ref(false)
const extraGitRepos = ref([])
const form = reactive({
  code: '', name: '', description: '',
  git_url: '', git_branch: 'main',
  build_profile: 'develop', build_platform: 'android',
  build_command: '', artifact_output_dir: '',
})

const canCreateProject = computed(() => {
  if (auth.user?.role === 'admin') return true
  return items.value.some((p) => EDIT_ROLES.includes(p.my_role))
})

const defaultArtifactHint = computed(() => defaultArtifactPath(form.code, editId.value))

const commandPreview = computed(() =>
  resolveBuildCommand(form.build_profile, form.build_platform, form.build_command),
)

function canEditProject(row) {
  return EDIT_ROLES.includes(row?.my_role)
}

function rowClassName({ row }) {
  return canEditProject(row) ? 'row-editable' : 'row-readonly'
}

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function syncCommandPreview() {
  if (form.build_profile !== 'custom') {
    form.build_command = commandPreview.value
  }
}

function onBuildPrefChange() {
  syncCommandPreview()
  if (editId.value) {
    saveProjectPrefs(editId.value, {
      build_profile: form.build_profile,
      build_platform: form.build_platform,
    })
  }
}

function onBranchChange(branch) {
  if (editId.value) {
    saveProjectPrefs(editId.value, { git_branch: branch || 'main' })
  }
}

function onPathSelected(path) {
  form.artifact_output_dir = path
}

function applyBuildPrefs(projectId, row = {}) {
  const prefs = getProjectPrefs(projectId)
  const profile = row.build_profile || prefs.build_profile || inferProfileFromCommand(row.build_command) || defaultBuildPrefs.build_profile
  const platform = row.build_platform || prefs.build_platform || inferPlatformFromCommand(row.build_command) || defaultBuildPrefs.build_platform
  return {
    git_branch: row.git_branch || prefs.git_branch || defaultBuildPrefs.git_branch,
    build_profile: profile,
    build_platform: platform,
  }
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
  extraGitRepos.value = []
  Object.assign(form, {
    code: '', name: '', description: '',
    git_url: '', git_branch: defaultBuildPrefs.git_branch,
    build_profile: defaultBuildPrefs.build_profile,
    build_platform: defaultBuildPrefs.build_platform,
    build_command: '', artifact_output_dir: '',
  })
  dialogVisible.value = true
}

function onRowClick(row) {
  if (!canEditProject(row)) return
  openEdit(row)
}

function openEdit(row) {
  editing.value = true
  editId.value = row.id
  const prefs = applyBuildPrefs(row.id, row)
  Object.assign(form, {
    code: row.code,
    name: row.name,
    description: row.description,
    git_url: row.git_url || '',
    git_branch: prefs.git_branch,
    build_profile: prefs.build_profile,
    build_platform: prefs.build_platform,
    build_command: row.build_command || resolveBuildCommand(prefs.build_profile, prefs.build_platform, ''),
    artifact_output_dir: row.artifact_output_dir || defaultArtifactPath(row.code, row.id),
  })
  extraGitRepos.value = parseGitRepos(row.git_repos).map((r) => ({
    name: r.name || '',
    git_url: r.git_url || '',
    git_branch: r.git_branch || 'main',
  }))
  dialogVisible.value = true
}

async function submit() {
  const payload = {
    ...form,
    git_branch: form.git_branch?.trim() || 'main',
    git_repos: JSON.stringify(
      extraGitRepos.value.filter((r) => r.git_url?.trim()).map((r) => ({
        name: r.name?.trim() || r.git_url.trim(),
        git_url: r.git_url.trim(),
        git_branch: r.git_branch?.trim() || 'main',
      })),
    ),
  }
  if (payload.build_profile !== 'custom') {
    payload.build_command = resolveBuildCommand(payload.build_profile, payload.build_platform, '')
  }
  if (!payload.artifact_output_dir) {
    payload.artifact_output_dir = defaultArtifactPath(payload.code, editId.value)
  }
  if (editing.value) {
    await updateProject(editId.value, payload)
    saveProjectPrefs(editId.value, {
      git_branch: payload.git_branch,
      build_profile: payload.build_profile,
      build_platform: payload.build_platform,
    })
    await projectStore.fetchProjects()
  } else {
    await createProject(payload)
    await projectStore.fetchProjects()
  }
  dialogVisible.value = false
  await load()
}

onMounted(load)
</script>

<style scoped>
.hint { font-size: 12px; color: #909399; margin-top: 6px; line-height: 1.4; }
.extra-repo-block { margin-bottom: 4px; }
.repo-fields { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.repo-name { max-width: 280px; }
:deep(.row-editable) { cursor: pointer; }
:deep(.row-readonly) { cursor: default; }
</style>
