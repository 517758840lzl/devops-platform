<template>
  <el-dialog v-model="visible" title="选择产物输出目录" width="560px" @open="onOpen">
    <div class="toolbar">
      <el-button size="small" :disabled="!parentPath" @click="goParent">上一级</el-button>
      <el-input v-model="currentPath" size="small" @keyup.enter="browse(currentPath)">
        <template #append>
          <el-button @click="browse(currentPath)">前往</el-button>
        </template>
      </el-input>
    </div>

    <div v-loading="loading" class="dir-list">
      <div
        v-for="d in entries"
        :key="d.path"
        class="dir-item"
        @click="browse(d.path)"
        @dblclick="select(d.path)"
      >
        <el-icon><Folder /></el-icon>
        <span>{{ d.name }}</span>
      </div>
      <el-empty v-if="!loading && !entries.length" description="无子目录" :image-size="48" />
    </div>

    <div class="selected">当前选择：<code>{{ currentPath }}</code></div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="confirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { browseDirs } from '../api'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  initialPath: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue', 'select'])

const visible = ref(false)
const loading = ref(false)
const currentPath = ref('')
const parentPath = ref('')
const entries = ref([])

watch(() => props.modelValue, (v) => { visible.value = v })
watch(visible, (v) => emit('update:modelValue', v))

async function browse(path) {
  loading.value = true
  try {
    const data = await browseDirs(path)
    currentPath.value = data.current
    parentPath.value = data.parent
    entries.value = data.entries || []
  } finally {
    loading.value = false
  }
}

function goParent() {
  if (parentPath.value) browse(parentPath.value)
}

function select(path) {
  currentPath.value = path
  confirm()
}

function confirm() {
  emit('select', currentPath.value)
  visible.value = false
}

function onOpen() {
  browse(props.initialPath || '')
}

defineExpose({ browse })
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.dir-list {
  height: 280px; overflow-y: auto; border: 1px solid #ebeef5;
  border-radius: 6px; padding: 4px 0;
}
.dir-item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; cursor: pointer;
}
.dir-item:hover { background: #f5f7fa; }
.selected { margin-top: 12px; font-size: 13px; color: #606266; }
code { background: #f5f7fa; padding: 2px 6px; border-radius: 4px; }
</style>
