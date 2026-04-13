import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
}

export const useUIStore = defineStore('ui', () => {
  const toasts = ref<Toast[]>([])
  const isGlobalLoading = ref(false)
  let nextId = 0

  function showToast(message: string, type: Toast['type'] = 'info') {
    const id = nextId++
    toasts.value.push({ id, message, type })
    setTimeout(() => dismissToast(id), 5000)
  }

  function dismissToast(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function setLoading(loading: boolean) {
    isGlobalLoading.value = loading
  }

  return { toasts, isGlobalLoading, showToast, dismissToast, setLoading }
})
