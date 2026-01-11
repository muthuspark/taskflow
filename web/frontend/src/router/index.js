import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import JobsView from '../views/JobsView.vue'
import JobCreateView from '../views/JobCreateView.vue'
import JobDetailView from '../views/JobDetailView.vue'
import RunsView from '../views/RunsView.vue'
import RunDetailView from '../views/RunDetailView.vue'
import RunLogsPrintView from '../views/RunLogsPrintView.vue'
import AnalyticsView from '../views/AnalyticsView.vue'
import AccountView from '../views/AccountView.vue'

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: DashboardView,
    meta: { requiresAuth: false }
  },
  {
    path: '/jobs',
    name: 'jobs',
    component: JobsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/jobs/new',
    name: 'job-create',
    component: JobCreateView,
    meta: { requiresAuth: true }
  },
  {
    path: '/jobs/:id',
    name: 'job-detail',
    component: JobDetailView,
    meta: { requiresAuth: true }
  },
  {
    path: '/runs',
    name: 'runs',
    component: RunsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/runs/:id',
    name: 'run-detail',
    component: RunDetailView,
    meta: { requiresAuth: true }
  },
  {
    path: '/runs/:id/logs',
    name: 'run-logs-print',
    component: RunLogsPrintView,
    meta: { requiresAuth: true }
  },
  {
    path: '/analytics',
    name: 'analytics',
    component: AnalyticsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/account',
    name: 'account',
    component: AccountView,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard - allow all routes, let App.vue handle auth display
router.beforeEach((to, from, next) => {
  next()
})

export default router
