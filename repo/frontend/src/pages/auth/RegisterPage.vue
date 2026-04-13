<template>
  <form @submit.prevent="handleRegister" class="register-form">
    <div class="form-group">
      <label>Username</label>
      <input v-model="form.username" type="text" required minlength="3" maxlength="32" class="form-input" />
      <small>3-32 characters, letters and numbers only</small>
    </div>
    <div class="form-group">
      <label>Email</label>
      <input v-model="form.email" type="email" required class="form-input" />
    </div>
    <div class="form-group">
      <label>Password</label>
      <input v-model="form.password" type="password" required minlength="8" class="form-input" />
      <small>Min 8 characters</small>
    </div>
    <div class="form-group">
      <label>Confirm Password</label>
      <input v-model="confirmPassword" type="password" required class="form-input" />
    </div>
    <p v-if="error" class="error-text">{{ error }}</p>
    <button type="submit" :disabled="isSubmitting" class="submit-btn">
      {{ isSubmitting ? 'Creating account...' : 'Register' }}
    </button>
    <p class="login-link">
      Have an account? <router-link :to="{ name: 'login' }">Sign in</router-link>
    </p>
  </form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()
const router = useRouter()
const form = reactive({ username: '', email: '', password: '' })
const confirmPassword = ref('')
const error = ref('')
const isSubmitting = ref(false)

async function handleRegister() {
  error.value = ''
  if (form.password !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  isSubmitting.value = true
  try {
    await authStore.register(form.username, form.email, form.password)
    router.push({ name: 'login' })
  } catch (e: any) {
    error.value = e.response?.data?.msg || e.response?.data?.message || 'Registration failed'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.register-form { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.3rem; }
.form-group label { font-weight: 600; font-size: 0.9rem; }
.form-group small { color: #888; font-size: 0.8rem; }
.form-input { padding: 0.6rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
.submit-btn { padding: 0.7rem; background: #646cff; color: white; border: none; border-radius: 4px; font-size: 1rem; cursor: pointer; }
.submit-btn:disabled { opacity: 0.6; }
.error-text { color: #e53e3e; font-size: 0.9rem; margin: 0; }
.login-link { text-align: center; font-size: 0.9rem; }
</style>
