export interface User {
  id: string
  username: string
  email: string
  role: string
  is_active: boolean
  created_at: string
}

export interface UserPreferences {
  locale: string
  timezone: string
  notification_settings: Record<string, boolean>
}
