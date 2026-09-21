<template>
  <div class="media-upload">
    <el-upload
      v-model:file-list="fileList"
      :auto-upload="autoUpload"
      :http-request="customUpload"
      :before-upload="beforeUpload"
      list-type="picture-card"
      accept="image/*,video/*"
      multiple
    >
      <el-icon><Plus /></el-icon>
      <template #file="{ file }">
        <div class="upload-item">
          <video v-if="isVideoFile(file)" :src="file.url" class="thumb" muted />
          <img v-else :src="file.url" class="thumb" alt="" />
          <span class="upload-actions">
            <el-icon @click.stop="preview(file)"><ZoomIn /></el-icon>
            <el-icon @click.stop="handleRemove(file)"><Delete /></el-icon>
          </span>
        </div>
      </template>
    </el-upload>
    <div class="hint">支持 jpg/png/gif/webp（≤10MB）、mp4/mov/webm（≤100MB）</div>

    <el-dialog v-model="previewVisible" title="预览" width="720px">
      <video v-if="previewType === 'video'" :src="previewUrl" controls style="width:100%;max-height:480px" />
      <img v-else :src="previewUrl" style="width:100%;max-height:480px;object-fit:contain" alt="" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { deleteAttachment, uploadAttachment } from '../api'

const props = defineProps({
  targetType: { type: String, default: '' },
  targetId: { type: Number, default: 0 },
  autoUpload: { type: Boolean, default: false },
})

const emit = defineEmits(['uploaded'])

const fileList = ref([])
const previewVisible = ref(false)
const previewUrl = ref('')
const previewType = ref('image')

function isVideoFile(file) {
  const name = file.name || file.original_name || ''
  const type = file.raw?.type || file.mime_type || ''
  return type.startsWith('video/') || /\.(mp4|mov|webm)$/i.test(name)
}

function beforeUpload(file) {
  const isImage = file.type.startsWith('image/')
  const isVideo = file.type.startsWith('video/')
  if (!isImage && !isVideo) {
    ElMessage.error('仅支持图片或视频')
    return false
  }
  const max = isVideo ? 100 * 1024 * 1024 : 10 * 1024 * 1024
  if (file.size > max) {
    ElMessage.error(isVideo ? '视频不能超过 100MB' : '图片不能超过 10MB')
    return false
  }
  if (!props.autoUpload) {
    file.url = URL.createObjectURL(file)
  }
  return true
}

async function customUpload(options) {
  try {
    const data = await uploadAttachment(options.file, props.targetType, props.targetId)
    options.file.url = data.url
    options.file.attachmentId = data.id
    emit('uploaded', data)
    options.onSuccess(data)
  } catch (e) {
    options.onError(e)
  }
}

async function handleRemove(file) {
  if (file.attachmentId) {
    await deleteAttachment(file.attachmentId)
  }
  fileList.value = fileList.value.filter((f) => f.uid !== file.uid)
}

function preview(file) {
  previewUrl.value = file.url
  previewType.value = isVideoFile(file) ? 'video' : 'image'
  previewVisible.value = true
}

async function uploadPending(targetType, targetId) {
  const uploaded = []
  for (const item of fileList.value) {
    if (item.attachmentId || !item.raw) continue
    const data = await uploadAttachment(item.raw, targetType, targetId)
    uploaded.push(data)
  }
  return uploaded
}

function reset() {
  fileList.value = []
}

defineExpose({ uploadPending, reset })
</script>

<style scoped>
.hint { font-size: 12px; color: #909399; margin-top: 8px; }
.upload-item { width: 100%; height: 100%; position: relative; }
.thumb { width: 100%; height: 100%; object-fit: cover; }
.upload-actions {
  position: absolute; inset: 0; display: none; align-items: center; justify-content: center;
  gap: 12px; background: rgba(0,0,0,0.45); color: #fff; font-size: 18px; cursor: pointer;
}
.upload-item:hover .upload-actions { display: flex; }
</style>
