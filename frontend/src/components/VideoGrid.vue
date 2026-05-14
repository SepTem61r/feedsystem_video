<script setup lang="ts">
import type { Video } from '../types'

defineProps<{ videos: Video[] }>()
const emit = defineEmits<{ click: [id: number] }>()

function placeholderImage(): string {
  return 'data:image/svg+xml,' + encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="200" viewBox="0 0 320 200">
      <rect fill="#e0e0e0" width="320" height="200"/>
      <text fill="#999" font-size="16" text-anchor="middle" x="160" y="105">暂无封面</text>
    </svg>`
  )
}
</script>

<template>
  <div v-if="videos.length === 0" class="grid-empty">
    暂无视频
  </div>
  <div v-else class="video-grid">
    <div
      v-for="v in videos"
      :key="v.id"
      class="grid-item"
      @click="emit('click', v.id)"
    >
      <div class="grid-cover">
        <img
          :src="v.cover_url || placeholderImage()"
          :alt="v.title"
          loading="lazy"
          @error="($event.target as HTMLImageElement).src = placeholderImage()"
        />
      </div>
      <p class="grid-title">{{ v.title || '无标题' }}</p>
      <p class="grid-likes">{{ v.likes_count }} 赞</p>
    </div>
  </div>
</template>

<style scoped>
.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px;
}

.grid-item {
  cursor: pointer;
  transition: transform 0.2s;
}
.grid-item:hover {
  transform: translateY(-2px);
}

.grid-cover {
  aspect-ratio: 16 / 10;
  border-radius: var(--radius);
  overflow: hidden;
  background: #e0e0e0;
}
.grid-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.grid-title {
  margin-top: 6px;
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.grid-likes {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.grid-empty {
  text-align: center;
  color: var(--color-text-secondary);
  padding: 32px 0;
}
</style>
