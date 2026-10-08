<template>
  <div class="store-assets">
    <div v-if="logoSlot" class="logo-block">
      <div class="block-title">{{ logoSlot.label }}</div>
      <div class="size-tip">{{ logoSlot.tip }}</div>
      <div class="logo-wrap">
        <div class="slot-card">
          <div class="slot-title">{{ logoSlot.label }}</div>
          <div
            class="preview logo-preview"
            :class="{ readonly: !canUpload }"
            :style="{ aspectRatio: logoSlot.aspect }"
            @click="canUpload && triggerUpload(logoSlot.key)"
          >
            <img v-if="assetMap[logoSlot.key]" :src="previewUrl(assetMap[logoSlot.key])" alt="" />
            <div v-else class="placeholder">
              <template v-if="canUpload">
                <el-icon size="28"><Plus /></el-icon>
                <span>上传</span>
              </template>
              <span v-else>暂无</span>
            </div>
          </div>
          <div v-if="assetMap[logoSlot.key]" class="slot-actions">
            <el-button link type="primary" size="small" @click="downloadOne(logoSlot.key)">下载</el-button>
            <el-button v-if="canDelete" link type="danger" size="small" @click="remove(logoSlot.key)">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="bannerSlot" class="banner-block">
      <div class="block-title">{{ bannerSlot.label }}</div>
      <div class="size-tip">{{ bannerSlot.tip }}</div>
      <div class="banner-wrap">
        <div class="slot-card">
          <div class="slot-title">{{ bannerSlot.label }}</div>
          <div
            class="preview"
            :class="{ readonly: !canUpload }"
            :style="{ aspectRatio: bannerSlot.aspect }"
            @click="canUpload && triggerUpload(bannerSlot.key)"
          >
            <img v-if="assetMap[bannerSlot.key]" :src="previewUrl(assetMap[bannerSlot.key])" alt="" />
            <div v-else class="placeholder">
              <template v-if="canUpload">
                <el-icon size="28"><Plus /></el-icon>
                <span>上传</span>
              </template>
              <span v-else>暂无</span>
            </div>
          </div>
          <div v-if="assetMap[bannerSlot.key]" class="slot-actions">
            <el-button link type="primary" size="small" @click="downloadOne(bannerSlot.key)">下载</el-button>
            <el-button v-if="canDelete" link type="danger" size="small" @click="remove(bannerSlot.key)">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="screenshotSlots.length" class="shots-block">
      <div class="shots-head">
        <div>
          <div class="block-title">应用截图</div>
          <div class="size-tip">{{ screenshotTip }}</div>
        </div>
        <el-button
          v-if="canUpload"
          size="small"
          type="primary"
          plain
          @click="triggerScreenshotUpload(screenshotSlots[0].key)"
        >
          一次选多张
        </el-button>
      </div>
      <div class="grid">
        <div v-for="slot in screenshotSlots" :key="slot.key" class="slot-card">
          <div class="slot-title">{{ slot.label }}</div>
          <div
            class="preview"
            :class="{ readonly: !canUpload }"
            :style="{ aspectRatio: slot.aspect }"
            @click="canUpload && triggerScreenshotUpload(slot.key)"
          >
            <img v-if="assetMap[slot.key]" :src="previewUrl(assetMap[slot.key])" alt="" />
            <div v-else class="placeholder">
              <template v-if="canUpload">
                <el-icon size="28"><Plus /></el-icon>
                <span>上传</span>
              </template>
              <span v-else>暂无</span>
            </div>
          </div>
          <div v-if="assetMap[slot.key]" class="slot-actions">
            <el-button link type="primary" size="small" @click="downloadOne(slot.key)">下载</el-button>
            <el-button v-if="canDelete" link type="danger" size="small" @click="remove(slot.key)">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      hidden
      :multiple="pickMultiple"
      @change="onFileChange"
    />
  </div>
</template>

<script setup>
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  deleteProjectStoreAsset,
  deleteReleaseStoreAsset,
  downloadProjectStoreAsset,
  listProjectStoreAssets,
  listReleaseStoreAssets,
  uploadProjectStoreAsset,
  uploadReleaseStoreAsset,
} from '../api'

const props = defineProps({
  projectId: { type: Number, default: null },
  releaseId: { type: Number, default: null },
  platform: { type: String, default: '' },
  canUpload: { type: Boolean, default: true },
  canDelete: { type: Boolean, default: false },
})

