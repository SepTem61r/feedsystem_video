<script setup lang="ts">
import type { FeedVideoItem } from '../types'

const props = defineProps<{ video: FeedVideoItem }>()
const emit = defineEmits<{ click: [id: number] }>()

function formatCount(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return String(n)
}

function placeholderImage(): string {
  // inline SVG数据URI作为默认封面
  return 'data:image/svg+xml,' + encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="200" viewBox="0 0 320 200">
      <rect fill="#e0e0e0" width="320" height="200"/>
      <text fill="#999" font-size="16" text-anchor="middle" x="160" y="105">暂无封面</text>
    </svg>`
  )
}
</script>

<template>
  <div class="video-card" @click="emit('click', video.id)">
    <div class="card-cover">
      <img
        :src="video.cover_url || placeholderImage()"
        :alt="video.title"
        loading="lazy"
        @error="($event.target as HTMLImageElement).src = placeholderImage()"
      />
      <div class="card-likes">{{ formatCount(video.likes_count) }} 赞</div>
    </div>
    <div class="card-body">
      <h3 class="card-title">{{ video.title || '无标题' }}</h3>
      <p class="card-author">@{{ video.author.username }}</p>
    </div>
  </div>
</template>

<style scoped>
.video-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}
.video-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow);
}

.card-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  background: #e0e0e0;
}
.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-likes {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.6);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.card-body {
  padding: 10px 12px;
}

.card-title {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-author {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}
</style>
