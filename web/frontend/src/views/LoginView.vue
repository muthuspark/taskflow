<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import authService from '../services/auth'

const emit = defineEmits(['login-success'])

const authStore = useAuthStore()

// State
const username = ref('')
const password = ref('')
const email = ref('')
const error = ref('')
const loading = ref(false)
const setupRequired = ref(false)
const checkingSetup = ref(true)

// Check if setup is required on mount
onMounted(async () => {
  try {
    const status = await authService.checkSetupStatus()
    setupRequired.value = status.needs_setup
  } catch (e) {
    // Assume login mode if check fails
    setupRequired.value = false
  } finally {
    checkingSetup.value = false
  }
})

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    if (setupRequired.value) {
      await authStore.setupAdmin(username.value, password.value, email.value)
    } else {
      await authStore.login(username.value, password.value)
    }
    emit('login-success')
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Authentication failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="top-border"></div>
    <div class="login-container">
      <div class="login-box">
        <div class="login-header">
          <h1>TaskFlow</h1>
          <p v-if="checkingSetup" class="text-small mb-0">Loading...</p>
          <p v-else-if="setupRequired" class="text-small mb-0">Create Admin Account</p>
          <p v-else class="text-small mb-0">Sign in to continue</p>
        </div>

        <div class="login-body">
          <form v-if="!checkingSetup" @submit.prevent="handleSubmit">
            <div class="form-group">
              <label for="username">Username</label>
              <input
                id="username"
                v-model="username"
                type="text"
                placeholder="Enter username"
                required
                :disabled="loading"
              />
            </div>

            <div v-if="setupRequired" class="form-group">
              <label for="email">Email</label>
              <input
                id="email"
                v-model="email"
                type="email"
                placeholder="Enter email"
                required
                :disabled="loading"
              />
            </div>

            <div class="form-group">
              <label for="password">Password</label>
              <input
                id="password"
                v-model="password"
                type="password"
                placeholder="Enter password"
                required
                :disabled="loading"
              />
            </div>

            <div v-if="error" class="error-message mb-10">
              {{ error }}
            </div>

            <button type="submit" class="btn btn-primary" style="width: 100%;" :disabled="loading">
              <span v-if="loading">
                <span class="spinner" style="width: 12px; height: 12px; border-width: 2px; margin-right: 6px; vertical-align: middle;"></span>
                {{ setupRequired ? 'Creating...' : 'Signing in...' }}
              </span>
              <span v-else>
                {{ setupRequired ? 'Create Admin Account' : 'Sign In' }}
              </span>
            </button>
          </form>

          <div v-if="setupRequired && !checkingSetup" class="mt-15" style="background: #f4f4f4; padding: 10px; border: 1px solid #ccc;">
            <p class="text-small mb-0">This is the first time setup. Create your admin account to get started.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  background-color: #ffffff;
}
</style>
