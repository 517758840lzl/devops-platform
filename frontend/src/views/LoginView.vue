<template>
  <div class="login-page">
    <el-card class="login-card" shadow="hover">
      <h2>研发交付管理平台</h2>
      <p class="sub">Vue + Go · 需求 · 任务 · Bug · 发布</p>
      <el-form :model="form" @submit.prevent="onSubmit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" size="large" show-password />
        </el-form-item>
        <el-button type="primary" size="large" style="width:100%" :loading="loading" native-type="submit">
          登录
        </el-button>
      </el-form>
      <p class="hint">admin/admin123 · dev/dev123 · tester/test123</p>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: 'admin', password: 'admin123' })

async function onSubmit() {
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1d2b3a 0%, #2c5364 100%);
}
.login-card {
  width: 400px;
  padding: 8px 12px 16px;
}
h2 { margin: 0 0 8px; text-align: center; }
.sub { text-align: center; color: #909399; margin: 0 0 24px; font-size: 13px; }
.hint { text-align: center; color: #909399; font-size: 12px; margin-top: 16px; }
</style>
