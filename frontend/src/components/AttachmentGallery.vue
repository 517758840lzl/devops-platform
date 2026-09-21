<template>
  <div v-if="items.length" class="gallery">
    <div v-for="a in items" :key="a.id" class="item" @click="preview(a)">
      <video v-if="a.media_type === 'video'" :src="a.url" muted />
      <img v-else :src="a.url" :alt="a.original_name" />
      <span class="name">{{ a.original_name }}</span>
    </div>

    <el-dialog v-model="visible" title="附件预览" width="720px">
      <video v-if="current?.media_type === 'video'" :src="current.url" controls style="width:100%;max-height:480px" />
      <img v-else-if="current" :src="current.url" style="width:100%;max-height:480px;object-fit:contain" :alt="current.original_name" />
    </el-dialog>
  </div>
  <el-empty v-else description="暂无附件" :image-size="48" />
</template>

<script setup>
import { ref, watch } from 'vue'
import { listAttachments } from '../api'

const props = defineProps({
  targetType: { type: String, required: true },
  targetId: { type: Number, required: true },
})

const items = ref([])
const visible = ref(false)
const current = ref(null)

async function load() {
  if (!props.targetId) return
  items.value = await listAttachments({ target_type: props.targetType, target_id: props.targetId })
}

function preview(a) {
  current.value = a
  visible.value = true
}

watch(() => props.targetId, load, { immediate: true })
defineExpose({ reload: load })
</script>

<style scoped>
.gallery { display: flex; flex-wrap: wrap; gap: 10px; margin: 12px 0; }
.item {
  width: 100px; cursor: pointer; border: 1px solid #ebeef5; border-radius: 6px; overflow: hidden;
}
.item img, .item video { width: 100%; height: 72px; object-fit: cover; display: block; }
.name { display: block; font-size: 13px; padding: 6px; color: #606266; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
