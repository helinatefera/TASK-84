import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const mockRegister = vi.fn()
vi.mock('@/stores/auth.store', () => ({
  useAuthStore: () => ({ register: mockRegister }),
}))

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import RegisterPage from '@/pages/auth/RegisterPage.vue'

async function fillForm(wrapper: any, overrides: Partial<{ username: string; email: string; password: string; confirm: string }> = {}) {
  const values = {
    username: 'alice',
    email: 'alice@example.com',
    password: 'SecurePass1',
    confirm: 'SecurePass1',
    ...overrides,
  }
  const inputs = wrapper.findAll('input')
  await inputs[0].setValue(values.username)
  await inputs[1].setValue(values.email)
  await inputs[2].setValue(values.password)
  await inputs[3].setValue(values.confirm)
}

describe('RegisterPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('renders 4 inputs (username, email, password, confirm) and a submit button', () => {
    const wrapper = mount(RegisterPage)
    expect(wrapper.findAll('input')).toHaveLength(4)
    expect(wrapper.find('button[type="submit"]').text()).toContain('Register')
  })

  it('shows client-side error when passwords do not match and does not call register', async () => {
    const wrapper = mount(RegisterPage)
    await fillForm(wrapper, { confirm: 'Different1' })
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('.error-text').text()).toBe('Passwords do not match')
    expect(mockRegister).not.toHaveBeenCalled()
  })

  it('submits form, calls register, and navigates to login on success', async () => {
    mockRegister.mockResolvedValue(undefined)
    const wrapper = mount(RegisterPage)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockRegister).toHaveBeenCalledWith('alice', 'alice@example.com', 'SecurePass1')
    expect(mockPush).toHaveBeenCalledWith({ name: 'login' })
  })

  it('shows server error message on failed registration', async () => {
    mockRegister.mockRejectedValue({ response: { data: { msg: 'Username already exists' } } })
    const wrapper = mount(RegisterPage)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('.error-text').text()).toBe('Username already exists')
    expect(mockPush).not.toHaveBeenCalled()
  })

  it('falls back to generic error when server response has no message', async () => {
    mockRegister.mockRejectedValue({})
    const wrapper = mount(RegisterPage)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('.error-text').text()).toBe('Registration failed')
  })
})
