<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { videoApi } from '../api/video'
import { likeApi } from '../api/like'
import { socialApi } from '../api/social'
import { accountApi } from '../api/account'
import VideoGrid from '../components/VideoGrid.vue'
import type { Video, AccountInfo, FollowerInfo } from '../types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

type ProfileTab = 'works' | 'likes' | 'following' | 'followers'

const profileUser = ref<AccountInfo | null>(null)
const videos = ref<Video[]>([])
const likedVideos = ref<Video[]>([])
const vloggers = ref<FollowerInfo[]>([])
const followers = ref<FollowerInfo[]>([])
const loading = ref(true)
const error = ref('')
const isFollowing = ref(false)
const followLoading = ref(false)
const activeTab = ref<ProfileTab>('works')
const tabLoading = ref(false)
const showRename = ref(false)
const renameText = ref('')
const renameLoading = ref(false)
const renameError = ref('')

const showChangePwd = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const changePwdLoading = ref(false)
const changePwdError = ref('')

function profileId(): number {
  return Number(route.params.id)
}

const isSelf = computed(() => auth.currentUser?.id === profileId())

const profileTabs = computed<{ key: ProfileTab; label: string; show: boolean }[]>(() => [
  { key: 'works', label: '作品', show: true },
  { key: 'likes', label: '点赞', show: isSelf.value && auth.isLoggedIn },
  { key: 'following', label: '关注', show: auth.isLoggedIn },
  { key: 'followers', label: '粉丝', show: auth.isLoggedIn },
])

