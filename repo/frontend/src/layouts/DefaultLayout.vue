<template>
  <div class="default-layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2>Local Insights</h2>
      </div>
      <nav class="sidebar-nav">
        <router-link v-for="item in navItems" :key="item.label" :to="item.to" class="nav-item" active-class="active">
          {{ item.label }}
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <span class="user-info">{{ authStore.user?.username }}</span>
        <button class="logout-btn" @click="handleLogout">Logout</button>
      </div>
    </aside>
    <main class="main-content">
      <header class="top-bar">
        <div class="top-bar-right">
          <router-link to="/notifications" class="notification-bell">
            Notifications
            <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
          </router-link>
        </div>
      </header>
      <div class="page-content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { Role } from '@/types/enums'

const authStore = useAuthStore()
const router = useRouter()
const unreadCount = ref(0)

const navItems = computed(() => {
  const role = authStore.userRole
  const items = [
    { label: 'Catalog', to: { name: 'catalog' }, roles: [Role.User, Role.Moderator, Role.Analyst, Role.Admin] },
    { label: 'Favorites', to: { name: 'favorites' }, roles: [Role.User, Role.Moderator, Role.Analyst, Role.Admin] },
    { label: 'Moderation', to: { name: 'moderationQueue' }, roles: [Role.Moderator, Role.Admin] },
    { label: 'Analytics', to: { name: 'analytics' }, roles: [Role.Analyst, Role.Admin] },
    { label: 'Experiments', to: { name: 'experiments' }, roles: [Role.Analyst, Role.Admin] },
    { label: 'Users', to: { name: 'adminUsers' }, roles: [Role.Admin] },
    { label: 'IP Rules', to: { name: 'adminIpRules' }, roles: [Role.Admin] },
    { label: 'Monitor', to: { name: 'adminMonitor' }, roles: [Role.Admin] },
  ]
  return items.filter((i) => i.roles.includes(role as Role))
})

async function handleLogout() {
  await authStore.logout()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.default-layout { display: flex; min-height: 100vh; }
.sidebar { width: 220px; background: #1a1a2e; color: white; display: flex; flex-direction: column; }
.sidebar-header { padding: 1.5rem 1rem; border-bottom: 1px solid #2a2a4e; }
.sidebar-header h2 { font-size: 1.1rem; margin: 0; }
.sidebar-nav { flex: 1; padding: 0.5rem 0; }
.nav-item { display: block; padding: 0.75rem 1rem; color: #aaa; text-decoration: none; transition: all 0.2s; }
.nav-item:hover { color: white; background: #2a2a4e; }
.nav-item.active { color: white; background: #3a3a5e; border-left: 3px solid #646cff; }
.sidebar-footer { padding: 1rem; border-top: 1px solid #2a2a4e; }
.user-info { display: block; font-size: 0.85rem; margin-bottom: 0.5rem; }
.logout-btn { background: none; border: 1px solid #aaa; color: #aaa; padding: 0.4rem 0.8rem; border-radius: 4px; cursor: pointer; width: 100%; }
.logout-btn:hover { border-color: white; color: white; }
.main-content { flex: 1; display: flex; flex-direction: column; background: #f5f7fa; }
.top-bar { height: 56px; background: white; border-bottom: 1px solid #e0e0e0; display: flex; align-items: center; justify-content: flex-end; padding: 0 1.5rem; }
.notification-bell { text-decoration: none; color: #333; position: relative; }
.badge { background: #e53e3e; color: white; border-radius: 10px; padding: 0.1rem 0.4rem; font-size: 0.7rem; margin-left: 4px; }
.page-content { flex: 1; padding: 1.5rem; overflow-y: auto; }
</style>
