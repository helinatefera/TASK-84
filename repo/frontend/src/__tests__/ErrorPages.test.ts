import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ForbiddenPage from '@/pages/errors/ForbiddenPage.vue'
import NotFoundPage from '@/pages/errors/NotFoundPage.vue'

const stubs = { RouterLink: { template: '<a><slot /></a>' } }

describe('ForbiddenPage', () => {
  it('renders 403 heading and access-denied message', () => {
    const wrapper = mount(ForbiddenPage, { global: { stubs } })
    expect(wrapper.find('h1').text()).toBe('403')
    expect(wrapper.text()).toContain('Access denied')
  })

  it('includes a home link', () => {
    const wrapper = mount(ForbiddenPage, { global: { stubs } })
    expect(wrapper.text()).toContain('Go Home')
  })
})

describe('NotFoundPage', () => {
  it('renders 404 heading and not-found message', () => {
    const wrapper = mount(NotFoundPage, { global: { stubs } })
    expect(wrapper.find('h1').text()).toBe('404')
    expect(wrapper.text()).toContain('Page not found')
  })

  it('includes a home link', () => {
    const wrapper = mount(NotFoundPage, { global: { stubs } })
    expect(wrapper.text()).toContain('Go Home')
  })
})
