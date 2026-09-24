<template>
  <div v-loading="loading">
    <div class="page-header">
      <h2>
        项目进度
        <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag>
      </h2>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-alert v-if="!projectStore.current" type="warning" show-icon :closable="false">
      请先在顶部选择项目
    </el-alert>

    <template v-else-if="progress">
      <el-card shadow="never" class="overview">
        <div class="overview-top">
          <div class="overview-main">
            <div class="pct">{{ displayPercent }}%</div>
            <div class="pct-label">总进度</div>
          </div>
          <div class="overview-meta">
            <div>
              <span class="meta-label">当前阶段</span>
              <el-tag type="warning" effect="plain">{{ progress.current_stage_title || '-' }}</el-tag>
            </div>
            <div>
              <span class="meta-label">阻塞项</span>
              <el-tag :type="progress.blocked_count ? 'danger' : 'success'" size="small">
                {{ progress.blocked_count || 0 }}
              </el-tag>
            </div>
            <div>
              <span class="meta-label">最近更新</span>
              <span class="meta-value">
                {{ progress.updated_by_name || '-' }}
                <template v-if="progress.updated_at"> · {{ progress.updated_at }}</template>
              </span>
            </div>
          </div>
        </div>
        <el-progress
          :percentage="Number(displayPercent)"
          :stroke-width="14"
          :status="progress.percent >= 100 ? 'success' : undefined"
        />
      </el-card>

      <el-card v-if="progress.hints?.length" shadow="never" class="hints-card">
        <template #header>
          <div class="hints-header">
            <span>缺口提醒（自动）</span>
            <el-button
              v-if="canEdit"
              size="small"
              type="warning"
              plain
              :loading="applyingHints"
              @click="onApplyHints"
            >
              根据缺口标为未完成
            </el-button>
          </div>
        </template>
        <div class="hints-list">
          <div
            v-for="h in progress.hints"
            :key="h.key"
            class="hint-item-wrap"
            role="link"
            @click="openHint(h)"
          >
            <el-alert
              :title="h.message"
              :type="hintAlertType(h.level)"
              show-icon
              :closable="false"
              class="hint-item"
            >
              <template #default>
                <span class="hint-cat">{{ hintCategoryLabel(h.category) }} · 点击前往</span>
              </template>
            </el-alert>
          </div>
        </div>
      </el-card>

      <el-card shadow="never" class="stages-card">
        <template #header>交付阶段</template>
        <div class="stage-axis">
          <button
            v-for="ms in progress.milestones"
            :key="ms.id"
            type="button"
            class="stage-chip"
            :class="{
              active: selectedMilestoneId === ms.id,
              done: ms.status === 'done',
              blocked: ms.status === 'blocked',
              doing: ms.status === 'doing',
            }"
            @click="selectedMilestoneId = ms.id"
          >
            <div class="stage-title">{{ ms.title }}</div>
            <div class="stage-count">{{ ms.done_count }}/{{ ms.total_count }}</div>
          </button>
        </div>
      </el-card>

      <el-card v-if="selectedMilestone" shadow="never" class="items-card">
        <template #header>
          <div class="items-header">
            <span>{{ selectedMilestone.title }} · 检查项</span>
            <el-tag size="small" effect="plain">{{ selectedMilestone.done_count }}/{{ selectedMilestone.total_count }}</el-tag>
          </div>
        </template>

        <el-alert v-if="!canEdit && loadedPerm" type="info" show-icon :closable="false" style="margin-bottom:12px">
          只读：tester / viewer 不可改进度，需 developer 及以上
        </el-alert>

        <el-table :data="selectedMilestone.items || []" stripe>
          <el-table-column prop="title" label="事项" min-width="200" />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <el-select
                :model-value="row.status"
                size="small"
                :disabled="!canEdit"
                @change="(v) => onStatusChange(row, v)"
              >
                <el-option label="待办" value="todo" />
                <el-option label="进行中" value="doing" />
                <el-option label="完成" value="done" />
                <el-option label="阻塞" value="blocked" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="负责人" width="160">
            <template #default="{ row }">
              <el-select
                :model-value="row.assignee_id || null"
                size="small"
                clearable
                filterable
                placeholder="未指派"
                :disabled="!canEdit"
                @change="(v) => onAssigneeChange(row, v)"
              >
                <el-option
                  v-for="m in members"
                  :key="m.user_id"
                  :label="m.name || m.username"
                  :value="m.user_id"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="备注" min-width="200">
            <template #default="{ row }">
              <el-input
                :model-value="row.note || ''"
                size="small"
                placeholder="备注"
                :disabled="!canEdit"
                @change="(v) => onNoteChange(row, v)"
              />
            </template>
          </el-table-column>
          <el-table-column label="权重" prop="weight" width="70" />
        </el-table>
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  applyProgressHints,
  getProjectProgress,
  listMembers,
  patchChecklistItem,
} from '../api'
import { useProjectPermission } from '../composables/useProjectPermission'
import { useProjectStore } from '../stores/project'

const router = useRouter()
const projectStore = useProjectStore()
const { canEditProjectConfig, loaded: loadedPerm, refresh: refreshPermission } = useProjectPermission()

const loading = ref(false)
const applyingHints = ref(false)
const progress = ref(null)
const members = ref([])
const selectedMilestoneId = ref(null)

