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
  <div id="app" class="min-h-screen flex flex-col">
    <!-- Show login view when not authenticated -->
    <LoginView v-if="!isAuthenticated" @login-success="handleLoginSuccess" />

    <!-- Show main app when authenticated -->
    <template v-else>
      <header class="app-header flex justify-between items-center px-8 h-[60px] bg-white border-b border-gray-light sticky top-0 z-100">
        <div class="flex items-center">
          <router-link to="/" class="logo text-3xl font-black text-black no-underline uppercase" style="letter-spacing: -0.5px">TaskFlow</router-link>
        </div>
        <nav class="main-nav flex gap-1">
          <router-link
            to="/"
            class="main-nav-link px-4 py-2 text-black no-underline font-bold uppercase tracking-tight transition-none"
            :class="{ 'bg-black text-white': isActive('/') && route.path === '/', 'hover:bg-gray-lighter': !(isActive('/') && route.path === '/') }">
            Dashboard
          </router-link>
          <router-link
            to="/jobs"
            class="main-nav-link px-4 py-2 text-black no-underline font-bold uppercase tracking-tight transition-none"
            :class="{ 'bg-black text-white': isActive('/jobs'), 'hover:bg-gray-lighter': !isActive('/jobs') }">
            Jobs
          </router-link>
          <router-link
            to="/runs"
            class="main-nav-link px-4 py-2 text-black no-underline font-bold uppercase tracking-tight transition-none"
            :class="{ 'bg-black text-white': isActive('/runs'), 'hover:bg-gray-lighter': !isActive('/runs') }">
            Runs
          </router-link>
        </nav>
        <div class="header-right flex items-center gap-4">
          <span class="user-info flex items-center gap-2 text-sm text-black font-bold">
            {{ currentUser?.username }}
            <span v-if="currentUser?.role === 'admin'" class="role-badge inline-block bg-black text-white text-xs px-2 py-0.5 uppercase font-black tracking-tight border border-gray-light">Admin</span>
          </span>
          <button @click="logout" class="btn btn-small">Logout</button>
        </div>
      </header>

      <main class="flex-1 p-8 bg-white">
        <router-view />
      </main>
    </template>
  </div>
</template>

<style scoped>
/* Responsive overrides for mobile */
@media (max-width: 768px) {
  .app-header {
    @apply flex-wrap h-auto gap-2 p-2;
  }

  .main-nav {
    @apply w-full justify-center py-2 border-t-4 border-black order-3;
  }

  main {
    @apply p-4;
  }
}
</style>
