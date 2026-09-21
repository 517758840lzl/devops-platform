<template>
  <div>
    <div class="page-header">
      <h2>发布管理 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <el-button v-if="canWrite" type="primary" @click="openCreate">新建发布单</el-button>
    </div>
    <el-table v-loading="loading" :data="items" stripe @row-click="openDetail">
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column prop="status" label="状态" width="140">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="release_notes" label="发布说明" min-width="200" show-overflow-tooltip />
      <el-table-column prop="updated_at" label="更新时间" width="170">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="detailVisible" :title="detail?.title || '发布详情'" size="560px">
      <template v-if="detail">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="版本">{{ detail.version }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(detail.status)" size="small">{{ statusLabel(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发布说明"><pre class="pre">{{ detail.release_notes }}</pre></el-descriptions-item>
          <el-descriptions-item v-if="detail.reject_reason" label="驳回原因">
            <span style="color:#f56c6c">{{ detail.reject_reason }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <div class="actions">
          <el-button v-if="canWrite && (detail.status === 'draft' || detail.status === 'rejected')" type="primary" @click="doSubmit">提交审批</el-button>
          <el-button v-if="canApprove && detail.status === 'pending_approval'" type="success" @click="doApprove">审批通过</el-button>
          <el-button v-if="canApprove && detail.status === 'pending_approval'" type="danger" @click="doReject">驳回</el-button>
          <el-button v-if="canWrite && detail.status === 'approved'" type="primary" plain :loading="building" @click="doBuild">拉取 Git 并打包</el-button>
          <el-button v-if="canWrite && detail.status === 'approved'" type="warning" @click="doPublish">执行发布</el-button>
        </div>

        <el-divider v-if="releaseBuilds.length">构建记录</el-divider>
        <el-table v-if="releaseBuilds.length" :data="releaseBuilds" size="small" stripe>
          <el-table-column label="#" width="50">
            <template #default="{ row }">{{ row.build_number || row.id }}</template>
          </el-table-column>
          <el-table-column prop="branch" label="分支" width="80" />
          <el-table-column prop="commit_sha" label="Commit" width="100">
            <template #default="{ row }">{{ row.commit_sha?.slice(0, 8) || '-' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="产物" width="100">
            <template #default="{ row }">
              <span v-if="row.artifacts_deleted" class="artifact-deleted">已删除</span>
              <el-button
                v-else-if="row.status === 'success' && row.artifact_path"
                link
                type="primary"
                :loading="downloadingBuildId === row.id"
                :disabled="downloadingBuildId !== null && downloadingBuildId !== row.id"
                @click.stop="downloadBuild(row.id)"
              >
                {{ downloadingBuildId === row.id ? '下载中' : '下载' }}
              </el-button>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>

        <el-divider>审批记录</el-divider>
        <el-timeline>
          <el-timeline-item v-for="a in detail.approvals || []" :key="a.id" :timestamp="formatTime(a.created_at)">
            <strong>{{ a.user_name }}</strong> — {{ approvalLabel(a.action) }}
            <div v-if="a.comment" class="approval-comment">{{ a.comment }}</div>
          </el-timeline-item>
        </el-timeline>

        <el-divider>评论与日志</el-divider>
        <CommentActivityPanel target-type="release" :target-id="detail.id" />
      </template>
    </el-drawer>

    <el-dialog v-model="dialogVisible" title="新建发布单" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="所属项目">
          <el-input :model-value="projectStore.current?.name" disabled />
        </el-form-item>
        <el-form-item label="版本"><el-input v-model="form.version" placeholder="1.2.0" /></el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="发布说明"><el-input v-model="form.release_notes" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import CommentActivityPanel from '../components/CommentActivityPanel.vue'
import { useProjectPermission } from '../composables/useProjectPermission'
import {
  approveRelease, createRelease, downloadBuildArtifacts, getRelease, listBuilds, listReleases,
  publishRelease, rejectRelease, submitRelease, triggerProjectBuild,
} from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const { canWrite, canApprove, refresh: refreshPermission } = useProjectPermission()
const loading = ref(false)
const items = ref([])
const detail = ref(null)
const detailVisible = ref(false)
const dialogVisible = ref(false)
const building = ref(false)
const downloadingBuildId = ref(null)
const releaseBuilds = ref([])
const form = reactive({ version: '', title: '', release_notes: '' })

const statusMap = {
  draft: '草稿',
  pending_approval: '待审批',
  approved: '已通过',
  rejected: '已驳回',
  published: '已发布',
}

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function statusLabel(s) {
  return statusMap[s] || s
}

function statusType(s) {
  if (s === 'published') return 'success'
  if (s === 'pending_approval') return 'warning'
  if (s === 'approved') return 'primary'
  if (s === 'rejected') return 'danger'
  return 'info'
}

function approvalLabel(a) {
  return { submit: '提交审批', approve: '审批通过', reject: '驳回', publish: '执行发布' }[a] || a
}

async function load() {
  loading.value = true
  try {
    items.value = await listReleases({ project_id: projectStore.currentIdOrDefault })
  } finally {
    loading.value = false
  }
}

async function loadReleaseBuilds(releaseId) {
  releaseBuilds.value = await listBuilds({ project_id: projectStore.currentIdOrDefault, release_id: releaseId })
}

async function openDetail(row) {
  detail.value = await getRelease(row.id)
  await loadReleaseBuilds(row.id)
  detailVisible.value = true
}

function openCreate() {
  Object.assign(form, { version: '', title: '', release_notes: '' })
  dialogVisible.value = true
}

async function submitCreate() {
  await createRelease({ ...form, project_id: projectStore.currentIdOrDefault })
  dialogVisible.value = false
  await load()
}

async function refreshDetail() {
  if (detail.value?.id) {
    detail.value = await getRelease(detail.value.id)
    await load()
  }
}

async function doSubmit() {
  await submitRelease(detail.value.id, { comment: '提交发布审批' })
  await refreshDetail()
}

async function doApprove() {
  await approveRelease(detail.value.id, { comment: '审批通过，可以发布' })
  await refreshDetail()
}

async function doReject() {
  const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回发布', { inputPattern: /.+/, inputErrorMessage: '原因不能为空' })
  await rejectRelease(detail.value.id, { reason: value })
  await refreshDetail()
}

async function doBuild() {
  building.value = true
  try {
    await triggerProjectBuild(projectStore.currentIdOrDefault, { release_id: detail.value.id })
    await loadReleaseBuilds(detail.value.id)
  } finally {
    building.value = false
  }
}

async function doPublish() {
  await publishRelease(detail.value.id, { comment: '已上线' })
  await refreshDetail()
}

async function downloadBuild(buildId) {
  if (downloadingBuildId.value !== null) return
  downloadingBuildId.value = buildId
  const loadingMsg = ElMessage.info({ message: '正在准备下载...', duration: 0 })
  try {
    const { filename, size } = await downloadBuildArtifacts(buildId)
    loadingMsg.close()
    const mb = size < 1024 * 1024 ? `${(size / 1024).toFixed(1)} KB` : `${(size / 1024 / 1024).toFixed(1)} MB`
    ElMessage.success(`已下载 ${filename}（${mb}）`)
  } catch (err) {
    loadingMsg.close()
    ElMessage.error(err.message || '下载失败')
  } finally {
    downloadingBuildId.value = null
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
.pre { white-space: pre-wrap; margin: 0; font-family: inherit; }
.actions { margin: 16px 0; display: flex; gap: 8px; flex-wrap: wrap; }
.approval-comment { color: #606266; font-size: 13px; margin-top: 4px; }
.artifact-deleted { color: #c0c4cc; font-size: 13px; cursor: not-allowed; user-select: none; }
</style>
