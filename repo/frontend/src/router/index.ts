import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  // Auth (guest only)
  {
    path: '/auth',
    component: () => import('@/layouts/AuthLayout.vue'),
    children: [
      { path: 'login', name: 'login', component: () => import('@/pages/auth/LoginPage.vue') },
      { path: 'register', name: 'register', component: () => import('@/pages/auth/RegisterPage.vue') },
    ],
  },
  // Authenticated
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'home', redirect: { name: 'catalog' } },
      { path: 'catalog', name: 'catalog', component: () => import('@/pages/catalog/CatalogPage.vue') },
      { path: 'catalog/:id', name: 'itemDetail', component: () => import('@/pages/catalog/ItemDetailPage.vue'), props: true },
      { path: 'catalog/:itemId/review', name: 'writeReview', component: () => import('@/pages/reviews/WriteReviewPage.vue'), props: true },
      { path: 'favorites', name: 'favorites', component: () => import('@/pages/favorites/FavoritesPage.vue') },
      // Moderation
      { path: 'moderation', name: 'moderationQueue', component: () => import('@/pages/moderation/ModerationQueuePage.vue'), meta: { roles: ['moderator', 'admin'] } },
      { path: 'moderation/words', name: 'sensitiveWords', component: () => import('@/pages/moderation/SensitiveWordsPage.vue'), meta: { roles: ['moderator', 'admin'] } },
      // Analytics
      { path: 'analytics', name: 'analytics', component: () => import('@/pages/analytics/AnalyticsDashboardPage.vue'), meta: { roles: ['product_analyst', 'admin'] } },
      { path: 'analytics/shared/:token', name: 'sharedAnalytics', component: () => import('@/pages/analytics/AnalyticsDashboardPage.vue'), props: (route) => ({ sharedToken: route.params.token, readonly: true }) },
      // Experiments
      { path: 'experiments', name: 'experiments', component: () => import('@/pages/experiments/ExperimentsListPage.vue'), meta: { roles: ['product_analyst', 'admin'] } },
      { path: 'experiments/:id', name: 'experimentDetail', component: () => import('@/pages/experiments/ExperimentDetailPage.vue'), props: true, meta: { roles: ['product_analyst', 'admin'] } },
      // Admin
      { path: 'admin/users', name: 'adminUsers', component: () => import('@/pages/admin/UserManagementPage.vue'), meta: { roles: ['admin'] } },
      { path: 'admin/ip-rules', name: 'adminIpRules', component: () => import('@/pages/admin/IpManagementPage.vue'), meta: { roles: ['admin'] } },
      { path: 'admin/monitor', name: 'adminMonitor', component: () => import('@/pages/admin/SystemMonitorPage.vue'), meta: { roles: ['admin'] } },
    ],
  },
  // Error pages
  { path: '/forbidden', name: 'forbidden', component: () => import('@/pages/errors/ForbiddenPage.vue') },
  { path: '/:pathMatch(.*)*', name: 'notFound', component: () => import('@/pages/errors/NotFoundPage.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guards
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('access_token')
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null

  // Auth required?
  if (to.matched.some((r) => r.meta.requiresAuth) && !token) {
    return next({ name: 'login' })
  }

  // Guest only?
  if (to.path.startsWith('/auth') && token) {
    return next({ name: 'catalog' })
  }

  // Role check
  const requiredRoles = to.meta.roles as string[] | undefined
  if (requiredRoles && user) {
    if (!requiredRoles.includes(user.role)) {
      return next({ name: 'forbidden' })
    }
  }

  next()
})

export default router
