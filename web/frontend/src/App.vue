<template>
  <div id="app">
    <header>
      <h1>TaskFlow</h1>
      <nav v-if="isLoggedIn">
        <router-link to="/">Dashboard</router-link>
        <router-link to="/jobs">Jobs</router-link>
        <button @click="logout">Logout</button>
      </nav>
    </header>

    <main>
      <div v-if="!isLoggedIn" class="login-container">
        <h2>Login</h2>
        <form @submit.prevent="handleLogin">
          <input
            v-model="loginForm.username"
            type="text"
            placeholder="Username"
            required
          />
          <input
            v-model="loginForm.password"
            type="password"
            placeholder="Password"
            required
          />
          <button type="submit">Login</button>
          <p v-if="loginError" class="error">{{ loginError }}</p>
        </form>

        <div v-if="showSetupForm">
          <h3>First Time Setup</h3>
          <form @submit.prevent="handleSetup">
            <input
              v-model="setupForm.username"
              type="text"
              placeholder="Admin Username"
              required
            />
            <input
              v-model="setupForm.password"
              type="password"
              placeholder="Admin Password"
              required
            />
            <input
              v-model="setupForm.email"
              type="email"
              placeholder="Email (optional)"
            />
            <button type="submit">Create Admin Account</button>
            <p v-if="setupError" class="error">{{ setupError }}</p>
          </form>
        </div>
      </div>

      <div v-else class="dashboard">
        <h2>Dashboard</h2>
        <p>Welcome, {{ currentUser?.username }}!</p>
        <div class="stats">
          <div class="stat">
            <h3>Active Jobs</h3>
            <p>{{ stats.activeJobs || 0 }}</p>
          </div>
          <div class="stat">
            <h3>Success Rate</h3>
            <p>{{ (stats.successRate * 100).toFixed(1) }}%</p>
          </div>
          <div class="stat">
            <h3>Running Now</h3>
            <p>{{ stats.runningNow || 0 }}</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const isLoggedIn = ref(false)
const currentUser = ref(null)
const showSetupForm = ref(false)
const loginForm = ref({ username: '', password: '' })
const setupForm = ref({ username: '', password: '', email: '' })
const loginError = ref('')
const setupError = ref('')
const stats = ref({
  activeJobs: 0,
  successRate: 0,
  runningNow: 0
})

const api = axios.create({
  baseURL: '/',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add token to requests
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

onMounted(async () => {
  // Check if user is logged in
  const token = localStorage.getItem('token')
  if (token) {
    isLoggedIn.value = true
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    currentUser.value = user
  } else {
    // Check if setup is needed
    try {
      const response = await api.get('/setup/status')
      if (response.data.data.needs_setup) {
        showSetupForm.value = true
      }
    } catch (error) {
      console.error('Failed to check setup status:', error)
    }
  }
})

const handleLogin = async () => {
  loginError.value = ''
  try {
    const response = await api.post('/api/auth/login', {
      username: loginForm.value.username,
      password: loginForm.value.password
    })

    const { token, user } = response.data.data
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))

    isLoggedIn.value = true
    currentUser.value = user

    loginForm.value = { username: '', password: '' }
  } catch (error) {
    loginError.value = error.response?.data?.error || 'Login failed'
  }
}

const handleSetup = async () => {
  setupError.value = ''
  try {
    const response = await api.post('/setup/admin', {
      username: setupForm.value.username,
      password: setupForm.value.password,
      email: setupForm.value.email
    })

    const { token, user } = response.data.data
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))

    isLoggedIn.value = true
    currentUser.value = user
    showSetupForm.value = false

    setupForm.value = { username: '', password: '', email: '' }
  } catch (error) {
    setupError.value = error.response?.data?.error || 'Setup failed'
  }
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  isLoggedIn.value = false
  currentUser.value = null
}
</script>

<style scoped>
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  border-bottom: 1px solid #ccc;
  background-color: #f5f5f5;
}

header h1 {
  margin: 0;
}

nav {
  display: flex;
  gap: 1rem;
  align-items: center;
}

nav a {
  color: #0066cc;
  text-decoration: none;
  padding: 0.5rem 1rem;
}

nav a:hover {
  text-decoration: underline;
}

main {
  padding: 2rem;
}

.login-container {
  max-width: 400px;
  margin: 2rem auto;
  padding: 2rem;
  border: 1px solid #ccc;
  border-radius: 8px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

input {
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}

.error {
  color: red;
  font-size: 0.9rem;
}

.dashboard {
  max-width: 1000px;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-top: 2rem;
}

.stat {
  padding: 1rem;
  border: 1px solid #ccc;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.stat h3 {
  margin: 0 0 0.5rem 0;
  font-size: 0.9rem;
  color: #666;
}

.stat p {
  margin: 0;
  font-size: 2rem;
  font-weight: bold;
  color: #0066cc;
}
</style>
