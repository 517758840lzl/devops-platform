<template>
  <div class="config-file-editor">
    <div class="toolbar">
      <div class="toolbar-meta">
        <el-tag v-if="meta.exists" size="small" type="success">已上传</el-tag>
        <el-tag v-else size="small" type="info">未上传</el-tag>
        <template v-if="meta.exists">
          <span class="meta-text">{{ meta.original_name || '—' }}</span>
          <span v-if="meta.size" class="meta-text muted">{{ formatSize(meta.size) }}</span>
          <span v-if="meta.item_count != null" class="meta-text muted">{{ meta.item_count }} 条</span>
          <span v-if="meta.updated_at" class="meta-text muted">{{ formatTime(meta.updated_at) }}</span>
        </template>
      </div>
      <div v-if="!readonly" class="toolbar-actions">
        <el-upload :show-file-list="false" :auto-upload="false" :accept="accept" @change="onPick">
          <el-button size="small">上传文件</el-button>
        </el-upload>
        <el-button size="small" @click="expanded = !expanded">{{ expanded ? '收起' : '展开编辑' }}</el-button>
      </div>
      <div v-else class="toolbar-actions">
        <el-button size="small" @click="expanded = !expanded">{{ expanded ? '收起' : '查看内容' }}</el-button>
      </div>
    </div>

    <div v-show="expanded" class="editor-panel">
      <el-input
        v-model="draft"
        type="textarea"
        :rows="rows"
        :placeholder="placeholder"
        class="editor"
        resize="vertical"
        :readonly="readonly"
      />
      <div v-if="!readonly" class="footer">
        <el-button type="primary" size="small" :loading="saving" @click="savePaste">保存粘贴内容</el-button>
      </div>
    </div>

    <div v-if="hint" class="hint">{{ hint }}</div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getConfigFileContent, pasteConfigFile, uploadConfigFile } from '../api'

const props = defineProps({
  projectId: { type: Number, required: true },
  category: { type: String, required: true },
  locale: { type: String, default: '' },
  accept: { type: String, default: '.dart,.txt' },
  placeholder: { type: String, default: '' },
  hint: { type: String, default: '' },
  rows: { type: Number, default: 12 },
  defaultExpanded: { type: Boolean, default: true },
  readonly: { type: Boolean, default: false },
})

const emit = defineEmits(['saved'])

const meta = ref({ exists: false })
const draft = ref('')
const expanded = ref(props.defaultExpanded)
const saving = ref(false)

async function load() {
  if (!props.projectId) return
  const data = await getConfigFileContent(props.projectId, props.category, props.locale)
  meta.value = data
  if (data.exists) draft.value = data.content || ''
}

async function onPick(uploadFile) {
  const file = uploadFile.raw
  if (!file) return
  saving.value = true
  try {
    const record = await uploadConfigFile(props.projectId, props.category, file, props.locale)
    meta.value = { exists: true, ...record, content: draft.value }
    ElMessage.success('上传成功')
    expanded.value = true
    emit('saved', record)
    await load()
  } catch (e) {
    ElMessage.error(e.message || '上传失败')
  } finally {
    saving.value = false
  }
}

async function savePaste() {
  if (!draft.value.trim()) {
    ElMessage.warning('请先粘贴内容')
    return
  }
  saving.value = true
  try {
    const record = await pasteConfigFile(props.projectId, props.category, {
      content: draft.value,
      locale: props.locale,
    })
    meta.value = { exists: true, ...record, content: draft.value }
    ElMessage.success('保存成功')
    emit('saved', record)
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function formatSize(size) {
  if (!size) return ''
  if (size < 1024) return `${size} B`
  return `${(size / 1024).toFixed(1)} KB`
}

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}

watch(() => [props.projectId, props.category, props.locale], load, { immediate: true })

defineExpose({ reload: load })
</script>

<style scoped>
.config-file-editor {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  padding: 10px 14px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
}

.toolbar-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.meta-text {
  font-size: 13px;
  color: #303133;
}

.meta-text.muted {
  color: #909399;
  font-size: 12px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.editor-panel {
  padding: 12px 14px 14px;
}

.editor {
  width: 100%;
}

.editor :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 1.55;
  padding: 12px;
}

.footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.hint {
  padding: 8px 14px 12px;
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
  border-top: 1px solid #f0f2f5;
  background: #fafafa;
}
</style>
