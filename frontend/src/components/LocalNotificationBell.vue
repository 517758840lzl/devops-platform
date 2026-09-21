<template>
  <el-popover placement="bottom-end" :width="360" trigger="click" @show="onOpen">
    <template #reference>
      <el-badge :value="unreadCount" :hidden="!unreadCount" :max="99" class="bell-badge">
        <el-button circle plain class="bell-btn" title="本地通知">
          <el-icon :size="18"><Bell /></el-icon>
        </el-button>
      </el-badge>
    </template>

    <div class="panel">
      <div class="panel-header">
        <span class="panel-title">本地通知</span>
        <div class="panel-actions">
          <el-button v-if="unreadCount" link type="primary" size="small" @click="markAllRead">全部已读</el-button>
          <el-button v-if="items.length" link type="danger" size="small" @click="clearAll">清空</el-button>
        </div>
      </div>
      <div class="panel-hint">仅保存在本浏览器，换设备或清缓存后会消失</div>

      <div v-if="!items.length" class="empty">暂无通知</div>
      <div v-else class="list">
        <div
          v-for="item in items"
          :key="item.id"
          class="item"
          :class="{ unread: !item.read }"
          @click="openItem(item)"
        >
          <div class="item-top">
            <el-tag size="small" :type="typeTag(item.type)">{{ typeLabel(item.type) }}</el-tag>
            <span class="time">{{ formatTime(item.createdAt) }}</span>
          </div>
          <div class="item-title">{{ item.title }}</div>
          <div class="item-msg">{{ item.message }}</div>
        </div>
      </div>
    </div>
  </el-popover>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { useLocalNotificationStore } from '../stores/localNotifications'

const router = useRouter()
const store = useLocalNotificationStore()

const items = computed(() => store.recentItems)
const unreadCount = computed(() => store.unreadCount)

function typeLabel(type) {
  return { build: '构建', bug: 'Bug', release: '发布' }[type] || type
}

function typeTag(type) {
  if (type === 'build') return 'success'
  if (type === 'bug') return 'danger'
  if (type === 'release') return 'warning'
  return 'info'
}

function formatTime(v) {
  if (!v) return ''
  const d = new Date(v)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return d.toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function markAllRead() {
  store.markAllRead()
}

function clearAll() {
  store.clearAll()
}

function onOpen() {
  // 打开面板时不自动全部已读，点击单条才标记
}

function openItem(item) {
  store.markRead(item.id)
  if (item.link) router.push(item.link)
}
</script>

<style scoped>
.bell-badge { margin-right: 4px; }
.bell-btn { border: none; }
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.panel-title { font-weight: 600; font-size: 15px; color: #303133; }
.panel-actions { display: flex; gap: 4px; }
.panel-hint { font-size: 12px; color: #909399; margin-bottom: 10px; }
.empty { text-align: center; color: #909399; padding: 24px 0; font-size: 13px; }
.list { max-height: 360px; overflow-y: auto; }
.item {
  padding: 10px 8px;
  border-radius: 6px;
  cursor: pointer;
  border-bottom: 1px solid #f0f2f5;
}
.item:last-child { border-bottom: none; }
.item:hover { background: #f5f7fa; }
.item.unread { background: #ecf5ff; }
.item-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.time { font-size: 12px; color: #909399; }
.item-title { font-size: 14px; font-weight: 500; color: #303133; }
.item-msg { font-size: 12px; color: #606266; margin-top: 2px; line-height: 1.4; }
</style>
