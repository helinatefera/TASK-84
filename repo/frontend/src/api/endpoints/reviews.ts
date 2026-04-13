import apiClient from '@/api/client'

export const reviewsApi = {
  listByItem(itemId: string, params: { page?: number; per_page?: number }) {
    return apiClient.get(`/items/${itemId}/reviews`, { params })
  },
  create(itemId: string, data: { rating: number; body?: string; image_ids?: number[] }, idempotencyKey: string) {
    return apiClient.post(`/items/${itemId}/reviews`, data, {
      headers: { 'X-Idempotency-Key': idempotencyKey },
    })
  },
  update(id: string, data: { rating?: number; body?: string }) {
    return apiClient.put(`/reviews/${id}`, data)
  },
  delete(id: string) {
    return apiClient.delete(`/reviews/${id}`)
  },
  uploadImages(reviewId: string, files: File[]) {
    const uploads = files.map((f) => {
      const formData = new FormData()
      formData.append('file', f)
      return apiClient.post(`/images/upload`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    })
    return Promise.all(uploads)
  },
}
