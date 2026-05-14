<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const confirm = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  error.value = ''
  if (!username.value.trim()) {
    error.value = '请输入用户名'
    return
  }
  if (password.value.length < 6) {
    error.value = '密码至少6个字符'
    return
  }
  if (password.value !== confirm.value) {
    error.value = '两次密码不一致'
    return
  }
  loading.value = true
  try {
    await auth.register(username.value, password.value)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || e.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="container auth-page">
    <div class="auth-card">
      <h1 class="auth-title">注册</h1>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label class="form-label">用户名</label>
          <input
            v-model="username"
            class="form-input"
            type="text"
            placeholder="请输入用户名"
            autocomplete="username"
          />
        </div>
        <div class="form-group">
          <label class="form-label">密码</label>
          <input
            v-model="password"
            class="form-input"
            type="password"
            placeholder="至少6个字符"
            autocomplete="new-password"
          />
        </div>
        <div class="form-group">
          <label class="form-label">确认密码</label>
          <input
            v-model="confirm"
            class="form-input"
            type="password"
            placeholder="再次输入密码"
            autocomplete="new-password"
          />
        </div>
        <p v-if="error" class="error-text">{{ error }}</p>
        <button class="btn btn-primary auth-btn" :disabled="loading" type="submit">
          <span v-if="loading" class="spinner"></span>
          <span v-else>注册</span>
        </button>
      </form>
      <p class="auth-footer">
        已有账号？<router-link to="/login" class="link">登录</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - var(--nav-height) - 32px);
  padding-top: 60px;
}

.auth-card {
  width: 100%;
  max-width: 380px;
  background: var(--color-surface);
  padding: 32px;
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.auth-title {
  font-size: 24px;
  margin-bottom: 24px;
  text-align: center;
}

.auth-btn {
  width: 100%;
  padding: 12px;
  margin-top: 8px;
  min-height: 42px;
}

.auth-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.link {
  color: var(--color-primary);
  font-weight: 500;
}
.link:hover {
  text-decoration: underline;
}
</style>
