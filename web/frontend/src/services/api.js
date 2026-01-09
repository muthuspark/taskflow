import axios from 'axios'

// Get API base path from environment variable or default to '/api'
// In production, this can be set via VITE_API_BASE_PATH at build time
const apiBasePath = import.meta.env.VITE_API_BASE_PATH || '/api'

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Export the base path for use in other services
export const API_BASE_PATH = apiBasePath

// Request interceptor to add Authorization header
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      // Clear auth data and redirect to login
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

export default api
