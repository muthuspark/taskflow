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
  <div class="flex items-center justify-center min-h-screen bg-white p-4">
    <div class="bg-white border-2 border-gray-light p-8 w-full max-w-[400px]">
      <div class="text-center mb-8 pb-6 border-b border-gray-light">
        <h1 class="text-4xl text-black m-0 mb-2 font-black uppercase tracking-tight">TaskFlow</h1>
        <p v-if="checkingSetup" class="text-gray-medium m-0 font-bold">Loading...</p>
        <p v-else-if="setupRequired" class="text-gray-medium m-0 font-bold">Create Admin Account</p>
        <p v-else class="text-gray-medium m-0 font-bold">Sign in to continue</p>
      </div>

      <form v-if="!checkingSetup" @submit.prevent="handleSubmit" class="flex flex-col gap-4">
        <div class="form-group">
          <label for="username" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="Enter username"
            required
            :disabled="loading"
            class="w-full px-3 py-3 text-base text-black bg-white border border-gray-light transition-none font-inherit font-medium"
          />
        </div>

        <div v-if="setupRequired" class="form-group">
          <label for="email" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="Enter email"
            required
            :disabled="loading"
            class="w-full px-3 py-3 text-base text-black bg-white border border-gray-light transition-none font-inherit font-medium"
          />
        </div>

        <div class="form-group">
          <label for="password" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Enter password"
            required
            :disabled="loading"
            class="w-full px-3 py-3 text-base text-black bg-white border border-gray-light transition-none font-inherit font-medium"
          />
        </div>

        <div v-if="error" class="bg-white text-black px-4 py-3 border border-gray-light text-sm text-center font-black">
          {{ error }}
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          <span v-if="loading">
            <span class="spinner-sm inline-block mr-2 align-middle"></span>
            {{ setupRequired ? 'Creating...' : 'Signing in...' }}
          </span>
          <span v-else>
            {{ setupRequired ? 'Create Admin Account' : 'Sign In' }}
          </span>
        </button>
      </form>

      <div v-if="setupRequired && !checkingSetup" class="mt-6 p-4 bg-gray-lighter border border-gray-light text-center">
        <p class="m-0 text-black text-sm font-bold">This is the first time setup. Create your admin account to get started.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* All styles now use Tailwind utilities in the template */
</style>