async function fetchProfile() {
  loading.value = true
  try {
    const [userRes, videoRes] = await Promise.all([
      accountApi.findByID({ id: profileId() }),
      videoApi.listByAuthorID({ author_id: profileId() }),
    ])
    profileUser.value = userRes.data
    videos.value = videoRes.data
    if (auth.isLoggedIn && !isSelf.value) {
      checkFollowStatus()
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载失败'
  } finally {
    loading.value = false
  }
}

async function checkFollowStatus() {
  try {
    const res = await socialApi.getAllVloggers({ follower_id: auth.currentUser!.id })
    isFollowing.value = res.data.vloggers?.some(v => v.id === profileId()) ?? false
  } catch { /* ignore */ }
}

async function toggleFollow() {
  followLoading.value = true
  try {
    if (isFollowing.value) {
      await socialApi.unfollow({ vlogger_id: profileId() })
      isFollowing.value = false
    } else {
      await socialApi.follow({ vlogger_id: profileId() })
      isFollowing.value = true
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '操作失败'
  } finally {
    followLoading.value = false
  }
}

async function switchTab(key: ProfileTab) {
  activeTab.value = key
  tabLoading.value = true
  try {
    switch (key) {
      case 'likes':
        if (likedVideos.value.length === 0) {
          const res = await likeApi.listMyLikedVideos()
          likedVideos.value = res.data.videos ?? []
        }
        break
      case 'following':
        if (vloggers.value.length === 0) {
          const res = await socialApi.getAllVloggers({ follower_id: profileId() })
          vloggers.value = res.data.vloggers ?? []
        }
        break
      case 'followers':
        if (followers.value.length === 0) {
          const res = await socialApi.getAllFollowers({ vlogger_id: profileId() })
          followers.value = res.data.followers ?? []
        }
        break
    }
  } catch { /* ignore */ }
  finally { tabLoading.value = false }
}

async function handleRename() {
  if (!renameText.value.trim()) return
  renameLoading.value = true
  renameError.value = ''
  try {
    await auth.rename(renameText.value.trim())
    profileUser.value!.username = renameText.value.trim()
    showRename.value = false
    renameText.value = ''
  } catch (e: any) {
    renameError.value = e.response?.data?.error || '修改失败'
  } finally {
    renameLoading.value = false
  }
}

async function handleChangePassword() {
  changePwdError.value = ''
  if (!oldPassword.value) { changePwdError.value = '请输入旧密码'; return }
  if (newPassword.value.length < 6) { changePwdError.value = '新密码至少6个字符'; return }
  if (newPassword.value !== confirmPassword.value) { changePwdError.value = '两次密码不一致'; return }
  changePwdLoading.value = true
  try {
    await accountApi.changePassword({
      username: auth.currentUser!.username,
      old_password: oldPassword.value,
      new_password: newPassword.value,
    })
    showChangePwd.value = false
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    alert('密码修改成功')
  } catch (e: any) {
    changePwdError.value = e.response?.data?.error || '修改密码失败'
  } finally {
    changePwdLoading.value = false
  }
}

function goDetail(id: number) {
  router.push(`/video/${id}`)
}
function goUser(id: number) {
  router.push(`/profile/${id}`)
}

onMounted(fetchProfile)
watch(() => route.params.id, () => {
  if (route.name === 'Profile') {
    activeTab.value = 'works'
    likedVideos.value = []
    vloggers.value = []
    followers.value = []
    fetchProfile()
  }
})
</script>

<template>
  <div class="container profile-page">
    <div v-if="loading" class="profile-loading">
      <span class="spinner"></span>
    </div>

    <div v-else-if="error" class="error-text profile-error">{{ error }}</div>

    <template v-else-if="profileUser">
      <div class="profile-header">
        <div class="profile-avatar">
          {{ profileUser.username.charAt(0).toUpperCase() }}
        </div>
        <div class="profile-info">
          <h1 class="profile-username">@{{ profileUser.username }}</h1>
          <p class="profile-id">ID: {{ profileUser.id }}</p>
        </div>
        <div class="profile-actions">
          <button
            v-if="!isSelf && auth.isLoggedIn"
            class="btn"
            :class="isFollowing ? 'btn-outline' : 'btn-primary'"
            :disabled="followLoading"
            @click="toggleFollow"
          >
            {{ isFollowing ? '已关注' : '+ 关注' }}
          </button>
          <button
            v-if="isSelf"
            class="btn btn-outline btn-sm"
            @click="showRename = !showRename; showChangePwd = false"
          >
            修改用户名
          </button>
          <button
            v-if="isSelf"
            class="btn btn-outline btn-sm"
            @click="showChangePwd = !showChangePwd; showRename = false"
          >
            修改密码
          </button>
        </div>
      </div>

      <div v-if="showRename" class="rename-form">
        <input
          v-model="renameText"
          class="form-input"
          placeholder="新的用户名"
        />
        <button
          class="btn btn-primary btn-sm"
          :disabled="renameLoading || !renameText.trim()"
          @click="handleRename"
        >
          {{ renameLoading ? '保存中...' : '保存' }}
        </button>
        <button class="btn btn-outline btn-sm" @click="showRename = false">取消</button>
        <p v-if="renameError" class="error-text">{{ renameError }}</p>
      </div>

      <div v-if="showChangePwd" class="password-form">
        <div class="form-group">
          <label class="form-label">旧密码</label>
          <input v-model="oldPassword" class="form-input" type="password" placeholder="输入旧密码" />
        </div>
        <div class="form-group">
          <label class="form-label">新密码</label>
          <input v-model="newPassword" class="form-input" type="password" placeholder="至少6个字符" />
        </div>
        <div class="form-group">
          <label class="form-label">确认新密码</label>
          <input v-model="confirmPassword" class="form-input" type="password" placeholder="再次输入新密码" />
        </div>
        <div class="form-actions">
          <button class="btn btn-primary btn-sm" :disabled="changePwdLoading" @click="handleChangePassword">
            {{ changePwdLoading ? '修改中...' : '确认修改' }}
          </button>
          <button class="btn btn-outline btn-sm" @click="showChangePwd = false">取消</button>
        </div>
        <p v-if="changePwdError" class="error-text">{{ changePwdError }}</p>
      </div>

      <!-- tabs -->
      <div class="profile-tabs">
        <button
          v-for="t in profileTabs.filter(t => t.show)"
          :key="t.key"
          class="profile-tab"
          :class="{ active: activeTab === t.key }"
          @click="switchTab(t.key)"
        >
          {{ t.label }}
          <span v-if="t.key === 'works'">({{ videos.length }})</span>
          <span v-if="t.key === 'likes'">({{ likedVideos.length }})</span>
          <span v-if="t.key === 'following'">({{ vloggers.length }})</span>
          <span v-if="t.key === 'followers'">({{ followers.length }})</span>
        </button>
      </div>

      <!-- tab content -->
      <div v-if="tabLoading" class="tab-loading"><span class="spinner"></span></div>

      <template v-else>
        <!-- 作品 -->
        <VideoGrid v-if="activeTab === 'works'" :videos="videos" @click="goDetail" />

        <!-- 点赞 -->
        <VideoGrid v-if="activeTab === 'likes'" :videos="likedVideos" @click="goDetail" />

        <!-- 关注 / 粉丝 -->
        <div v-if="activeTab === 'following'" class="user-list">
          <div v-for="u in vloggers" :key="u.id" class="user-item" @click="goUser(u.id)">
            <div class="user-avatar-sm">{{ u.username.charAt(0).toUpperCase() }}</div>
            <span class="user-name">@{{ u.username }}</span>
          </div>
          <p v-if="vloggers.length === 0" class="list-empty">还没有关注任何人</p>
        </div>

        <div v-if="activeTab === 'followers'" class="user-list">
          <div v-for="u in followers" :key="u.id" class="user-item" @click="goUser(u.id)">
            <div class="user-avatar-sm">{{ u.username.charAt(0).toUpperCase() }}</div>
            <span class="user-name">@{{ u.username }}</span>
          </div>
          <p v-if="followers.length === 0" class="list-empty">还没有粉丝</p>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped>
.profile-page {
  padding-top: 24px;
  padding-bottom: 32px;
}

.profile-loading {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.profile-error {
  text-align: center;
  padding: 40px 0;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  margin-bottom: 24px;
}

.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
}

.profile-username {
  font-size: 20px;
}

.profile-id {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.profile-actions {
  flex-shrink: 0;
}

.rename-form {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.rename-form .form-input {
  max-width: 200px;
}

.password-form {
  margin-bottom: 16px;
  padding: 16px;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.password-form .form-input {
  max-width: 280px;
}
.form-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.section-heading {
  font-size: 16px;
  margin-bottom: 14px;
}

/* tabs */
.profile-tabs {
  display: flex;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 16px;
}
.profile-tab {
  flex: 1;
  background: none;
  border: none;
  padding: 10px 0;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
}
.profile-tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-loading {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}

/* user list */
.user-list {
  display: flex;
  flex-direction: column;
}
.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius);
  cursor: pointer;
  transition: background 0.15s;
}
.user-item:hover {
  background: var(--color-bg);
}
.user-avatar-sm {
  width: 40px; height: 40px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 16px; font-weight: 700;
  flex-shrink: 0;
}
.user-name {
  font-size: 14px;
  font-weight: 500;
}
.list-empty {
  text-align: center;
  color: var(--color-text-secondary);
  padding: 32px 0;
  font-size: 14px;
}
</style>
