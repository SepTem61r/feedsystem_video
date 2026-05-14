<script setup lang="ts">
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { feedApi } from '../api/feed'
import VideoCard from '../components/VideoCard.vue'
import type { FeedVideoItem } from '../types'

const router = useRouter()
const auth = useAuthStore()

type Tab = 'latest' | 'popular' | 'following' | 'likes'

const tabs: { key: Tab; label: string }[] = [
  { key: 'latest', label: '最新' },
  { key: 'popular', label: '最热' },
  { key: 'following', label: '关注' },
  { key: 'likes', label: '点赞排行' },
]

const activeTab = ref<Tab>('latest')
const loading = ref(false)
const error = ref('')
const sentinel = ref<HTMLElement | null>(null)

interface TabState {
  videos: FeedVideoItem[]
  hasMore: boolean
  cursor: any
}

function initTabState(): TabState {
  return { videos: [], hasMore: true, cursor: null }
}

const states = reactive<Record<Tab, TabState>>({
  latest: initTabState(),
  popular: initTabState(),
  following: initTabState(),
  likes: initTabState(),
})

async function loadFeed() {
  const state = states[activeTab.value]
  if (loading.value || !state.hasMore) return
  if (activeTab.value === 'following' && !auth.isLoggedIn) {
    error.value = '请先登录查看关注列表'
    return
  }
  error.value = ''
  loading.value = true
  try {
    let resp
    switch (activeTab.value) {
      case 'latest': {
        const latestTime = state.cursor?.nextTime ?? 0
        resp = await feedApi.listLatest({ limit: 10, latest_time: latestTime })
        state.cursor = { nextTime: resp.data.next_time }
        break
      }
      case 'popular': {
        const asOf = state.cursor?.asOf ?? 0
        const offset = state.cursor?.offset ?? 0
        resp = await feedApi.listByPopularity({
          limit: 10,
          as_of: asOf,
          offset,
          latest_popularity: state.cursor?.latestPopularity ?? 0,
          latest_before: state.cursor?.latestBefore ?? undefined,
          latest_id_before: state.cursor?.latestIDBefore ?? null,
        })
        state.cursor = {
          asOf: resp.data.as_of,
          offset: resp.data.next_offset,
          latestPopularity: resp.data.next_latest_popularity,
          latestBefore: resp.data.next_latest_before,
          latestIDBefore: resp.data.next_latest_id_before,
        }
        break
      }
      case 'following': {
        const latestTime = state.cursor?.nextTime ?? 0
        resp = await feedApi.listByFollowing({ limit: 10, latest_time: latestTime })
        state.cursor = { nextTime: resp.data.next_time }
        break
      }
      case 'likes': {
        const req: any = { limit: 10 }
        if (state.cursor) {
          req.likes_count_before = state.cursor.likesCount
          req.id_before = state.cursor.id
        }
        resp = await feedApi.listLikesCount(req)
        if (resp.data.next_likes_count_before != null && resp.data.next_id_before != null) {
          state.cursor = { likesCount: resp.data.next_likes_count_before, id: resp.data.next_id_before }
        } else {
          state.cursor = null
        }
        break
      }
    }
    if (resp) {
      state.videos.push(...resp.data.video_list)
      state.hasMore = resp.data.has_more
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载失败'
  } finally {
    loading.value = false
  }
}

function switchTab(key: Tab) {
  activeTab.value = key
  error.value = ''
  if (states[key].videos.length === 0 && states[key].hasMore) {
    loadFeed()
  }
}

let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting) {
      loadFeed()
    }
  }, { threshold: 0.1 })
  if (sentinel.value) observer.observe(sentinel.value)
  if (states[activeTab.value].videos.length === 0) {
    loadFeed()
  }
})

onUnmounted(() => {
  observer?.disconnect()
})

watch(sentinel, (el) => {
  observer?.disconnect()
  if (el) observer?.observe(el)
})

function goVideo(id: number) {
  router.push(`/video/${id}`)
}
</script>

<template>
  <div class="container feed-page">
    <div class="tabs">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="tab-btn"
        :class="{ active: activeTab === t.key }"
        @click="switchTab(t.key)"
      >
        {{ t.label }}
      </button>
    </div>

    <p v-if="error" class="error-text feed-error">{{ error }}</p>

    <div class="video-grid">
      <VideoCard
        v-for="v in states[activeTab].videos"
        :key="v.id"
        :video="v"
        @click="goVideo"
      />
    </div>

    <div v-if="loading" class="feed-loading">
      <span class="spinner"></span>
    </div>

    <p v-if="!states[activeTab].hasMore && states[activeTab].videos.length > 0" class="feed-end">
      — 到底了 —
    </p>

    <p v-if="!states[activeTab].hasMore && states[activeTab].videos.length === 0 && !loading" class="feed-empty">
      暂无视频
    </p>

    <div ref="sentinel" class="sentinel"></div>
  </div>
</template>

<style scoped>
.feed-page {
  padding-top: 16px;
  padding-bottom: 32px;
}

.tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 16px;
}

.tab-btn {
  flex: 1;
  padding: 10px 0;
  background: none;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-secondary);
  border-bottom: 2px solid transparent;
  transition: color 0.2s, border-color 0.2s;
}
.tab-btn.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.feed-error {
  text-align: center;
  margin-bottom: 12px;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.feed-loading {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

.feed-end,
.feed-empty {
  text-align: center;
  color: var(--color-text-secondary);
  font-size: 13px;
  padding: 24px 0;
}

.sentinel {
  height: 1px;
}
</style>
