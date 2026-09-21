<template>
  <div class="api-constants-panel">
    <el-alert type="info" show-icon :closable="false" style="margin-bottom:12px">
      维护混淆后的 API 路径映射，对应 <code>lib/core/constants/api_constants.dart</code>，构建前请与仓库代码保持一致
    </el-alert>
    <ConfigFileEditor
      v-if="projectId"
      :key="`api-map-${projectId}`"
      :project-id="projectId"
      category="api_constants"
      accept=".dart"
      placeholder="static const String login = '/primecl/system/auth/enter';"
      hint="每行一个 static const String 接口标识 = '混淆后路径'"
      :rows="12"
      :readonly="readonly"
      :default-expanded="defaultExpanded"
      @saved="refresh"
    />
    <el-table v-if="rows.length" :data="rows" size="small" stripe max-height="320" class="api-map-table">
      <el-table-column prop="key" label="接口标识" min-width="180" />
      <el-table-column prop="path" label="混淆后路径" min-width="260" show-overflow-tooltip />
    </el-table>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import ConfigFileEditor from './ConfigFileEditor.vue'
import { getConfigFilePreview } from '../api'

const props = defineProps({
  projectId: { type: Number, default: null },
  readonly: { type: Boolean, default: false },
  defaultExpanded: { type: Boolean, default: false },
})

const rows = ref([])

async function refresh() {
  if (!props.projectId) {
    rows.value = []
    return
  }
  try {
    const data = await getConfigFilePreview(props.projectId, 'api_constants')
    const items = data.items || {}
    rows.value = Object.entries(items)
      .map(([key, path]) => ({ key, path }))
      .sort((a, b) => a.key.localeCompare(b.key))
  } catch {
    rows.value = []
  }
}

watch(() => props.projectId, refresh, { immediate: true })

defineExpose({ refresh })
</script>

<style scoped>
.api-constants-panel { width: 100%; }
.api-map-table { margin-top: 12px; width: 100%; }
</style>
