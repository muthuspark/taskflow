import api, { API_BASE_PATH } from './api'

/**
 * Analytics service for fetching analytics data
 */
const analyticsService = {
  /**
   * Get overall system statistics
   * @returns {Promise<object>}
   */
  async getOverview() {
    const response = await api.get(`${API_BASE_PATH}/analytics/overview`)
    return response.data.data
  },

  /**
   * Get execution trends over time
   * @param {number} days - Number of days to fetch (default: 30)
   * @returns {Promise<object>} - { trends: [], days: number }
   */
  async getExecutionTrends(days = 30) {
    const response = await api.get(`${API_BASE_PATH}/analytics/execution-trends?days=${days}`)
    return response.data.data
  },

  /**
   * Get statistics for all jobs
   * @returns {Promise<object>} - { jobs: [], total: number }
   */
  async getJobStats() {
    const response = await api.get(`${API_BASE_PATH}/analytics/job-stats`)
    return response.data.data
  },

  /**
   * Get duration trends for a specific job
   * @param {string} jobId - Job ID
   * @param {number} days - Number of days to fetch (default: 30)
   * @returns {Promise<object>} - { job_id, job_name, trends: [], days }
   */
  async getJobDurationTrends(jobId, days = 30) {
    const response = await api.get(`${API_BASE_PATH}/analytics/jobs/${jobId}/duration-trends?days=${days}`)
    return response.data.data
  }
}

export default analyticsService
