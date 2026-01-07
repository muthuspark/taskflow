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
      <header class="app-header">
        <div class="header-left">
          <router-link to="/" class="logo">TaskFlow</router-link>
        </div>
        <nav class="main-nav">
          <router-link to="/" :class="{ active: isActive('/') && route.path === '/' }">
            Dashboard
          </router-link>
          <router-link to="/jobs" :class="{ active: isActive('/jobs') }">
            Jobs
          </router-link>
          <router-link to="/runs" :class="{ active: isActive('/runs') }">
            Runs
          </router-link>
        </nav>
        <div class="header-right">
          <span class="user-info">
            {{ currentUser?.username }}
            <span v-if="currentUser?.role === 'admin'" class="role-badge">Admin</span>
          </span>
          <button @click="logout" class="btn btn-logout">Logout</button>
        </div>
      </header>

      <main class="app-main">
        <router-view />
      </main>
    </template>
  </div>
</template>

<style scoped>
/* Uses global color variables from style.css */
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2rem;
  height: 60px;
  background: var(--white);
  border-bottom: 1px solid var(--gray-light);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo {
  font-size: 1.5rem;
  font-weight: 900;
  color: var(--black);
  text-decoration: none;
  letter-spacing: -0.5px;
  text-transform: uppercase;
}

.logo:hover {
  color: var(--black);
}

.main-nav {
  display: flex;
  gap: 0.5rem;
}

.main-nav a {
  color: var(--black);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 0;
  font-weight: 700;
  transition: none;
  border: 2px solid transparent;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.main-nav a:hover {
  color: var(--black);
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
}

.main-nav a.active {
  color: var(--white);
  background: var(--black);
  border: 1px solid var(--gray-light);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--black);
  font-weight: 700;
}

.role-badge {
  background: var(--black);
  color: var(--white);
  font-size: 0.625rem;
  padding: 0.125rem 0.5rem;
  border-radius: 0;
  text-transform: uppercase;
  font-weight: 900;
  letter-spacing: 0.05em;
  border: 1px solid var(--gray-light);
}

.btn-logout {
  background: var(--white);
  border: 1px solid var(--gray-light);
  color: var(--black);
  padding: 0.375rem 0.75rem;
  border-radius: 0;
  font-size: 0.875rem;
  cursor: pointer;
  transition: none;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.btn-logout:hover {
  background: var(--black);
  color: var(--white);
  border: 1px solid var(--gray-light);
}

.app-main {
  flex: 1;
  padding: 2rem;
  background: var(--white);
}

@media (max-width: 768px) {
  .app-header {
    padding: 0 1rem;
    flex-wrap: wrap;
    height: auto;
    gap: 0.5rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
  }

  .main-nav {
    order: 3;
    width: 100%;
    justify-content: center;
    padding-top: 0.5rem;
    border-top: 3px solid var(--black);
  }

  .app-main {
    padding: 1rem;
  }
}
</style>
