import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

// --- Mocks ---
const mockGet = vi.fn()
const mockPost = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

const mockLogin = vi.fn()
const mockIncrementLoginFailures = vi.fn()
const mockLoginFailureCount = { value: 0 }
vi.mock('@/stores/auth.store', () => ({
  useAuthStore: () => ({
    login: mockLogin,
    incrementLoginFailures: mockIncrementLoginFailures,
    get loginFailureCount() { return mockLoginFailureCount.value },
  }),
}))

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import LoginPage from '@/pages/auth/LoginPage.vue'

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    mockLoginFailureCount.value = 0
  })

  it('renders username and password inputs and a submit button', () => {
    const wrapper = mount(LoginPage)
    expect(wrapper.find('input[type="text"]').exists()).toBe(true)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toContain('Sign In')
  })

  it('does not show CAPTCHA when loginFailureCount < 5', () => {
    const wrapper = mount(LoginPage)
    expect(wrapper.find('.captcha-img').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('CAPTCHA')
  })

  it('shows CAPTCHA field once login failures reach 5', async () => {
    mockLoginFailureCount.value = 5
    mockGet.mockResolvedValue({ data: { captcha_id: 'abc', captcha_image: 'data:image/png;base64,xxx' } })
    const wrapper = mount(LoginPage)
    await flushPromises()
    expect(wrapper.text()).toContain('CAPTCHA')
    expect(mockGet).toHaveBeenCalledWith('/captcha/generate')
  })

  it('calls authStore.login and redirects to catalog on success', async () => {
    mockLogin.mockResolvedValue(undefined)
    const wrapper = mount(LoginPage)
    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('Secret1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockLogin).toHaveBeenCalledWith('alice', 'Secret1', undefined, undefined)
    expect(mockPush).toHaveBeenCalledWith({ name: 'catalog' })
  })

  it('increments failure counter and shows error message when login fails', async () => {
    mockLogin.mockRejectedValue({ response: { data: { msg: 'Invalid credentials' } } })
    const wrapper = mount(LoginPage)
    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('wrong')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockIncrementLoginFailures).toHaveBeenCalled()
    expect(wrapper.find('.error-text').text()).toBe('Invalid credentials')
    expect(mockPush).not.toHaveBeenCalled()
  })

  it('passes captcha_id and captcha_answer to login when CAPTCHA is active', async () => {
    mockLoginFailureCount.value = 5
    mockGet.mockResolvedValue({ data: { captcha_id: 'cap-123', captcha_image: 'img' } })
    mockLogin.mockResolvedValue(undefined)
    const wrapper = mount(LoginPage)
    await flushPromises()

    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('Secret1')
    const inputs = wrapper.findAll('input')
    // The CAPTCHA answer input comes after username + password.
    await inputs[inputs.length - 1].setValue('42')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockLogin).toHaveBeenCalledWith('alice', 'Secret1', 'cap-123', '42')
  })

  it('falls back to a generic error message when the server provides none', async () => {
    mockLogin.mockRejectedValue({})
    const wrapper = mount(LoginPage)
    await wrapper.find('input[type="text"]').setValue('a')
    await wrapper.find('input[type="password"]').setValue('b')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('.error-text').text()).toBe('Login failed')
  })
})
