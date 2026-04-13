export interface Review {
  id: string
  rating: number
  body: string | null
  fraud_status: string
  created_at: string
  updated_at: string
  author?: { id: string; username: string }
  images?: ReviewImage[]
}

export interface ReviewImage {
  hash: string
  sort_order: number
  mime_type: string
}
