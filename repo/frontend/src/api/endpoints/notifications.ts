import apiClient from '@/api/client'

export const notificationsApi = {
  list(params: { page?: number; per_page?: number; unread_only?: boolean }) {
    return apiClient.get('/notifications', { params })
  },
  getById(id: string) {
    return apiClient.get(`/notifications/${id}`)
  },
  unreadCount() {
    return apiClient.get('/notifications/unread-count')
  },
  markRead(id: string) {
    return apiClient.put(`/notifications/${id}/read`)
  },
  markAllRead() {
    return apiClient.put('/notifications/read-all')
  },
}
