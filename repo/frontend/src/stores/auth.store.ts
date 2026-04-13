import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api/endpoints/auth'
import type { User } from '@/types/models/user'
import { Role } from '@/types/enums'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refresh_token'))
  const loginFailureCount = ref(parseInt(sessionStorage.getItem('login_failures') ?? '0'))

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const userRole = computed(() => user.value?.role ?? '')

  async function login(username: string, password: string, captchaId?: string, captchaToken?: string) {
    const { data } = await authApi.login({ username, password, captcha_id: captchaId, captcha_token: captchaToken })
    token.value = data.access_token
    refreshToken.value = data.refresh_token
    user.value = data.user
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    loginFailureCount.value = 0
    sessionStorage.setItem('login_failures', '0')
  }

  async function register(username: string, email: string, password: string) {
    await authApi.register({ username, email, password })
  }

  async function logout() {
    if (refreshToken.value) {
      try { await authApi.logout(refreshToken.value) } catch { /* ignore */ }
    }
    user.value = null
    token.value = null
    refreshToken.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  function incrementLoginFailures() {
    loginFailureCount.value++
    sessionStorage.setItem('login_failures', loginFailureCount.value.toString())
  }

  function hasRole(...roles: string[]): boolean {
    return roles.includes(userRole.value)
  }

  return {
    user, token, refreshToken, loginFailureCount,
    isAuthenticated, userRole,
    login, register, logout, incrementLoginFailures, hasRole,
  }
})
