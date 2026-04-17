import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'

// These pages are currently "Coming soon" stubs; these smoke tests lock in
// their rendered contract so the audit records real component coverage and
// any future regression that removes the heading is caught.
import SensitiveWordsPage from '@/pages/moderation/SensitiveWordsPage.vue'
import UserManagementPage from '@/pages/admin/UserManagementPage.vue'
import IpManagementPage from '@/pages/admin/IpManagementPage.vue'
import SystemMonitorPage from '@/pages/admin/SystemMonitorPage.vue'

describe('SensitiveWordsPage', () => {
  it('renders the Sensitive Words heading', () => {
    const wrapper = mount(SensitiveWordsPage)
    expect(wrapper.find('h1').text()).toBe('Sensitive Words')
  })
  it('indicates that the feature is coming soon', () => {
    expect(mount(SensitiveWordsPage).text()).toContain('Coming soon')
  })
})

describe('UserManagementPage', () => {
  it('renders the User Management heading', () => {
    const wrapper = mount(UserManagementPage)
    expect(wrapper.find('h1').text()).toBe('User Management')
  })
  it('indicates that the feature is coming soon', () => {
    expect(mount(UserManagementPage).text()).toContain('Coming soon')
  })
})

describe('IpManagementPage', () => {
  it('renders the IP Rules heading', () => {
    const wrapper = mount(IpManagementPage)
    expect(wrapper.find('h1').text()).toBe('IP Rules')
  })
  it('indicates that the feature is coming soon', () => {
    expect(mount(IpManagementPage).text()).toContain('Coming soon')
  })
})

describe('SystemMonitorPage', () => {
  it('renders the System Monitor heading', () => {
    const wrapper = mount(SystemMonitorPage)
    expect(wrapper.find('h1').text()).toBe('System Monitor')
  })
  it('indicates that the feature is coming soon', () => {
    expect(mount(SystemMonitorPage).text()).toContain('Coming soon')
  })
})
