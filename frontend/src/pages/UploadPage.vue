<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { videoApi } from '../api/video'

const router = useRouter()

const title = ref('')
const description = ref('')
const videoFile = ref<File | null>(null)
const coverFile = ref<File | null>(null)
const videoPreview = ref('')
const coverPreview = ref('')
const loading = ref(false)
const error = ref('')

interface UploadStage {
  uploading: boolean
  videoDone: boolean
  coverDone: boolean
}

const stage = reactive<UploadStage>({
  uploading: false,
  videoDone: false,
  coverDone: false,
})

function handleVideoFile(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  if (!file) return
  if (file.size > 200 * 1024 * 1024) {
    error.value = '视频文件不能超过200MB'
    return
  }
  if (!file.name.toLowerCase().endsWith('.mp4')) {
    error.value = '仅支持 .mp4 格式'
    return
  }
  videoFile.value = file
  videoPreview.value = URL.createObjectURL(file)
  error.value = ''
}

function handleCoverFile(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  if (!file) return
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (!ext || !['jpg', 'jpeg', 'png', 'webp'].includes(ext)) {
    error.value = '仅支持 jpg / png / webp 格式封面'
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    error.value = '封面文件不能超过10MB'
    return
  }
  coverFile.value = file
  coverPreview.value = URL.createObjectURL(file)
  error.value = ''
}

async function handleSubmit() {
  error.value = ''

  if (!title.value.trim()) {
    error.value = '请输入视频标题'
    return
  }
  if (!videoFile.value) {
    error.value = '请选择视频文件'
    return
  }

  loading.value = true
  stage.uploading = true
  stage.videoDone = false
  stage.coverDone = false

  try {
    const videoRes = await videoApi.uploadVideo(videoFile.value)
    const playUrl = videoRes.data.url
    stage.videoDone = true

    let coverUrl = videoRes.data.cover_url || ''
    if (coverFile.value) {
      const coverRes = await videoApi.uploadCover(coverFile.value)
      coverUrl = coverRes.data.url
    }
    stage.coverDone = true

    const publishRes = await videoApi.publish({
      title: title.value.trim(),
      description: description.value.trim(),
      play_url: playUrl,
      cover_url: coverUrl,
    })

    router.push(`/video/${publishRes.data.id}`)
  } catch (e: any) {
    error.value = e.response?.data?.error || e.response?.data?.message || '发布失败'
  } finally {
    loading.value = false
    stage.uploading = false
  }
}
</script>

<template>
  <div class="container upload-page">
    <div class="upload-card">
      <h1 class="upload-title">发布视频</h1>

      <div class="form-group">
        <label class="form-label">标题</label>
        <input
          v-model="title"
          class="form-input"
          type="text"
          placeholder="输入视频标题"
          maxlength="100"
        />
      </div>

      <div class="form-group">
        <label class="form-label">描述</label>
        <textarea
          v-model="description"
          class="form-input"
          placeholder="输入视频描述（可选）"
          rows="3"
          maxlength="500"
        ></textarea>
      </div>

      <div class="form-group">
        <label class="form-label">视频文件 (.mp4, 最大200MB)</label>
        <input
          type="file"
          accept="video/mp4"
          class="file-input"
          @change="handleVideoFile"
        />
        <video
          v-if="videoPreview"
          :src="videoPreview"
          controls
          class="preview-video"
        ></video>
      </div>

      <div class="form-group">
        <label class="form-label">封面图片 (可选, jpg/png/webp)</label>
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp"
          class="file-input"
          @change="handleCoverFile"
        />
        <img
          v-if="coverPreview"
          :src="coverPreview"
          class="preview-cover"
          alt="封面预览"
        />
      </div>

      <!-- 上传进度 -->
      <div v-if="stage.uploading" class="progress-box">
        <p>
          <span v-if="!stage.videoDone"><span class="spinner"></span> 正在上传视频...</span>
          <span v-else-if="!stage.coverDone && coverFile"><span class="spinner"></span> 正在上传封面...</span>
          <span v-else>正在发布...</span>
        </p>
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>

      <button
        class="btn btn-primary upload-btn"
        :disabled="loading"
        @click="handleSubmit"
      >
        {{ loading ? '发布中...' : '发布' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.upload-page {
  display: flex;
  justify-content: center;
  padding-top: 24px;
  padding-bottom: 40px;
}

.upload-card {
  width: 100%;
  max-width: 560px;
  background: var(--color-surface);
  padding: 28px;
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.upload-title {
  font-size: 22px;
  margin-bottom: 20px;
}

.file-input {
  display: block;
  font-size: 14px;
}

.preview-video {
  width: 100%;
  max-height: 300px;
  margin-top: 10px;
  border-radius: var(--radius);
  background: #000;
}

.preview-cover {
  width: 100%;
  max-height: 260px;
  margin-top: 10px;
  border-radius: var(--radius);
  object-fit: cover;
}

.progress-box {
  padding: 12px;
  background: #f0f0f0;
  border-radius: var(--radius);
  font-size: 14px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.upload-btn {
  width: 100%;
  padding: 12px;
  margin-top: 8px;
}
</style>
