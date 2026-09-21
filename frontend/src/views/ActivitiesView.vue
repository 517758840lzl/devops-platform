<template>
  <div>
    <div class="page-header">
      <h2>
        操作日志
        <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag>
      </h2>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-alert type="info" show-icon :closable="false" style="margin-bottom:12px">
      汇总当前项目内所有成员的操作（配置、构建、Bug、发布、成员、素材、进度等）
    </el-alert>

    <div class="filters">
      <el-select v-model="filters.userId" clearable placeholder="操作人" style="width: 160px" @change="load">
        <el-option
          v-for="m in members"
          :key="m.user_id"
          :label="`${m.name} (${m.username})`"
          :value="m.user_id"
        />
      </el-select>
      <el-select v-model="filters.scopeType" clearable placeholder="对象类型" style="width: 140px" @change="load">
        <el-option label="项目" value="project" />
        <el-option label="Bug/任务" value="issue" />
        <el-option label="发布" value="release" />
      </el-select>
      <el-select v-model="filters.actionGroup" clearable placeholder="操作分类" style="width: 140px" @change="onActionGroupChange">
        <el-option label="构建" value="build" />
        <el-option label="配置" value="config" />
        <el-option label="成员" value="member" />
        <el-option label="发布" value="release" />
        <el-option label="Bug/任务" value="bug" />
        <el-option label="上架素材" value="asset" />
        <el-option label="项目进度" value="progress" />
      </el-select>
      <el-select v-model="filters.action" clearable placeholder="具体操作" style="width: 160px" @change="load">
        <el-option v-for="a in actionOptions" :key="a.value" :label="a.label" :value="a.value" />
      </el-select>
      <el-date-picker
        v-model="filters.dateRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        style="width: 260px"
        @change="load"
      />
      <el-input
        v-model="filters.keyword"
        clearable
        placeholder="关键词（操作内容）"
        style="width: 200px"
        @clear="load"
        @keyup.enter="load"
      />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <el-table v-loading="loading" :data="items" stripe>
      <el-table-column label="时间" width="180">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作人" width="120">
        <template #default="{ row }">{{ row.user_name || `#${row.user_id}` }}</template>
      </el-table-column>
      <el-table-column label="对象" min-width="260" show-overflow-tooltip>
        <template #default="{ row }">{{ formatActivityTarget(row) }}</template>
      </el-table-column>
      <el-table-column label="操作" min-width="280">
        <template #default="{ row }">{{ formatActivityAction(row) }}</template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && !items.length" description="暂无操作日志" />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { listActivities, listMembers } from '../api'
import { useProjectStore } from '../stores/project'
import { formatActivityAction, formatActivityTarget } from '../utils/activityLabel'

const projectStore = useProjectStore()
const loading = ref(false)
const items = ref([])
const members = ref([])

const filters = reactive({
  userId: null,
  scopeType: '',
  actionGroup: '',
  action: '',
  dateRange: null,
  keyword: '',
})

const allActions = [
  { value: 'trigger_build', label: '触发构建', group: 'build' },
  { value: 'build_success', label: '构建成功', group: 'build' },
  { value: 'build_failed', label: '构建失败', group: 'build' },
  { value: 'delete_build_artifacts', label: '删除产物', group: 'build' },
  { value: 'update_settings', label: '更新项目配置', group: 'config' },
  { value: 'update_config_file', label: '更新配置文件', group: 'config' },
  { value: 'add_member', label: '添加成员', group: 'member' },
  { value: 'update_member_role', label: '变更角色', group: 'member' },
  { value: 'remove_member', label: '移除成员', group: 'member' },
  { value: 'create', label: '创建', group: 'release' },
  { value: 'submit_approval', label: '提交审批', group: 'release' },
  { value: 'approve', label: '审批通过', group: 'release' },
  { value: 'reject', label: '驳回发布', group: 'release' },
  { value: 'publish', label: '执行发布', group: 'release' },
  { value: 'status_change', label: '状态变更', group: 'bug' },
  { value: 'comment', label: '评论', group: 'bug' },
  { value: 'upload_attachment', label: '上传附件', group: 'bug' },
  { value: 'upload_store_asset', label: '上传素材', group: 'asset' },
  { value: 'delete_store_asset', label: '删除素材', group: 'asset' },
  { value: 'update_progress', label: '更新进度', group: 'progress' },
  { value: 'apply_progress_hints', label: '应用缺口提醒', group: 'progress' },
]

const actionOptions = computed(() => {
  if (!filters.actionGroup) return allActions
  return allActions.filter((a) => a.group === filters.actionGroup)
})

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function onActionGroupChange() {
  if (filters.action && !actionOptions.value.some((a) => a.value === filters.action)) {
    filters.action = ''
  }
  load()
}

function resetFilters() {
  filters.userId = null
  filters.scopeType = ''
  filters.actionGroup = ''
  filters.action = ''
  filters.dateRange = null
  filters.keyword = ''
  load()
}

async function loadMeta() {
  const projectId = projectStore.currentIdOrDefault
  if (!projectId) {
    members.value = []
    return
  }
  members.value = await listMembers(projectId)
}

async function load() {
  const projectId = projectStore.currentIdOrDefault
  if (!projectId) {
    items.value = []
    return
  }
  loading.value = true
  try {
    const params = { project_id: projectId, limit: 500 }
    if (filters.userId) params.user_id = filters.userId
    if (filters.scopeType) params.scope_type = filters.scopeType
    if (filters.action) params.action = filters.action
    else if (filters.actionGroup) params.action_group = filters.actionGroup
    if (filters.keyword?.trim()) params.keyword = filters.keyword.trim()
    if (filters.dateRange?.length === 2) {
      params.from = filters.dateRange[0]
      params.to = filters.dateRange[1]
    }
    items.value = await listActivities(params)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadMeta()
  await load()
})
watch(() => projectStore.currentId, async () => {
  resetFilters()
  await loadMeta()
})
</script>

<style scoped>
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  align-items: center;
}
</style>
