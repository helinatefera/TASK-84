<template>
  <div class="write-review">
    <h1>Write a Review</h1>
    <form @submit.prevent="handleSubmit">
      <div class="rating-input">
        <label>Rating</label>
        <div class="stars">
          <button v-for="star in 5" :key="star" type="button" :class="{ filled: star <= rating }" @click="rating = star">
            {{ star <= rating ? '\u2605' : '\u2606' }}
          </button>
        </div>
      </div>

      <div class="form-group">
        <label>Review (optional, max 2000 chars)</label>
        <textarea v-model="body" maxlength="2000" rows="6" class="form-input"></textarea>
        <small>{{ body.length }}/2000</small>
      </div>

      <div class="form-group">
        <label>Images (up to 6, JPEG/PNG/WebP, max 5MB each)</label>
        <input type="file" multiple accept="image/jpeg,image/png,image/webp" @change="handleFiles" />
        <div class="image-previews">
          <div v-for="(img, i) in imagePreviews" :key="i" class="preview-item">
            <img :src="img" />
            <button type="button" @click="removeImage(i)">x</button>
          </div>
        </div>
      </div>

      <p v-if="draftSaved" class="draft-indicator">Draft saved {{ draftSavedAgo }}</p>
      <p v-if="error" class="error-text">{{ error }}</p>
      <p v-if="uploadProgress" class="upload-indicator">{{ uploadProgress }}</p>

      <button type="submit" :disabled="isLocked || rating === 0" class="submit-btn">
        {{ isLocked ? 'Submitting...' : 'Submit Review' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { reviewsApi } from '@/api/endpoints/reviews'
import apiClient from '@/api/client'

const props = defineProps<{ itemId: string }>()
const router = useRouter()

const rating = ref(0)
const body = ref('')
const images = ref<File[]>([])
const imagePreviews = ref<string[]>([])
const error = ref('')
const isLocked = ref(false)
const draftSaved = ref(false)
const draftSavedAgo = ref('')
const uploadProgress = ref('')
const idempotencyKey = ref(crypto.randomUUID())

// Draft autosave
const draftKey = `review-draft-${props.itemId}`
let autosaveInterval: ReturnType<typeof setInterval>

function saveDraft() {
  const draft = { rating: rating.value, body: body.value, savedAt: new Date().toISOString() }
  localStorage.setItem(draftKey, JSON.stringify(draft))
  draftSaved.value = true
  draftSavedAgo.value = 'just now'
}

function loadDraft() {
  const saved = localStorage.getItem(draftKey)
  if (saved) {
    const draft = JSON.parse(saved)
    rating.value = draft.rating || 0
    body.value = draft.body || ''
    draftSaved.value = true
  }
}

onMounted(() => {
  loadDraft()
  autosaveInterval = setInterval(saveDraft, 10000)
})

onUnmounted(() => {
  clearInterval(autosaveInterval)
})

function handleFiles(e: Event) {
  const target = e.target as HTMLInputElement
  if (!target.files) return
  const newFiles = Array.from(target.files)
  const remaining = 6 - images.value.length
  const toAdd = newFiles.slice(0, remaining)

  for (const file of toAdd) {
    if (file.size > 5 * 1024 * 1024) {
      error.value = `${file.name} exceeds 5MB limit`
      continue
    }
    images.value.push(file)
    imagePreviews.value.push(URL.createObjectURL(file))
  }
}

function removeImage(index: number) {
  URL.revokeObjectURL(imagePreviews.value[index])
  images.value.splice(index, 1)
  imagePreviews.value.splice(index, 1)
}

async function uploadImages(): Promise<number[]> {
  const imageIds: number[] = []
  if (images.value.length === 0) return imageIds
  for (let i = 0; i < images.value.length; i++) {
    uploadProgress.value = `Uploading image ${i + 1} of ${images.value.length}...`
    const formData = new FormData()
    formData.append('file', images.value[i])
    try {
      const resp = await apiClient.post('/images/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      if (resp.data?.image_id) {
        imageIds.push(resp.data.image_id)
      }
    } catch (e: any) {
      error.value = `Failed to upload image ${i + 1}: ${e.response?.data?.msg || e.response?.data?.message || 'Upload error'}`
    }
  }
  uploadProgress.value = ''
  return imageIds
}

async function handleSubmit() {
  if (isLocked.value || rating.value === 0) return
  isLocked.value = true
  error.value = ''
  uploadProgress.value = ''

  try {
    // Upload images first to get their IDs
    const imageIds = await uploadImages()
    if (error.value) {
      // Image upload had an error but we can still proceed with any successful ones
    }

    // Create review with linked image IDs
    await reviewsApi.create(
      props.itemId,
      { rating: rating.value, body: body.value || undefined, image_ids: imageIds.length > 0 ? imageIds : undefined },
      idempotencyKey.value
    )

    localStorage.removeItem(draftKey)
    router.push({ name: 'itemDetail', params: { id: props.itemId } })
  } catch (e: any) {
    error.value = e.response?.data?.msg || e.response?.data?.message || 'Failed to submit review'
    // Keep same idempotency key for retry
  }

  setTimeout(() => { isLocked.value = false }, 3000)
}
</script>

<style scoped>
.write-review { max-width: 600px; }
.rating-input { margin-bottom: 1rem; }
.stars { display: flex; gap: 0.25rem; }
.stars button { background: none; border: none; font-size: 2rem; cursor: pointer; color: #ccc; padding: 0; }
.stars button.filled { color: #f5a623; }
.form-group { display: flex; flex-direction: column; gap: 0.3rem; margin-bottom: 1rem; }
.form-group label { font-weight: 600; }
.form-group small { color: #888; font-size: 0.8rem; }
.form-input { padding: 0.6rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; resize: vertical; }
.image-previews { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-top: 0.5rem; }
.preview-item { position: relative; width: 80px; height: 80px; }
.preview-item img { width: 100%; height: 100%; object-fit: cover; border-radius: 4px; }
.preview-item button { position: absolute; top: -4px; right: -4px; background: #e53e3e; color: white; border: none; border-radius: 50%; width: 20px; height: 20px; font-size: 0.7rem; cursor: pointer; }
.draft-indicator { color: #888; font-size: 0.85rem; }
.upload-indicator { color: #646cff; font-size: 0.85rem; }
.error-text { color: #e53e3e; }
.submit-btn { padding: 0.7rem 1.5rem; background: #646cff; color: white; border: none; border-radius: 4px; font-size: 1rem; cursor: pointer; }
.submit-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
