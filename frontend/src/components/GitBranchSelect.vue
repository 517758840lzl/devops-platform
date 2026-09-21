<template>
  <div class="branch-row">
    <el-select
      v-model="innerValue"
      filterable
      allow-create
      default-first-option
      clearable
      :loading="loading"
      :disabled="disabled"
      placeholder="main"
      class="branch-select"
    >
      <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
    </el-select>
    <el-button :loading="loading" :disabled="disabled || !gitUrl?.trim()" @click="fetchBranches">
      选择分支
    </el-button>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { listGitBranches } from '../api'

const props = defineProps({
  modelValue: { type: String, default: 'main' },
  gitUrl: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue'])

const innerValue = ref(props.modelValue || 'main')
const branches = ref([])
const loading = ref(false)

watch(() => props.modelValue, (v) => {
  innerValue.value = v || 'main'
})

watch(innerValue, (v) => {
  emit('update:modelValue', v || 'main')
})

watch(() => props.gitUrl, () => {
  branches.value = []
})

async function fetchBranches() {
  const url = props.gitUrl?.trim()
  if (!url) return
  loading.value = true
  try {
    branches.value = await listGitBranches(url)
    if (branches.value.length && !innerValue.value) {
      innerValue.value = branches.value.includes('main') ? 'main' : branches.value[0]
    }
    ElMessage.success(`已拉取 ${branches.value.length} 个分支`)
  } catch (err) {
    // axios 拦截器已弹过后端 message，这里只补兜底
    const msg = err?.message || err?.response?.data?.message
    if (msg && !String(msg).includes('Request failed')) {
      ElMessage.error(msg)
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.branch-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}
.branch-select {
  flex: 1;
  min-width: 0;
}
</style>
