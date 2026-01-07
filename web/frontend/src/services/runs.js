import api from './api'

/**
 * Runs service for handling run operations
 */
const runsService = {
  /**
   * List runs with optional filtering
   * @param {string|null} jobId - Filter by job ID
   * @param {number} limit - Max results
   * @param {number} offset - Pagination offset
   * @returns {Promise<Array>}
   */
  async list(jobId = null, limit = 100, offset = 0) {
    const params = new URLSearchParams()
    if (jobId) {
      params.append('job_id', jobId)
    }
    params.append('limit', limit.toString())
    params.append('offset', offset.toString())

    const response = await api.get(`/api/runs?${params.toString()}`)
    return response.data.data.runs
  },

  /**
   * Get a single run by ID
   * @param {string} id
   * @returns {Promise<object>}
   */
  async get(id) {
    const response = await api.get(`/api/runs/${id}`)
    return response.data.data
  },

  /**
   * Get logs for a run
   * @param {string} id
   * @returns {Promise<Array>}
   */
  async getLogs(id) {
    const response = await api.get(`/api/runs/${id}/logs`)
    return response.data.data.logs
  }
}

export default runsService
