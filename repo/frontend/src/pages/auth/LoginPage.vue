<template>
  <form @submit.prevent="handleLogin" class="login-form">
    <div class="form-group">
      <label>Username</label>
      <input v-model="username" type="text" required class="form-input" />
    </div>
    <div class="form-group">
      <label>Password</label>
      <input v-model="password" type="password" required class="form-input" />
    </div>
    <div v-if="showCaptcha" class="form-group">
      <label>CAPTCHA</label>
      <img v-if="captchaImage" :src="captchaImage" alt="CAPTCHA" class="captcha-img" />
      <input v-model="captchaAnswer" type="text" placeholder="Enter answer" class="form-input" />
      <button type="button" @click="loadCaptcha" class="captcha-refresh">Refresh</button>
    </div>
    <p v-if="error" class="error-text">{{ error }}</p>
    <button type="submit" :disabled="isSubmitting" class="submit-btn">
      {{ isSubmitting ? 'Signing in...' : 'Sign In' }}
    </button>
    <p class="register-link">
      No account? <router-link :to="{ name: 'register' }">Register</router-link>
    </p>
  </form>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import apiClient from '@/api/client'

const authStore = useAuthStore()
const router = useRouter()

const username = ref('')
const password = ref('')
const captchaAnswer = ref('')
const captchaId = ref('')
const captchaImage = ref('')
const error = ref('')
const isSubmitting = ref(false)

const showCaptcha = computed(() => authStore.loginFailureCount >= 5)

onMounted(() => {
  if (showCaptcha.value) loadCaptcha()
})

async function loadCaptcha() {
  try {
    const { data } = await apiClient.get('/captcha/generate')
    captchaId.value = data.captcha_id
    captchaImage.value = data.captcha_image
  } catch {
    error.value = 'Failed to load CAPTCHA'
  }
}

async function handleLogin() {
  error.value = ''
  isSubmitting.value = true
  try {
    await authStore.login(
      username.value,
      password.value,
      showCaptcha.value ? captchaId.value : undefined,
      showCaptcha.value ? captchaAnswer.value : undefined
    )
    router.push({ name: 'catalog' })
  } catch (e: any) {
    authStore.incrementLoginFailures()
    error.value = e.response?.data?.msg || e.response?.data?.message || 'Login failed'
    if (showCaptcha.value) loadCaptcha()
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.login-form { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.3rem; }
.form-group label { font-weight: 600; font-size: 0.9rem; }
.form-input { padding: 0.6rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
.submit-btn { padding: 0.7rem; background: #646cff; color: white; border: none; border-radius: 4px; font-size: 1rem; cursor: pointer; }
.submit-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.error-text { color: #e53e3e; font-size: 0.9rem; margin: 0; }
.register-link { text-align: center; font-size: 0.9rem; }
.captcha-img { max-width: 200px; border: 1px solid #ddd; border-radius: 4px; }
.captcha-refresh { background: none; border: none; color: #646cff; cursor: pointer; font-size: 0.85rem; }
</style>
