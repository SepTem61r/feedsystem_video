<script setup lang="ts">
import { computed } from 'vue'
import type { Comment } from '../types'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{ comment: Comment }>()
const emit = defineEmits<{ delete: [id: number] }>()

const auth = useAuthStore()
const isOwner = computed(() => auth.currentUser?.username === props.comment.username)

function formatTime(iso: string): string {
  const d = new Date(iso)
  const now = Date.now()
  const diff = now - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  return d.toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="comment-item">
    <div class="comment-head">
      <span class="comment-user">@{{ comment.username }}</span>
      <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
    </div>
    <p class="comment-content">{{ comment.content }}</p>
    <button
      v-if="isOwner"
      class="btn btn-sm btn-outline comment-del"
      @click="emit('delete', comment.id)"
    >
      删除
    </button>
  </div>
</template>

<style scoped>
.comment-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}
.comment-item:last-child {
  border-bottom: none;
}

.comment-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.comment-user {
  font-size: 13px;
  font-weight: 600;
}

.comment-time {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.comment-content {
  font-size: 14px;
  line-height: 1.5;
}

.comment-del {
  margin-top: 6px;
}
</style>