const emit = defineEmits(['change'])

// App Store Connect 官方 6.9" Display 接受尺寸（竖屏 + 横屏）
// https://developer.apple.com/help/app-store-connect/reference/app-information/screenshot-specifications/
const IOS_SCREENSHOT_ALLOWED_SIZES = [
  [1320, 2868], [2868, 1320],
  [1290, 2796], [2796, 1290],
  [1260, 2736], [2736, 1260],
]

const iosScreenshotSlots = [
  { key: 'screenshot_1', label: '图 1', aspect: '1320/2868', allowedSizes: IOS_SCREENSHOT_ALLOWED_SIZES },
  { key: 'screenshot_2', label: '图 2', aspect: '1320/2868', allowedSizes: IOS_SCREENSHOT_ALLOWED_SIZES },
  { key: 'screenshot_3', label: '图 3', aspect: '1320/2868', allowedSizes: IOS_SCREENSHOT_ALLOWED_SIZES },
  { key: 'screenshot_4', label: '图 4', aspect: '1320/2868', allowedSizes: IOS_SCREENSHOT_ALLOWED_SIZES },
  { key: 'screenshot_5', label: '图 5', aspect: '1320/2868', allowedSizes: IOS_SCREENSHOT_ALLOWED_SIZES },
]

const androidScreenshotSlots = [
  { key: 'screenshot_1', label: '图 1', aspect: '9/16', androidScreenshot: true },
  { key: 'screenshot_2', label: '图 2', aspect: '9/16', androidScreenshot: true },
  { key: 'screenshot_3', label: '图 3', aspect: '9/16', androidScreenshot: true },
  { key: 'screenshot_4', label: '图 4', aspect: '9/16', androidScreenshot: true },
  { key: 'screenshot_5', label: '图 5', aspect: '9/16', androidScreenshot: true },
]

const iosLogoSlot = {
  key: 'logo',
  label: 'App Icon',
  aspect: '1/1',
  tip: 'App Store 图标：1024×1024 正方形；不要自带圆角；不要透明/Alpha；PNG 或 JPEG',
  exactWidth: 1024,
  exactHeight: 1024,
}

const androidLogoSlot = {
  key: 'logo',
  label: 'App Icon',
  aspect: '1/1',
  tip: 'Google Play 图标：512×512 正方形 PNG（≤1MB）；不要自带圆角和阴影（商店会自动裁圆角加阴影）；可用透明底',
  exactWidth: 512,
  exactHeight: 512,
}

const androidBannerSlot = {
  key: 'banner',
  label: 'Feature Banner',
  aspect: '1024/500',
  tip: 'Google Play 宣传图，建议 1024×500 px，JPG/PNG（不建议透明）',
  exactWidth: 1024,
  exactHeight: 500,
}

const legacySlots = [
  { key: 'screenshot_1', label: '图 1', aspect: '9/16' },
  { key: 'screenshot_2', label: '图 2', aspect: '9/16' },
  { key: 'screenshot_3', label: '图 3', aspect: '9/16' },
  { key: 'screenshot_4', label: '图 4', aspect: '9/16' },
  { key: 'screenshot_5', label: '图 5', aspect: '9/16' },
]

const screenshotSlots = computed(() => {
  if (props.platform === 'ios') return iosScreenshotSlots
  if (props.platform === 'android') return androidScreenshotSlots
  return legacySlots
})

const logoSlot = computed(() => {
  if (props.platform === 'ios') return iosLogoSlot
  if (props.platform === 'android') return androidLogoSlot
  return null
})

const bannerSlot = computed(() => (props.platform === 'android' ? androidBannerSlot : null))

const screenshotTip = computed(() => {
  if (props.platform === 'ios') {
    return '可一次选多张，按点击槽位向后填入。须符合 App Store 6.9″ 官方尺寸：1320×2868 / 1290×2796 / 1260×2736（竖屏或对应横屏）'
  }
  if (props.platform === 'android') {
    return '可一次选多张，按点击槽位向后填入。Google Play 手机截图：9:16 或 16:9；边长 320–3840px；最短边 ≥1080；建议 1080×1920'
  }
  return '点击槽位可一次选多张，按顺序填入空位，支持 jpg/png/webp'
})

