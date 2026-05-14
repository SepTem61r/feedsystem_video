<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { videoApi } from '../api/video'
import { socialApi } from '../api/social'
import { accountApi } from '../api/account'
import VideoGrid from '../components/VideoGrid.vue'
import type { Video, AccountInfo } from '../types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const profileUser = ref<AccountInfo | null>(null)
const videos = ref<Video[]>([])
const loading = ref(true)
const error = ref('')
const isFollowing = ref(false)
const followLoading = ref(false)
const showRename = ref(false)
const renameText = ref('')
const renameLoading = ref(false)
const renameError = ref('')

function profileId(): number {
  return Number(route.params.id)
}

const isSelf = computed(() => auth.currentUser?.id === profileId())

async function fetchProfile() {
  loading.value = true
  try {
    const [userRes, videoRes] = await Promise.all([
      accountApi.findByID({ id: profileId() }),
      videoApi.listByAuthorID({ author_id: profileId() }),
    ])
    profileUser.value = userRes.data
    videos.value = videoRes.data
    if (auth.isLoggedIn && !isSelf.value) {
      checkFollowStatus()
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载失败'
  } finally {
    loading.value = false
  }
}

async function checkFollowStatus() {
  try {
    const res = await socialApi.getAllVloggers({ follower_id: auth.currentUser!.id })
    isFollowing.value = res.data.vloggers?.some(v => v.id === profileId()) ?? false
  } catch { /* ignore */ }
}

async function toggleFollow() {
  followLoading.value = true
  try {
    if (isFollowing.value) {
      await socialApi.unfollow({ vlogger_id: profileId() })
      isFollowing.value = false
    } else {
      await socialApi.follow({ vlogger_id: profileId() })
      isFollowing.value = true
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '操作失败'
  } finally {
    followLoading.value = false
  }
}

async function handleRename() {
  if (!renameText.value.trim()) return
  renameLoading.value = true
  renameError.value = ''
  try {
    await auth.rename(renameText.value.trim())
    profileUser.value!.username = renameText.value.trim()
    showRename.value = false
    renameText.value = ''
  } catch (e: any) {
    renameError.value = e.response?.data?.error || '修改失败'
  } finally {
    renameLoading.value = false
  }
}

function goDetail(id: number) {
  router.push(`/video/${id}`)
}

onMounted(fetchProfile)
watch(() => route.params.id, () => {
  if (route.name === 'Profile') fetchProfile()
})
</script>

<template>
  <div class="container profile-page">
    <div v-if="loading" class="profile-loading">
      <span class="spinner"></span>
    </div>

    <div v-else-if="error" class="error-text profile-error">{{ error }}</div>

    <template v-else-if="profileUser">
      <div class="profile-header">
        <div class="profile-avatar">
          {{ profileUser.username.charAt(0).toUpperCase() }}
        </div>
        <div class="profile-info">
          <h1 class="profile-username">@{{ profileUser.username }}</h1>
          <p class="profile-id">ID: {{ profileUser.id }}</p>
        </div>
        <div class="profile-actions">
          <button
            v-if="!isSelf && auth.isLoggedIn"
            class="btn"
            :class="isFollowing ? 'btn-outline' : 'btn-primary'"
            :disabled="followLoading"
            @click="toggleFollow"
          >
            {{ isFollowing ? '已关注' : '+ 关注' }}
          </button>
          <button
            v-if="isSelf"
            class="btn btn-outline btn-sm"
            @click="showRename = !showRename"
          >
            修改用户名
          </button>
        </div>
      </div>

      <div v-if="showRename" class="rename-form">
        <input
          v-model="renameText"
          class="form-input"
          placeholder="新的用户名"
        />
        <button
          class="btn btn-primary btn-sm"
          :disabled="renameLoading || !renameText.trim()"
          @click="handleRename"
        >
          {{ renameLoading ? '保存中...' : '保存' }}
        </button>
        <button class="btn btn-outline btn-sm" @click="showRename = false">取消</button>
        <p v-if="renameError" class="error-text">{{ renameError }}</p>
      </div>

      <h2 class="section-heading">发布的视频 ({{ videos.length }})</h2>

      <VideoGrid :videos="videos" @click="goDetail" />
    </template>
  </div>
</template>

<style scoped>
.profile-page {
  padding-top: 24px;
  padding-bottom: 32px;
}

.profile-loading {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.profile-error {
  text-align: center;
  padding: 40px 0;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  margin-bottom: 24px;
}

.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
}

.profile-username {
  font-size: 20px;
}

.profile-id {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.profile-actions {
  flex-shrink: 0;
}

.rename-form {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.rename-form .form-input {
  max-width: 200px;
}

.section-heading {
  font-size: 16px;
  margin-bottom: 14px;
}
</style>
