import api, { API_BASE_PATH } from './api'

/**
 * Auth service for handling authentication operations
 */
const authService = {
  /**
   * Login with username and password
   * @param {string} username
   * @param {string} password
   * @returns {Promise<{token: string, user: object}>}
   */
  async login(username, password) {
    const response = await api.post(`${API_BASE_PATH}/auth/login`, { username, password })
    const { token, user } = response.data.data

    // Store token and user in localStorage
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))

    return { token, user }
  },

  /**
   * Setup first admin account
   * @param {string} username
   * @param {string} password
   * @param {string} email
   * @returns {Promise<{token: string, user: object}>}
   */
  async setupAdmin(username, password, email) {
    const response = await api.post('/setup/admin', { username, password, email })
    const { token, user } = response.data.data

    // Store token and user in localStorage
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))

    return { token, user }
  },

  /**
   * Check if setup is required (no users exist)
   * @returns {Promise<{setup_required: boolean}>}
   */
  async checkSetupStatus() {
    const response = await api.get('/setup/status')
    return response.data.data
  },

  /**
   * Logout and clear stored credentials
   */
  logout() {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  },

  /**
   * Get current user from localStorage
   * @returns {object|null}
   */
  getCurrentUser() {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        return JSON.parse(userStr)
      } catch {
        return null
      }
    }
    return null
  },

  /**
   * Check if user is authenticated
   * @returns {boolean}
   */
  isAuthenticated() {
    return !!localStorage.getItem('token')
  },

  /**
   * Change password for the current user
   * @param {string} currentPassword
   * @param {string} newPassword
   * @returns {Promise<{message: string}>}
   */
  async changePassword(currentPassword, newPassword) {
    const response = await api.put(`${API_BASE_PATH}/auth/password`, {
      current_password: currentPassword,
      new_password: newPassword
    })
    return response.data.data
  },

  /**
   * Update email for the current user
   * @param {string} email
   * @returns {Promise<{message: string, email: string}>}
   */
  async updateEmail(email) {
    const response = await api.put(`${API_BASE_PATH}/auth/email`, { email })
    return response.data.data
  },

  /**
   * Get SMTP settings (admin only)
   * @returns {Promise<{server: string, port: number, username: string, password: string, from_name: string, from_email: string}>}
   */
  async getSMTPSettings() {
    const response = await api.get(`${API_BASE_PATH}/settings/smtp`)
    return response.data.data
  },

  /**
   * Update SMTP settings (admin only)
   * @param {object} settings
   * @returns {Promise<{message: string}>}
   */
  async updateSMTPSettings(settings) {
    const response = await api.put(`${API_BASE_PATH}/settings/smtp`, settings)
    return response.data.data
  },

  /**
   * Test SMTP settings by sending a test email (admin only)
   * @returns {Promise<{message: string}>}
   */
  async testSMTPSettings() {
    const response = await api.post(`${API_BASE_PATH}/settings/smtp/test`)
    return response.data.data
  }
}

export default authService