const allSlotKeys = computed(() => {
  const keys = []
  if (logoSlot.value) keys.push(logoSlot.value.key)
  if (bannerSlot.value) keys.push(bannerSlot.value.key)
  keys.push(...screenshotSlots.value.map((s) => s.key))
  return keys
})

function slotLabel(slot) {
  if (logoSlot.value?.key === slot) return logoSlot.value.label
  if (bannerSlot.value?.key === slot) return bannerSlot.value.label
  return screenshotSlots.value.find((s) => s.key === slot)?.label || slot
}

const assetMap = reactive({})
const fileInput = ref(null)
const pendingSlot = ref('')
const pickMultiple = ref(false)

function previewUrl(asset) {
  if (!asset?.url) return ''
  const t = asset.updated_at
    ? new Date(asset.updated_at).getTime()
    : (asset.created_at ? new Date(asset.created_at).getTime() : asset.id || Date.now())
  const sep = asset.url.includes('?') ? '&' : '?'
  return `${asset.url}${sep}t=${t}`
}

async function load() {
  if (!props.projectId && !props.releaseId) return
  let list
  if (props.projectId) {
    list = await listProjectStoreAssets(props.projectId, props.platform || undefined)
    if (props.platform) {
      list = list.filter((a) => a.platform === props.platform)
    }
  } else {
    list = await listReleaseStoreAssets(props.releaseId)
  }
  allSlotKeys.value.forEach((k) => { delete assetMap[k] })
  list.forEach((a) => { assetMap[a.slot] = a })
  emit('change', list.length)
}

async function triggerUpload(slot, multiple = false) {
  pendingSlot.value = slot
  pickMultiple.value = multiple
  await nextTick()
  fileInput.value?.click()
}

function triggerScreenshotUpload(slot) {
  return triggerUpload(slot, true)
}

function screenshotKeysFrom(startKey) {
  const keys = screenshotSlots.value.map((s) => s.key)
  const start = keys.indexOf(startKey)
  if (start < 0) return []
  return keys.slice(start)
}

function slotSpec(slotKey) {
  if (logoSlot.value?.key === slotKey) return logoSlot.value
  if (bannerSlot.value?.key === slotKey) return bannerSlot.value
  return screenshotSlots.value.find((s) => s.key === slotKey) || null
}

function readImageSize(file) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve({ width: img.naturalWidth, height: img.naturalHeight })
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法读取图片尺寸'))
    }
    img.src = url
  })
}

function isNearRatio(width, height, targetW, targetH, tolerance = 0.02) {
  const actual = width / height
  const expected = targetW / targetH
  return Math.abs(actual - expected) / expected <= tolerance
}

function validateAndroidScreenshot(width, height, label) {
  const minSide = Math.min(width, height)
  const maxSide = Math.max(width, height)
  if (minSide < 320 || maxSide > 3840) {
    ElMessage.error(
      `${label} 边长须在 320–3840px（当前 ${width}×${height}）`,
    )
    return false
  }
  if (maxSide > minSide * 2) {
    ElMessage.error(
      `${label} 最长边不能超过最短边的 2 倍（当前 ${width}×${height}）`,
    )
    return false
  }
  if (minSide < 1080) {
    ElMessage.error(
      `${label} 最短边须 ≥1080px（当前 ${width}×${height}）`,
    )
    return false
  }
  const portrait916 = isNearRatio(width, height, 9, 16)
  const landscape169 = isNearRatio(width, height, 16, 9)
  if (!portrait916 && !landscape169) {
    ElMessage.error(
      `${label} 须为 9:16 竖屏或 16:9 横屏（当前 ${width}×${height}）`,
    )
    return false
  }
  return true
}

async function validateImageSize(file, slotKey) {
  const spec = slotSpec(slotKey)
  if (!spec) return true
  const { width, height } = await readImageSize(file)

  if (spec.exactWidth && spec.exactHeight) {
    if (width !== spec.exactWidth || height !== spec.exactHeight) {
      ElMessage.error(
        `${spec.label} 尺寸须为 ${spec.exactWidth}×${spec.exactHeight}，当前为 ${width}×${height}`,
      )
      return false
    }
    return true
  }

  if (spec.allowedSizes?.length) {
    const ok = spec.allowedSizes.some(([w, h]) => w === width && h === height)
    if (!ok) {
      const examples = spec.allowedSizes
        .filter(([w, h]) => w <= h)
        .map(([w, h]) => `${w}×${h}`)
        .join(' / ')
      ElMessage.error(
        `${spec.label} 尺寸不符合 App Store 要求（当前 ${width}×${height}）。可用：${examples} 及其横屏`,
      )
      return false
    }
    return true
  }

  if (spec.androidScreenshot) {
    return validateAndroidScreenshot(width, height, spec.label)
  }
  return true
}

