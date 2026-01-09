<script setup>
import { computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import LoginView from './views/LoginView.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const currentUser = computed(() => authStore.user)

// Initialize auth state on mount
onMounted(() => {
  authStore.initialize()
})

function handleLoginSuccess() {
  // Navigate to dashboard after login
  router.push('/')
}

function logout() {
  authStore.logout()
  router.push('/')
}

function isActive(path) {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}
</script>

<template>
  <div id="app">
    <!-- Show login view when not authenticated -->
    <LoginView v-if="!isAuthenticated" @login-success="handleLoginSuccess" />

    <!-- Show main app when authenticated -->
    <template v-else>
      <!-- Top gradient border -->
      <div class="top-border"></div>

      <!-- Navigation -->
      <header class="app-header">
        <div class="nav-container">
          <nav class="main-nav">
            <router-link to="/" :class="{ active: isActive('/') && route.path === '/' }">Home</router-link>
            <router-link to="/jobs" :class="{ active: isActive('/jobs') }">Jobs</router-link>
            <router-link to="/runs" :class="{ active: isActive('/runs') }">Runs</router-link>
            <router-link to="/jobs/new" :class="{ active: route.path === '/jobs/new' }">Create Job</router-link>
            <span style="color: #999;">|</span>
            <span style="font-size: 11px; color: #666;">
              {{ currentUser?.username }}
              <span v-if="currentUser?.role === 'admin'" style="color: #0066cc;">(admin)</span>
            </span>
            <a href="#" @click.prevent="logout">Logout</a>
          </nav>
        </div>
      </header>

      <!-- Featured Banner -->
      <div class="featured-banner">
        TaskFlow - Lightweight Task Scheduler
      </div>

      <!-- Main Content -->
      <main>
        <router-view />
      </main>
    </template>
  </div>
</template>

<style scoped>
/* No scoped styles needed - using global CSS */
</style>
