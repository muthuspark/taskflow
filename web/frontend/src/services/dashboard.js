import api from './api'

/**
 * Dashboard service for fetching dashboard statistics
 */
const dashboardService = {
  /**
   * Get dashboard statistics
   * @returns {Promise<object>}
   */
  async getStats() {
    const response = await api.get('/api/dashboard/stats')
    return response.data.data
  }
}

export default dashboardService
