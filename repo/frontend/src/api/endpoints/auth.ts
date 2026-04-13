import apiClient from '@/api/client'

export interface LoginPayload {
  username: string
  password: string
  captcha_id?: string
  captcha_token?: string
}

export interface RegisterPayload {
  username: string
  email: string
  password: string
}

export const authApi = {
  login(data: LoginPayload) {
    return apiClient.post('/auth/login', data)
  },
  register(data: RegisterPayload) {
    return apiClient.post('/auth/register', data)
  },
  refresh(refresh_token: string) {
    return apiClient.post('/auth/refresh', { refresh_token })
  },
  logout(refresh_token: string) {
    return apiClient.post('/auth/logout', { refresh_token })
  },
}
