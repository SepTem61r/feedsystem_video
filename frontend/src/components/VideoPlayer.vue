<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ src: string; poster?: string }>()

const validPoster = computed(() => {
  if (!props.poster) return undefined
  if (props.poster.includes('/videos/')) return undefined
  return props.poster
})
</script>

<template>
  <div class="player-wrapper">
    <video
      class="video-player"
      :src="src"
      :poster="validPoster"
      controls
      playsinline
      preload="metadata"
    >
      您的浏览器不支持视频播放
    </video>
  </div>
</template>

<style scoped>
.player-wrapper {
  background: #000;
  border-radius: var(--radius);
  overflow: hidden;
}

.video-player {
  width: 100%;
  max-height: 60vh;
  display: block;
}
</style>
