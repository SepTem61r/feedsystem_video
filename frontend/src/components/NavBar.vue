<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const hideNav = computed(() => route.meta.hideNav === true)

function goUpload() {
  if (!auth.isLoggedIn) {
    router.push('/login')
  } else {
    router.push('/upload')
  }
}
</script>

<template>
  <nav v-if="!hideNav" class="navbar">
    <div class="navbar-inner container">
      <router-link to="/" class="logo">FeedVideo</router-link>
      <div class="navbar-actions">
        <button class="btn btn-primary btn-sm" @click="goUpload">发布</button>
        <template v-if="auth.isLoggedIn">
          <router-link
            :to="`/profile/${auth.currentUser?.id}`"
            class="nav-user"
          >
            {{ auth.currentUser?.username }}
          </router-link>
          <button class="btn btn-outline btn-sm" @click="auth.logout()">
            退出
          </button>
        </template>
        <template v-else>
          <router-link to="/login" class="btn btn-outline btn-sm">登录</router-link>
        </template>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--nav-height);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  z-index: 100;
}

.navbar-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.logo {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-primary);
}

.navbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.nav-user {
  font-size: 14px;
  font-weight: 500;
}
.nav-user:hover {
  color: var(--color-primary);
}
</style>
