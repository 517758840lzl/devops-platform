<template>
  <div>
    <div class="page-header">
      <h2>成员与权限 <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag></h2>
      <el-button v-if="canManage" type="primary" @click="openAdd">添加成员</el-button>
    </div>

    <el-alert type="info" show-icon :closable="false" style="margin-bottom:16px">
      角色说明：owner/超级管理员/admin 可管理成员、审批发布、删除上架素材；developer 可编辑项目配置；tester 可提交 Bug、查看配置；viewer 只读
    </el-alert>

    <el-table v-loading="loading" :data="members" stripe>
      <el-table-column prop="name" label="姓名" width="120" />
      <el-table-column prop="username" label="账号" width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="role" label="项目角色" width="160">
        <template #default="{ row }">
          <el-select v-model="row.role" size="small" :disabled="!canManage || row.role === 'owner'" @change="(v) => onRoleChange(row, v)">
            <el-option label="负责人 owner" value="owner" />
            <el-option label="超级管理员 super_admin" value="super_admin" />
            <el-option label="管理员 admin" value="admin" />
            <el-option label="开发 developer" value="developer" />
            <el-option label="测试 tester" value="tester" />
            <el-option label="只读 viewer" value="viewer" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="canManage && row.role !== 'owner'" link type="danger" @click="onRemove(row)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-card shadow="never" style="margin-top:16px">
      <template #header>项目操作日志</template>
      <CommentActivityPanel target-type="project" :target-id="projectStore.currentIdOrDefault" />
    </el-card>

    <el-dialog v-model="addVisible" title="添加成员" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户">
          <el-select v-model="form.user_id" filterable placeholder="选择用户">
            <el-option v-for="u in users" :key="u.id" :label="`${u.name} (${u.username})`" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="开发" value="developer" />
            <el-option label="测试" value="tester" />
            <el-option label="只读" value="viewer" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAdd">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import CommentActivityPanel from '../components/CommentActivityPanel.vue'
import { useProjectPermission } from '../composables/useProjectPermission'
import { addMember, listMembers, listUsers, removeMember, updateMember } from '../api'
import { useProjectStore } from '../stores/project'

const projectStore = useProjectStore()
const { canManage, refresh: refreshPermission } = useProjectPermission()
const loading = ref(false)
const members = ref([])
const users = ref([])
const addVisible = ref(false)
const form = reactive({ user_id: null, role: 'developer' })

async function load() {
  loading.value = true
  try {
    members.value = await listMembers(projectStore.currentIdOrDefault)
  } finally {
    loading.value = false
  }
}

function openAdd() {
  form.user_id = null
  form.role = 'developer'
  addVisible.value = true
}

async function submitAdd() {
  await addMember(projectStore.currentIdOrDefault, form)
  addVisible.value = false
  await load()
}

async function onRoleChange(row, role) {
  await updateMember(projectStore.currentIdOrDefault, row.id, { role })
}

async function onRemove(row) {
  await ElMessageBox.confirm(`确定移除 ${row.name}？`, '提示')
  await removeMember(projectStore.currentIdOrDefault, row.id)
  await load()
}

onMounted(async () => {
  await refreshPermission()
  users.value = await listUsers()
  await load()
})
watch(() => projectStore.currentId, async () => {
  await refreshPermission()
  await load()
})
</script>
