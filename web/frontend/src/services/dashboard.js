import api, { API_BASE_PATH } from './api'

/**
 * Dashboard service for fetching dashboard statistics
 */
const dashboardService = {
  /**
   * Get dashboard statistics
   * @returns {Promise<object>}
   */
  async getStats() {
    const response = await api.get(`${API_BASE_PATH}/dashboard/stats`)
    return response.data.data
  }
}

export default dashboardService
