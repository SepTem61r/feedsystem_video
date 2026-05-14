<script setup lang="ts">
import type { Video } from '../types'

const props = defineProps<{ videos: Video[] }>()
const emit = defineEmits<{ click: [id: number] }>()

function getCoverSrc(coverUrl: string): string {
  if (!coverUrl || coverUrl.includes('/videos/')) return ''
  return coverUrl
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
          v-if="getCoverSrc(v.cover_url)"
          :src="v.cover_url"
          :alt="v.title"
          loading="lazy"
          @error="($event.target as HTMLImageElement).style.display = 'none'"
        />
        <div v-if="!getCoverSrc(v.cover_url)" class="cover-placeholder">暂无封面</div>
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
  position: relative;
}
.grid-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 13px;
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
