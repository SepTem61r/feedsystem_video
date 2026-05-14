<script setup lang="ts">
import { computed } from 'vue'
import type { FeedVideoItem } from '../types'

const props = defineProps<{ video: FeedVideoItem }>()
const emit = defineEmits<{ click: [id: number] }>()

function formatCount(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return String(n)
}

// cover_url 为视频路径（旧数据）时跳过，其余非空即当图片加载
function isVideoPath(url: string): boolean {
  return url.includes('/videos/')
}

const coverSrc = computed(() => {
  const u = props.video.cover_url
  if (!u || isVideoPath(u)) return ''
  return u
})
</script>

<template>
  <div class="video-card" @click="emit('click', video.id)">
    <div class="card-cover">
      <img
        v-if="coverSrc"
        :src="coverSrc"
        :alt="video.title"
        loading="lazy"
        @error="($event.target as HTMLImageElement).style.display = 'none'"
      />
      <div v-if="!coverSrc" class="cover-placeholder">暂无封面</div>
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
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
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
