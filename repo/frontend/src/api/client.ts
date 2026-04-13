import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

let csrfBootstrapped = false

// Fetch a CSRF token from the server. The middleware sets the csrf_token
// cookie on the response, which subsequent requests read.
async function ensureCsrfToken(): Promise<void> {
  if (csrfBootstrapped) return
  try {
    await apiClient.get('/csrf')
    csrfBootstrapped = true
  } catch {
    // CSRF endpoint may not exist if disabled — proceed without it
    csrfBootstrapped = true
  }
}

apiClient.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  // Bootstrap CSRF token before first mutating request
  if (['post', 'put', 'patch', 'delete'].includes(config.method ?? '')) {
    await ensureCsrfToken()
    const csrfToken = getCookie('csrf_token')
    if (csrfToken) {
      config.headers['X-CSRF-Token'] = csrfToken
    }
  }

  // Attach idempotency key on POST requests to prevent duplicate submissions
  if (config.method === 'post' && !config.headers['X-Idempotency-Key']) {
    config.headers['X-Idempotency-Key'] = crypto.randomUUID()
  }

  // Attach JWT Bearer token
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }

  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config

    // On CSRF rejection, re-fetch token and retry once
    if (error.response?.status === 403 && error.response?.data?.code === 'CSRF_INVALID' && !original._csrfRetried) {
      original._csrfRetried = true
      csrfBootstrapped = false
      await ensureCsrfToken()
      const csrfToken = getCookie('csrf_token')
      if (csrfToken) {
        original.headers['X-CSRF-Token'] = csrfToken
      }
      return apiClient(original)
    }

    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)

function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : null
}

export default apiClient
