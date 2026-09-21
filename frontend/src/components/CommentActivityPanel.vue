<template>
  <div class="panel">
    <el-tabs v-model="tab">
      <el-tab-pane label="评论" name="comments">
        <div class="comment-list">
          <div v-for="c in comments" :key="c.id" class="comment-item">
            <div class="meta">
              <strong>{{ c.user_name }}</strong>
              <span>{{ formatTime(c.created_at) }}</span>
            </div>
            <div class="content">{{ c.content }}</div>
          </div>
          <el-empty v-if="!comments.length" description="暂无评论" :image-size="60" />
        </div>
        <div class="comment-input">
          <el-input v-model="newComment" type="textarea" :rows="2" placeholder="添加评论..." />
          <el-button type="primary" size="small" style="margin-top:8px" :loading="submitting" @click="submitComment">发送</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="操作日志" name="activities">
        <el-timeline>
          <el-timeline-item v-for="a in activities" :key="a.id" :timestamp="formatTime(a.created_at)" placement="top">
            <strong>{{ a.user_name }}</strong>
            {{ actionLabel(a) }}
          </el-timeline-item>
        </el-timeline>
        <el-empty v-if="!activities.length" description="暂无日志" :image-size="60" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { createComment, listActivities, listComments } from '../api'
import { formatActivityAction } from '../utils/activityLabel'

const props = defineProps({
  targetType: { type: String, required: true },
  targetId: { type: Number, required: true },
})

const tab = ref('comments')
const comments = ref([])
const activities = ref([])
const newComment = ref('')
const submitting = ref(false)

function formatTime(v) {
  return v ? new Date(v).toLocaleString('zh-CN') : '-'
}

function actionLabel(a) {
  return formatActivityAction(a)
}

async function load() {
  if (!props.targetId) return
  const params = { target_type: props.targetType, target_id: props.targetId }
  comments.value = await listComments(params)
  activities.value = await listActivities(params)
}

async function submitComment() {
  if (!newComment.value.trim()) return
  submitting.value = true
  try {
    await createComment({
      target_type: props.targetType,
      target_id: props.targetId,
      content: newComment.value.trim(),
    })
    newComment.value = ''
    await load()
  } finally {
    submitting.value = false
  }
}

watch(() => props.targetId, load, { immediate: true })
</script>

<style scoped>
.panel { min-height: 200px; }
.comment-list { max-height: 240px; overflow-y: auto; margin-bottom: 12px; }
.comment-item { padding: 8px 0; border-bottom: 1px solid #ebeef5; }
.meta { display: flex; justify-content: space-between; font-size: 12px; color: #909399; margin-bottom: 4px; }
.content { font-size: 14px; color: #303133; }
</style>
