import apiClient from '@/api/client'

export const itemsApi = {
  list(params: { page?: number; per_page?: number; search?: string; category?: string }) {
    return apiClient.get('/items', { params })
  },
  getById(id: string) {
    return apiClient.get(`/items/${id}`)
  },
}
