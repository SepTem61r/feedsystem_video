<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { videoApi } from '../api/video'
import { likeApi } from '../api/like'
import { socialApi } from '../api/social'
import VideoPlayer from '../components/VideoPlayer.vue'
import CommentSection from '../components/CommentSection.vue'
import type { Video } from '../types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const video = ref<Video | null>(null)
const loading = ref(true)
const error = ref('')
const isLiked = ref(false)
const likeLoading = ref(false)
const isFollowing = ref(false)
const followLoading = ref(false)

function videoId(): number {
  return Number(route.params.id)
}

async function fetchVideo() {
  loading.value = true
  try {
    const res = await videoApi.getDetail({ id: videoId() })
    video.value = res.data
    if (auth.isLoggedIn) {
      checkLikeStatus()
      checkFollowStatus()
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载失败'
  } finally {
    loading.value = false
  }
}

async function checkLikeStatus() {
  try {
    const res = await likeApi.isLiked({ video_id: videoId() })
    isLiked.value = res.data.is_liked
  } catch { /* ignore */ }
}

async function checkFollowStatus() {
  if (!video.value || video.value.author_id === auth.currentUser?.id) return
  try {
    const res = await socialApi.getAllVloggers({ follower_id: auth.currentUser!.id })
    isFollowing.value = res.data.vloggers?.some(v => v.id === video.value!.author_id) ?? false
  } catch { /* ignore */ }
}

async function toggleLike() {
  if (!auth.isLoggedIn) {
    router.push('/login')
    return
  }
  likeLoading.value = true
  try {
    if (isLiked.value) {
      await likeApi.unlike({ video_id: videoId() })
      isLiked.value = false
      if (video.value) video.value.likes_count = Math.max(0, video.value.likes_count - 1)
    } else {
      await likeApi.like({ video_id: videoId() })
      isLiked.value = true
      if (video.value) video.value.likes_count++
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '操作失败'
  } finally {
    likeLoading.value = false
  }
}

async function toggleFollow() {
  if (!auth.isLoggedIn || !video.value) return
  followLoading.value = true
  try {
    if (isFollowing.value) {
      await socialApi.unfollow({ vlogger_id: video.value.author_id })
      isFollowing.value = false
    } else {
      await socialApi.follow({ vlogger_id: video.value.author_id })
      isFollowing.value = true
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '操作失败'
  } finally {
    followLoading.value = false
  }
}

const deleteLoading = ref(false)

async function handleDelete() {
  if (!confirm('确定要删除这个视频吗？')) return
  deleteLoading.value = true
  try {
    await videoApi.delete({ id: videoId() })
    router.push('/')
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

function goProfile(id: number) {
  router.push(`/profile/${id}`)
}

onMounted(fetchVideo)
watch(() => route.params.id, () => {
  if (route.name === 'VideoDetail') fetchVideo()
})
</script>

<template>
  <div class="container detail-page">
    <div v-if="loading" class="detail-loading">
      <span class="spinner"></span>
    </div>

    <div v-else-if="error" class="error-text detail-error">{{ error }}</div>

    <template v-else-if="video">
      <VideoPlayer :src="video.play_url" :poster="video.cover_url" />

      <div class="detail-info">
        <h1 class="detail-title">{{ video.title || '无标题' }}</h1>
        <p class="detail-desc" v-if="video.description">{{ video.description }}</p>

        <div class="detail-meta">
          <span
            class="detail-author"
            @click="goProfile(video.author_id)"
          >
            @{{ video.username }}
          </span>
          <span class="detail-time">{{ new Date(video.create_time).toLocaleDateString('zh-CN') }}</span>
        </div>

        <div class="detail-actions">
          <button
            class="btn"
            :class="isLiked ? 'btn-primary' : 'btn-outline'"
            :disabled="likeLoading"
            @click="toggleLike"
          >
            {{ isLiked ? '❤️' : '🤍' }} {{ video.likes_count }}
          </button>
          <button
            v-if="video.author_id !== auth.currentUser?.id"
            class="btn btn-outline"
            :disabled="followLoading"
            @click="toggleFollow"
          >
            {{ isFollowing ? '已关注' : '+ 关注' }}
          </button>
          <button
            v-if="video.author_id === auth.currentUser?.id"
            class="btn btn-outline btn-danger"
            :disabled="deleteLoading"
            @click="handleDelete"
          >
            {{ deleteLoading ? '删除中...' : '删除视频' }}
          </button>
        </div>
      </div>

      <CommentSection :video-id="video.id" />
    </template>
  </div>
</template>

<style scoped>
.detail-page {
  padding-top: 16px;
  padding-bottom: 32px;
  max-width: 720px;
}

.detail-loading {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.detail-error {
  text-align: center;
  padding: 40px 0;
}

.detail-info {
  margin-top: 16px;
  padding: 16px;
  background: var(--color-surface);
  border-radius: var(--radius);
}

.detail-title {
  font-size: 20px;
  margin-bottom: 8px;
}

.detail-desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-bottom: 12px;
  line-height: 1.6;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: 14px;
}

.detail-author {
  font-weight: 600;
  color: var(--color-text);
  cursor: pointer;
}
.detail-author:hover {
  color: var(--color-primary);
}

.detail-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.btn-danger {
  border-color: var(--color-error);
  color: var(--color-error);
}
.btn-danger:hover {
  background: var(--color-error);
  color: #fff;
}
</style>
