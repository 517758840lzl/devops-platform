<template>
  <div>
    <div class="page-header">
      <h2>
        打包构建
        <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag>
      </h2>
      <div class="header-actions">
        <el-button
          type="primary"
          :disabled="!projectStore.current"
          @click="configVisible = true"
        >
          拉取并构建
        </el-button>
      </div>
    </div>

    <el-alert v-if="selectedRepo" type="success" :closable="false" show-icon style="margin-bottom:12px">
      {{ selectedRepo.name }} ·
      <template v-if="buildCommit.trim()">Commit {{ buildCommit.trim().slice(0, 8) }}</template>
      <template v-else>{{ buildBranch }}</template>
      · {{ buildProfileLabel }} / {{ buildPlatformLabel }}
    </el-alert>
    <el-alert v-else type="warning" :closable="false" show-icon style="margin-bottom:12px">
      当前项目未配置 Git 仓库，请前往「项目」页编辑项目设置
    </el-alert>

    <el-dialog
      v-model="configVisible"
      title="拉取并构建"
      width="720px"
      class="build-dialog"
      destroy-on-close
    >
      <div class="build-config">
        <section class="config-section">
          <div class="section-head">源码</div>
          <div v-if="gitRepoOptions.length > 1" class="config-row">
            <span class="config-label">仓库</span>
            <div class="config-field">
              <el-select v-model="selectedGitURL" style="width: 100%" placeholder="选择仓库">
                <el-option
                  v-for="repo in gitRepoOptions"
                  :key="repo.key"
                  :label="`${repo.name} (${repo.git_branch})`"
                  :value="repo.git_url"
                />
              </el-select>
            </div>
          </div>
          <div class="config-row">
            <span class="config-label">分支</span>
            <div class="config-field">
              <GitBranchSelect
                v-if="selectedRepo"
                v-model="buildBranch"
                :git-url="selectedRepo.git_url"
                :disabled="!canBuild"
              />
              <span v-else class="hint">请先配置 Git 仓库</span>
            </div>
          </div>
          <div class="config-row config-row-top">
            <span class="config-label">Commit</span>
            <div class="config-field">
              <el-input
                v-model="buildCommit"
                clearable
                placeholder="可选，填 Commit ID 则按该提交构建"
              />
              <p class="hint">留空打分支最新提交；填写后优先按 Commit 构建（可填短 SHA）</p>
            </div>
          </div>
        </section>

        <section class="config-section">
          <div class="section-head">构建选项</div>
          <div class="config-row">
            <span class="config-label">构建环境</span>
            <div class="config-field">
              <el-radio-group v-model="buildProfile" @change="onBuildPrefChange">
                <el-radio-button value="develop">Develop (Debug)</el-radio-button>
                <el-radio-button value="release">Release</el-radio-button>
                <el-radio-button value="custom">自定义</el-radio-button>
              </el-radio-group>
            </div>
          </div>
          <div class="config-row">
            <span class="config-label">目标平台</span>
            <div class="config-field">
              <el-radio-group v-model="buildPlatform" :disabled="buildProfile === 'custom'" @change="onBuildPrefChange">
                <el-radio-button value="android">Android (APK)</el-radio-button>
                <el-radio-button value="ios">iOS</el-radio-button>
              </el-radio-group>
            </div>
          </div>
          <div class="config-row config-row-top">
            <span class="config-label">构建命令</span>
            <div class="config-field">
              <el-input
                v-model="buildCommand"
                type="textarea"
                :rows="3"
                :disabled="buildProfile !== 'custom'"
                :placeholder="commandPreview"
              />
              <p v-if="buildProfile !== 'custom'" class="hint">当前将执行：{{ commandPreview }}</p>
            </div>
          </div>
        </section>

        <section class="config-section">
          <div class="section-head">产物</div>
          <div class="config-row config-row-top">
            <span class="config-label">产物目录</span>
            <div class="config-field">
              <el-input v-model="artifactOutputDir" :placeholder="defaultArtifactHint">
                <template #append>
                  <el-button @click="pathPickerVisible = true">选择路径</el-button>
                </template>
              </el-input>
              <p class="hint">仅支持 <code>data/artifacts/...</code>；每次构建写入其下的 <code>build_编号/</code></p>
            </div>
          </div>
        </section>
      </div>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" :loading="triggering" :disabled="!canBuild" @click="triggerBuild">
          拉取并构建
        </el-button>
      </template>
    </el-dialog>

    <PathPickerDialog
      v-model="pathPickerVisible"
      :initial-path="artifactOutputDir || defaultArtifactHint"
      @select="onPathSelected"
    />

    <el-table v-loading="loading" :data="items" stripe @row-click="openDetail">
      <el-table-column label="#" width="70">
        <template #default="{ row }">{{ row.build_number || row.id }}</template>
      </el-table-column>
      <el-table-column prop="branch" label="分支" width="100" />
      <el-table-column prop="commit_sha" label="Commit" min-width="120">
        <template #default="{ row }">{{ row.commit_sha?.slice(0, 8) || '-' }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="buildStatusType(row.status)" size="small">{{ buildStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="触发时间" width="170">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="产物" width="160" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'success'">
            <span v-if="row.artifacts_deleted" class="artifact-deleted">已删除</span>
            <template v-else-if="row.artifact_path">
              <el-button
                link
                type="primary"
                :loading="downloadingId === row.id"
                :disabled="downloadingId !== null && downloadingId !== row.id"
                @click.stop="downloadArtifacts(row.id)"
              >
                {{ downloadingId === row.id ? '下载中' : '下载' }}
              </el-button>
              <el-button
                link
                type="danger"
                :loading="deletingId === row.id"
                :disabled="deletingId !== null && deletingId !== row.id"
                @click.stop="deleteArtifacts(row)"
              >
                删除
              </el-button>
            </template>
            <span v-else>-</span>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="detailVisible" title="构建日志" size="640px">
      <template v-if="current">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="状态">
            <el-tag :type="buildStatusType(current.status)" size="small">{{ buildStatusLabel(current.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分支">{{ current.branch }}</el-descriptions-item>
          <el-descriptions-item label="Commit" :span="2">{{ current.commit_sha || '-' }}</el-descriptions-item>
          <el-descriptions-item label="仓库" :span="2">{{ current.git_url }}</el-descriptions-item>
          <el-descriptions-item v-if="current.artifact_path" label="产物目录" :span="2">{{ current.artifact_path }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="current.status === 'success' && current.artifacts_deleted" class="artifact-actions">
          <span class="artifact-deleted">产物已删除</span>
        </div>
        <div v-else-if="current.status === 'success' && artifactFiles.length" class="artifact-list">
          <div class="section-title">构建产物</div>
          <div v-for="f in artifactFiles" :key="f.path" class="artifact-row">
            <span class="artifact-name">{{ f.download_name || f.name }}</span>
            <span class="artifact-size">{{ formatSize(f.size) }}</span>
          </div>
        </div>
        <div v-if="current.status === 'success' && current.artifact_path && !current.artifacts_deleted" class="artifact-actions">
          <el-button
            type="primary"
            plain
            size="small"
            :loading="downloadingId === current.id"
            :disabled="downloadingId !== null && downloadingId !== current.id"
            @click="downloadArtifacts(current.id)"
          >
            {{ downloadButtonLabel }}
          </el-button>
          <el-button
            type="danger"
            plain
            size="small"
            :loading="deletingId === current.id"
            @click="deleteArtifacts(current)"
          >
            删除产物
          </el-button>
        </div>
        <pre class="log">{{ current.log || '等待日志...' }}</pre>
        <el-button v-if="current.status === 'running' || current.status === 'pending'" size="small" @click="refreshDetail">刷新</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import GitBranchSelect from '../components/GitBranchSelect.vue'
import PathPickerDialog from '../components/PathPickerDialog.vue'
import {
  deleteBuildArtifacts,
  downloadBuildArtifacts,
  getBuild,
  listBuildArtifacts,
  getProject,
  listBuilds,
  triggerProjectBuild,
  updateProject,
} from '../api'
import { useLocalNotificationStore } from '../stores/localNotifications'
import { useProjectStore } from '../stores/project'
import {
  defaultArtifactPath,
  inferPlatformFromCommand,
  inferProfileFromCommand,
  resolveBuildCommand,
} from '../utils/buildPresets'
import { listProjectGitRepos, parseGitRepos } from '../utils/gitRepos'
import { defaultBuildPrefs, getProjectPrefs, saveProjectPrefs } from '../utils/projectPrefs'

const projectStore = useProjectStore()
const notif = useLocalNotificationStore()
const loading = ref(false)
const triggering = ref(false)
const downloadingId = ref(null)
const deletingId = ref(null)
const artifactFiles = ref([])
const items = ref([])
const detailVisible = ref(false)
const current = ref(null)
const pathPickerVisible = ref(false)
const configVisible = ref(false)
let pollTimer = null

const gitRepoOptions = computed(() => listProjectGitRepos(projectStore.current))
const selectedGitURL = ref('')
const buildBranch = ref('main')
const buildCommit = ref('')
const buildProfile = ref(defaultBuildPrefs.build_profile)
const buildPlatform = ref(defaultBuildPrefs.build_platform)
const buildCommand = ref('')
const artifactOutputDir = ref('')

const selectedRepo = computed(() =>
  gitRepoOptions.value.find((r) => r.git_url === selectedGitURL.value) || gitRepoOptions.value[0] || null,
)
const canBuild = computed(() => gitRepoOptions.value.length > 0)
const defaultArtifactHint = computed(() =>
  defaultArtifactPath(projectStore.current?.code, projectStore.currentIdOrDefault),
)
const commandPreview = computed(() =>
  resolveBuildCommand(buildProfile.value, buildPlatform.value, buildCommand.value),
)
const buildProfileLabel = computed(() => ({
  develop: 'Debug',
  release: 'Release',
  custom: '自定义',
}[buildProfile.value] || buildProfile.value))
const buildPlatformLabel = computed(() => ({
  android: 'Android',
  ios: 'iOS',
}[buildPlatform.value] || buildPlatform.value))
const downloadButtonLabel = computed(() => {
  if (downloadingId.value === current.value?.id) return '下载中...'
  if (artifactFiles.value.length === 1) {
    return `下载 ${artifactFiles.value[0].download_name || artifactFiles.value[0].name}`
  }
  if (artifactFiles.value.length > 1) return `下载全部 (${artifactFiles.value.length} 个文件)`
  return '下载构建产物'
})

function syncBuildBranch() {
  const repo = selectedRepo.value
  if (!repo) {
    buildBranch.value = 'main'
    return
  }
  const prefs = getProjectPrefs(projectStore.currentIdOrDefault)
  buildBranch.value = repo.isPrimary
    ? (prefs.git_branch || repo.git_branch || 'main')
    : (repo.git_branch || 'main')
}

function syncBuildConfig() {
  const project = projectStore.current
  if (!project) {
    buildProfile.value = defaultBuildPrefs.build_profile
    buildPlatform.value = defaultBuildPrefs.build_platform
    buildCommand.value = ''
    artifactOutputDir.value = ''
    return
  }
  const prefs = getProjectPrefs(project.id)
  const profile = project.build_profile || prefs.build_profile || inferProfileFromCommand(project.build_command) || defaultBuildPrefs.build_profile
  const platform = project.build_platform || prefs.build_platform || inferPlatformFromCommand(project.build_command) || defaultBuildPrefs.build_platform
  buildProfile.value = profile
  buildPlatform.value = platform
  buildCommand.value = project.build_command || resolveBuildCommand(profile, platform, '')
  artifactOutputDir.value = project.artifact_output_dir || defaultArtifactPath(project.code, project.id)
}

function onBuildPrefChange() {
  if (buildProfile.value !== 'custom') {
    buildCommand.value = resolveBuildCommand(buildProfile.value, buildPlatform.value, '')
  }
  const projectId = projectStore.currentIdOrDefault
  if (projectId) {
    saveProjectPrefs(projectId, {
      build_profile: buildProfile.value,
      build_platform: buildPlatform.value,
    })
  }
}

function onPathSelected(path) {
  artifactOutputDir.value = path
}

watch(gitRepoOptions, (repos) => {
  if (!repos.length) {
    selectedGitURL.value = ''
    buildBranch.value = 'main'
    return
  }
  if (!repos.some((r) => r.git_url === selectedGitURL.value)) {
    selectedGitURL.value = repos[0].git_url
  }
  syncBuildBranch()
}, { immediate: true })

watch(selectedGitURL, () => syncBuildBranch())
watch(() => projectStore.current, syncBuildConfig, { immediate: true })

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function formatSize(size) {
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function buildStatusLabel(s) {
  return { pending: '排队中', running: '构建中', success: '成功', failed: '失败' }[s] || s
}

function buildStatusType(s) {
  if (s === 'success') return 'success'
  if (s === 'failed') return 'danger'
  if (s === 'running') return 'warning'
  return 'info'
}

async function load() {
  loading.value = true
  try {
    items.value = await listBuilds({ project_id: projectStore.currentIdOrDefault })
  } finally {
    loading.value = false
  }
}

async function persistBuildSettings() {
  const projectId = projectStore.currentIdOrDefault
  const repo = selectedRepo.value
  const branch = buildBranch.value?.trim() || 'main'
  if (!projectId || !repo) return

  const profile = buildProfile.value || defaultBuildPrefs.build_profile
  const platform = buildPlatform.value || defaultBuildPrefs.build_platform
  const command = profile === 'custom'
    ? (buildCommand.value?.trim() || '')
    : resolveBuildCommand(profile, platform, '')
  const artifactDir = artifactOutputDir.value?.trim() || defaultArtifactPath(projectStore.current?.code, projectId)

  saveProjectPrefs(projectId, {
    git_branch: branch,
    build_profile: profile,
    build_platform: platform,
  })

  const project = await getProject(projectId)
  const payload = {
    ...project,
    build_profile: profile,
    build_platform: platform,
    build_command: command,
    artifact_output_dir: artifactDir,
  }

  if (repo.isPrimary) {
    payload.git_branch = branch
  } else {
    payload.git_repos = JSON.stringify(
      parseGitRepos(project.git_repos).map((r) => (
        r.git_url === repo.git_url ? { ...r, git_branch: branch } : r
      )),
    )
  }

  await updateProject(projectId, payload)
  await projectStore.fetchProjects()
  syncBuildConfig()
}

async function triggerBuild() {
  if (!selectedRepo.value) return
  if (buildProfile.value === 'custom' && !buildCommand.value?.trim()) {
    ElMessage.warning('请填写自定义构建命令')
    return
  }
  triggering.value = true
  try {
    await persistBuildSettings()
    const branch = buildBranch.value?.trim() || 'main'
    const commitSha = buildCommit.value?.trim() || ''
    await triggerProjectBuild(projectStore.currentIdOrDefault, {
      git_url: selectedRepo.value.git_url,
      branch,
      ...(commitSha ? { commit_sha: commitSha } : {}),
    })
    configVisible.value = false
    await load()
  } catch (err) {
    ElMessage.error(err.message || '触发构建失败')
  } finally {
    triggering.value = false
  }
}

async function downloadArtifacts(buildId) {
  if (downloadingId.value !== null) return
  downloadingId.value = buildId
  const loadingMsg = ElMessage.info({ message: '正在准备下载...', duration: 0 })
  try {
    const { filename, size } = await downloadBuildArtifacts(buildId)
    loadingMsg.close()
    ElMessage.success(`已下载 ${filename}（${formatSize(size)}）`)
  } catch (err) {
    loadingMsg.close()
    ElMessage.error(err.message || '下载失败')
  } finally {
    downloadingId.value = null
  }
}

async function deleteArtifacts(row) {
  if (!row?.id || deletingId.value !== null) return
  try {
    await ElMessageBox.confirm(
      `确定删除构建 #${row.build_number || row.id} 的产物与工作区？删除后不可再下载，可释放磁盘空间。`,
      '删除产物',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  deletingId.value = row.id
  try {
    const updated = await deleteBuildArtifacts(row.id)
    Object.assign(row, updated)
    if (current.value?.id === row.id) {
      current.value = { ...current.value, ...updated }
      artifactFiles.value = []
    }
    const listRow = items.value.find((i) => i.id === row.id)
    if (listRow) Object.assign(listRow, updated)
    notif.add({
      key: `build:${row.id}:artifacts_deleted`,
      type: 'build',
      title: '产物已删除',
      message: `构建 #${row.build_number || row.id} · ${row.branch}${row.commit_sha ? ` · ${String(row.commit_sha).slice(0, 8)}` : ''}`,
      link: '/builds',
    })
    ElMessage.success('产物已删除')
  } catch (err) {
    ElMessage.error(err.message || '删除失败')
  } finally {
    deletingId.value = null
  }
}

async function loadArtifactFiles(buildId) {
  try {
    artifactFiles.value = await listBuildArtifacts(buildId)
  } catch {
    artifactFiles.value = []
  }
}

async function openDetail(row) {
  current.value = await getBuild(row.id)
  await loadArtifactFiles(row.id)
  detailVisible.value = true
  startPoll()
}

async function refreshDetail() {
  if (!current.value) return
  current.value = await getBuild(current.value.id)
  await loadArtifactFiles(current.value.id)
}

function startPoll() {
  stopPoll()
  pollTimer = setInterval(async () => {
    if (!detailVisible.value || !current.value) return
    if (current.value.status === 'success' || current.value.status === 'failed') {
      stopPoll()
      return
    }
    await refreshDetail()
    await load()
  }, 2000)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  await projectStore.fetchProjects()
  syncBuildConfig()
  await load()
})
watch(() => projectStore.currentId, async () => {
  await projectStore.fetchProjects()
  syncBuildConfig()
  await load()
})
onUnmounted(stopPoll)
</script>

<style scoped>
.header-actions { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.build-config {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.config-section {
  padding: 16px 18px 8px;
  border: 1px solid #ebeef5;
  border-radius: 10px;
  background: #fafbfc;
}
.section-head {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 14px;
  letter-spacing: 0.02em;
}
.config-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 18px;
}
.config-row-top { align-items: flex-start; }
.config-label {
  width: 88px;
  flex-shrink: 0;
  color: #606266;
  font-size: 14px;
  line-height: 32px;
  text-align: right;
}
.config-field { flex: 1; min-width: 0; }
.hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
}
.hint code {
  background: #f0f2f5;
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 12px;
}
.section-title { font-weight: 600; margin: 16px 0 8px; font-size: 14px; color: #303133; }
.artifact-list { margin-top: 8px; }
.artifact-row { display: flex; justify-content: space-between; gap: 12px; font-size: 13px; padding: 6px 0; border-bottom: 1px solid #f0f2f5; }
.artifact-name { color: #303133; word-break: break-all; }
.artifact-size { color: #909399; flex-shrink: 0; }
.artifact-actions { margin-top: 12px; display: flex; gap: 8px; align-items: center; }
.artifact-deleted { color: #c0c4cc; font-size: 13px; cursor: not-allowed; user-select: none; }
.log {
  margin-top: 16px;
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
  max-height: 480px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

<style>
.build-dialog .el-dialog__body {
  padding: 12px 20px 8px;
}
.build-dialog .el-dialog__footer {
  padding: 12px 20px 20px;
}
</style>