const canEdit = computed(() => canEditProjectConfig.value || progress.value?.can_edit)

const displayPercent = computed(() => {
  const p = progress.value?.percent ?? 0
  return Math.round(p * 10) / 10
})

const selectedMilestone = computed(() => {
  const list = progress.value?.milestones || []
  return list.find((m) => m.id === selectedMilestoneId.value) || list[0] || null
})

function hintAlertType(level) {
  if (level === 'critical') return 'error'
  if (level === 'warning') return 'warning'
  return 'info'
}

function hintCategoryLabel(cat) {
  return ({ compliance: '协议', store: '素材', bug: '测试', build: '构建' })[cat] || cat
}

function hintRoute(h) {
  switch (h.category) {
    case 'store':
      return { path: '/config', query: { tab: 'store' } }
    case 'compliance':
      return { path: '/config', query: { tab: 'compliance' } }
    case 'bug':
      return { path: '/bugs' }
    case 'build':
      return { path: '/builds' }
    default:
      return null
  }
}

function openHint(h) {
  const to = hintRoute(h)
  if (to) router.push(to)
}

async function loadMembers() {
  const id = projectStore.currentIdOrDefault
  if (!id) {
    members.value = []
    return
  }
  members.value = await listMembers(id)
}

async function load() {
  const id = projectStore.currentIdOrDefault
  if (!id) {
    progress.value = null
    return
  }
  loading.value = true
  try {
    await refreshPermission()
    await loadMembers()
    const data = await getProjectProgress(id)
    progress.value = data
    const current = data.milestones?.find((m) => m.key === data.current_stage_key)
    selectedMilestoneId.value = current?.id || data.milestones?.[0]?.id || null
  } catch {
    progress.value = null
  } finally {
    loading.value = false
  }
}

async function patchItem(row, payload) {
  const id = projectStore.currentIdOrDefault
  if (!id || !canEdit.value) return
  try {
    const updated = await patchChecklistItem(id, row.id, payload)
    Object.assign(row, updated)
    // refresh summary numbers without full spinner
    const data = await getProjectProgress(id)
    const keepId = selectedMilestoneId.value
    progress.value = data
    if (keepId && data.milestones?.some((m) => m.id === keepId)) {
      selectedMilestoneId.value = keepId
    }
  } catch (e) {
    ElMessage.error(e.message || '更新失败')
    await load()
  }
}

function onStatusChange(row, status) {
  patchItem(row, { status })
}

function onAssigneeChange(row, assigneeId) {
  if (assigneeId == null || assigneeId === '') {
    patchItem(row, { clear_assignee: true })
  } else {
    patchItem(row, { assignee_id: assigneeId })
  }
}

function onNoteChange(row, note) {
  patchItem(row, { note: note ?? '' })
}

async function onApplyHints() {
  try {
    await ElMessageBox.confirm(
      '将根据当前缺口，把相关已完成检查项改回「待办」。不会自动勾选完成项。',
      '根据缺口标为未完成',
      { type: 'warning' },
    )
  } catch {
    return
  }
  applyingHints.value = true
  try {
    const res = await applyProgressHints(projectStore.currentIdOrDefault)
    progress.value = res.progress
    ElMessage.success(res.updated ? `已更新 ${res.updated} 项` : '没有需要改回的已完成项')
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    applyingHints.value = false
  }
}

watch(() => projectStore.currentId, () => load())

onMounted(load)
</script>

<style scoped>
.overview { margin-bottom: 16px; }
.overview-top {
  display: flex;
  gap: 32px;
  align-items: center;
  margin-bottom: 16px;
}
.overview-main { text-align: center; min-width: 88px; }
.pct { font-size: 36px; font-weight: 700; line-height: 1.1; color: #303133; }
.pct-label { color: #909399; font-size: 13px; margin-top: 4px; }
.overview-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 20px 28px;
  flex: 1;
}
.meta-label {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
.meta-value { color: #606266; font-size: 13px; }

.hints-card { margin-bottom: 16px; }
.hints-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.hints-list { display: flex; flex-direction: column; gap: 8px; }
.hint-item :deep(.el-alert__description) { margin-top: 2px; }
.hint-item-wrap {
  cursor: pointer;
}
.hint-item-wrap:hover {
  filter: brightness(0.97);
}
.hint-cat {
  font-size: 12px;
  color: #909399;
}

.stages-card { margin-bottom: 16px; }
.stage-axis {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}
.stage-chip {
  flex: 0 0 auto;
  min-width: 108px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 8px;
  padding: 10px 12px;
  cursor: pointer;
  text-align: left;
  transition: border-color .15s, box-shadow .15s;
}
.stage-chip:hover { border-color: #409eff; }
.stage-chip.active {
  border-color: #409eff;
  box-shadow: 0 0 0 1px #409eff inset;
}
.stage-chip.done { background: #f0f9eb; border-color: #b3e19d; }
.stage-chip.blocked { background: #fef0f0; border-color: #fbc4c4; }
.stage-chip.doing { background: #ecf5ff; }
.stage-title { font-size: 13px; font-weight: 600; color: #303133; }
.stage-count { margin-top: 4px; font-size: 12px; color: #909399; }

.items-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
