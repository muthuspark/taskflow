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
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>TaskFlow</h1>
        <p v-if="checkingSetup">Loading...</p>
        <p v-else-if="setupRequired">Create Admin Account</p>
        <p v-else>Sign in to continue</p>
      </div>

      <form v-if="!checkingSetup" @submit.prevent="handleSubmit" class="login-form">
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

        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          <span v-if="loading">
            <span class="spinner"></span>
            {{ setupRequired ? 'Creating...' : 'Signing in...' }}
          </span>
          <span v-else>
            {{ setupRequired ? 'Create Admin Account' : 'Sign In' }}
          </span>
        </button>
      </form>

      <div v-if="setupRequired && !checkingSetup" class="setup-notice">
        <p>This is the first time setup. Create your admin account to get started.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--white);
  padding: 1rem;
}

.login-card {
  background: var(--white);
  border-radius: 0;
  border: 2px solid var(--gray-light);
  box-shadow: none;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--gray-light);
}

.login-header h1 {
  font-size: 2rem;
  color: var(--black);
  margin: 0 0 0.5rem 0;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.login-header p {
  color: var(--gray-medium);
  margin: 0;
  font-weight: 700;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 900;
  color: var(--black);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.form-group input {
  padding: 0.75rem 1rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-size: 1rem;
  transition: none;
  background: var(--white);
  color: var(--black);
  font-family: inherit;
  font-weight: 500;
}

.form-group input:focus {
  outline: none;
  border: 1px solid var(--gray-light);
  box-shadow: none;
}

.form-group input:disabled {
  background: var(--gray-lighter);
  cursor: not-allowed;
  opacity: 0.7;
}

.error-message {
  background: var(--white);
  color: var(--black);
  padding: 0.75rem 1rem;
  border-radius: 0;
  border: 1px solid var(--gray-light);
  font-size: 0.875rem;
  text-align: center;
  font-weight: 900;
}

.btn-block {
  width: 100%;
  padding: 0.875rem 1rem;
  font-size: 1rem;
}

.setup-notice {
  margin-top: 1.5rem;
  padding: 1rem;
  background: var(--gray-lighter);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  text-align: center;
}

.setup-notice p {
  margin: 0;
  color: var(--black);
  font-size: 0.875rem;
  font-weight: 700;
}

.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 0.5rem;
  vertical-align: middle;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
