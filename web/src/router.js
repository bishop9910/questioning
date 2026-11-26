import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from './api'

const routes = [
  { path: '/login', component: () => import('./pages/LoginPage.vue'), meta: { public: true } },
  { path: '/', redirect: '/surveys' },
  { path: '/surveys', component: () => import('./pages/SurveyListPage.vue') },
  { path: '/surveys/:id/edit', component: () => import('./pages/SurveyEditPage.vue') },
  { path: '/surveys/:id/stats', component: () => import('./pages/StatsPage.vue') },
  { path: '/s/:id', component: () => import('./pages/FillPage.vue'), meta: { public: true } },
  { path: '/:pathMatch(.*)*', redirect: '/surveys' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  if (!to.meta.public && !getToken()) return { path: '/login', query: { redirect: to.fullPath } }
  if (to.path === '/login' && getToken()) return '/surveys'
  return true
})

export default router