async function uploadOne(slot, file) {
  if (props.projectId) {
    if (!props.platform) {
      throw new Error('缺少 platform 参数')
    }
    await uploadProjectStoreAsset(props.projectId, props.platform, slot, file)
  } else {
    await uploadReleaseStoreAsset(props.releaseId, slot, file)
  }
}

async function onFileChange(e) {
  const files = [...(e.target.files || [])]
  const startSlot = pendingSlot.value
  e.target.value = ''
  if (!files.length || !startSlot) return

  const targets = pickMultiple.value
    ? screenshotKeysFrom(startSlot).slice(0, files.length)
    : [startSlot]
  if (pickMultiple.value && files.length > targets.length) {
    ElMessage.warning(`最多还能填 ${targets.length} 个槽位，已忽略多余文件`)
  }

  let ok = 0
  try {
    for (let i = 0; i < targets.length; i += 1) {
      const slot = targets[i]
      const file = files[i]
      if (!(await validateImageSize(file, slot))) continue
      await uploadOne(slot, file)
      ok += 1
    }
    if (ok) ElMessage.success(ok === 1 ? '上传成功' : `已上传 ${ok} 张`)
  } catch (err) {
    ElMessage.error(err.message || '上传失败')
  } finally {
    pickMultiple.value = false
    pendingSlot.value = ''
  }
  await load()
}

async function downloadOne(slot) {
  const asset = assetMap[slot]
  if (!asset) return
  try {
    if (props.projectId && asset.id) {
      await downloadProjectStoreAsset(props.projectId, asset.id, asset.original_name)
    } else if (asset.url) {
      const a = document.createElement('a')
      a.href = asset.url
      a.download = asset.original_name || slot
      a.target = '_blank'
      a.click()
    }
  } catch (err) {
    ElMessage.error(err.message || '下载失败')
  }
}

async function remove(slot) {
  const asset = assetMap[slot]
  if (!asset || !props.canDelete) return
  const name = asset.original_name || slotLabel(slot)
  try {
    await ElMessageBox.confirm(
      `确定删除「${name}」？删除后协同成员将无法再下载此素材。`,
      '删除上架素材',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    if (props.projectId) {
      await deleteProjectStoreAsset(props.projectId, asset.id)
    } else {
      await deleteReleaseStoreAsset(props.releaseId, asset.id)
    }
    ElMessage.success('已删除')
  } catch (err) {
    ElMessage.error(err.message || '删除失败')
    return
  }
  await load()
}

watch(() => [props.projectId, props.releaseId, props.platform], load, { immediate: true })
</script>

<style scoped>
.store-assets { width: 100%; }
.block-title { font-size: 14px; font-weight: 600; color: #303133; margin-bottom: 4px; }
.size-tip { font-size: 12px; color: #909399; margin-bottom: 10px; line-height: 1.5; }
.shots-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.shots-head .size-tip { margin-bottom: 0; }
.logo-block { margin-bottom: 20px; }
.logo-wrap { max-width: 160px; }
.banner-block { margin-bottom: 20px; }
.banner-wrap { max-width: 420px; }
.grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; }
.slot-card { text-align: center; }
.slot-title { font-size: 13px; margin-bottom: 6px; color: #606266; }
.preview {
  border: 1px dashed #dcdfe6; border-radius: 8px;
  overflow: hidden; cursor: pointer; background: #fafafa;
  display: flex; align-items: center; justify-content: center;
}
.preview.logo-preview {
  border-radius: 0;
  background: #fff;
}
.preview.readonly { cursor: default; }
.preview img { width: 100%; height: 100%; object-fit: cover; }
.placeholder {
  display: flex; flex-direction: column; align-items: center;
  color: #909399; font-size: 12px; gap: 4px;
}
.slot-actions { display: flex; justify-content: center; gap: 8px; margin-top: 4px; }
</style>
