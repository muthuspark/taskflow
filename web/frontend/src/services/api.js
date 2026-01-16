import axios from 'axios'

// API base path - will be loaded from backend config
let apiBasePath = import.meta.env.VITE_API_BASE_PATH || '/taskflow/api'

// Derive setup base path from API base path (e.g., /taskflow/api -> /taskflow/setup)
function deriveSetupBasePath(apiPath) {
  if (apiPath.endsWith('/api')) {
    return apiPath.slice(0, -4) + '/setup'
  }
  return '/taskflow/setup'
}
let setupBasePath = deriveSetupBasePath(apiBasePath)

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Export getter for the base paths (allows dynamic updates)
export let API_BASE_PATH = apiBasePath
export let SETUP_BASE_PATH = setupBasePath

/**
 * Initialize API configuration from backend
 * This should be called once at app startup
 */
export async function initApiConfig() {
  try {
    const response = await axios.get('/taskflow-app/config')
    if (response.data?.data?.api_base_path) {
      apiBasePath = response.data.data.api_base_path
      API_BASE_PATH = apiBasePath
      setupBasePath = deriveSetupBasePath(apiBasePath)
      SETUP_BASE_PATH = setupBasePath
    }
  } catch (error) {
    // Use default if config endpoint fails
    console.warn('Failed to load API config, using default:', apiBasePath)
  }
  return apiBasePath
}

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
