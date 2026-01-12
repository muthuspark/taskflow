import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import authService from '../services/auth'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref(authService.getCurrentUser())

  // Computed
  const isAuthenticated = computed(() => {
    return !!user.value && authService.isAuthenticated()
  })

  const isAdmin = computed(() => {
    return user.value?.role === 'admin'
  })

  // Actions
  async function login(username, password) {
    try {
      const result = await authService.login(username, password)
      user.value = result.user
      return result
    } catch (error) {
      throw error
    }
  }

  async function setupAdmin(username, password, email) {
    try {
      const result = await authService.setupAdmin(username, password, email)
      user.value = result.user
      return result
    } catch (error) {
      throw error
    }
  }

  function logout() {
    authService.logout()
    user.value = null
  }

  function updateEmail(newEmail) {
    if (user.value) {
      user.value = { ...user.value, email: newEmail }
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  // Initialize from localStorage on store creation
  function initialize() {
    user.value = authService.getCurrentUser()
  }

  return {
    // State
    user,
    // Computed
    isAuthenticated,
    isAdmin,
    // Actions
    login,
    setupAdmin,
    logout,
    updateEmail,
    initialize
  }
})
