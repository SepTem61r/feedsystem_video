<script setup lang="ts">
defineProps<{ src: string; poster?: string }>()
</script>

<template>
  <div class="player-wrapper">
    <video
      class="video-player"
      :src="src"
      :poster="poster"
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
