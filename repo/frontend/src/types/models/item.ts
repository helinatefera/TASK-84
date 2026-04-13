export interface Item {
  id: string
  title: string
  description: string | null
  category: string | null
  lifecycle_state: string
  created_at: string
  published_at: string | null
}

export interface RatingAggregate {
  avg_rating: number
  rating_count: number
  rating_1: number
  rating_2: number
  rating_3: number
  rating_4: number
  rating_5: number
}
