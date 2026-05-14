<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { commentApi } from '../api/comment'
import { useAuthStore } from '../stores/auth'
import CommentItem from './CommentItem.vue'
import type { Comment } from '../types'

const props = defineProps<{ videoId: number }>()
const auth = useAuthStore()

const comments = ref<Comment[]>([])
const content = ref('')
const loading = ref(false)
const error = ref('')
const publishLoading = ref(false)

async function fetchComments() {
  loading.value = true
  try {
    const res = await commentApi.getAll({ video_id: props.videoId })
    comments.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载评论失败'
  } finally {
    loading.value = false
  }
}

async function handlePublish() {
  if (!content.value.trim()) return
  publishLoading.value = true
  error.value = ''
  try {
    await commentApi.publish({ video_id: props.videoId, content: content.value.trim() })
    content.value = ''
    await fetchComments()
  } catch (e: any) {
    error.value = e.response?.data?.error || '发表评论失败'
  } finally {
    publishLoading.value = false
  }
}

async function handleDelete(commentId: number) {
  try {
    await commentApi.delete({ comment_id: commentId })
    await fetchComments()
  } catch (e: any) {
    error.value = e.response?.data?.error || '删除失败'
  }
}

onMounted(fetchComments)
</script>

<template>
  <div class="comment-section">
    <h3 class="section-title">评论 ({{ comments.length }})</h3>

    <div v-if="auth.isLoggedIn" class="comment-form">
      <textarea
        v-model="content"
        class="form-input comment-input"
        placeholder="说点什么..."
        rows="2"
      ></textarea>
      <button
        class="btn btn-primary btn-sm"
        :disabled="publishLoading || !content.trim()"
        @click="handlePublish"
      >
        {{ publishLoading ? '发送中...' : '发表' }}
      </button>
    </div>
    <p v-else class="comment-login-hint">
      <router-link to="/login">登录</router-link>后即可评论
    </p>

    <p v-if="error" class="error-text">{{ error }}</p>

    <div v-if="loading" class="comment-loading">
      <span class="spinner"></span>
    </div>

    <div v-else-if="comments.length === 0" class="comment-empty">
      暂无评论
    </div>

    <div v-else class="comment-list">
      <CommentItem
        v-for="c in comments"
        :key="c.id"
        :comment="c"
        @delete="handleDelete"
      />
    </div>
  </div>
</template>

<style scoped>
.comment-section {
  margin-top: 24px;
}

.section-title {
  font-size: 16px;
  margin-bottom: 12px;
}

.comment-form {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.comment-input {
  flex: 1;
  resize: vertical;
}

.comment-login-hint {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: 12px;
}
.comment-login-hint a {
  color: var(--color-primary);
}

.comment-loading {
  display: flex;
  justify-content: center;
  padding: 16px 0;
}

.comment-empty {
  text-align: center;
  color: var(--color-text-secondary);
  font-size: 14px;
  padding: 16px 0;
}

.comment-list {
  margin-top: 12px;
}
</style>
