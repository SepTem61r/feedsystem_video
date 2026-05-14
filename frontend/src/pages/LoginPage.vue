<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  if (!username.value.trim() || !password.value.trim()) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: any) {
    error.value = e.response?.data?.error || e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="container auth-page">
    <div class="auth-card">
      <h1 class="auth-title">登录</h1>
      <form @submit.prevent="handleLogin">
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
            placeholder="请输入密码"
            autocomplete="current-password"
          />
        </div>
        <p v-if="error" class="error-text">{{ error }}</p>
        <button class="btn btn-primary auth-btn" :disabled="loading" type="submit">
          <span v-if="loading" class="spinner"></span>
          <span v-else>登录</span>
        </button>
      </form>
      <p class="auth-footer">
        还没有账号？<router-link to="/register" class="link">注册</router-link>
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
