import api, { API_BASE_PATH } from './api'

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

    const response = await api.get(`${API_BASE_PATH}/runs?${params.toString()}`)
    return response.data.data.runs
  },

  /**
   * Get a single run by ID
   * @param {string} id
   * @returns {Promise<object>}
   */
  async get(id) {
    const response = await api.get(`${API_BASE_PATH}/runs/${id}`)
    return response.data.data
  },

  /**
   * Get logs for a run with optional pagination
   * @param {string} id
   * @param {number|null} limit - Max log entries to return
   * @param {number|null} offset - Offset for pagination
   * @returns {Promise<{logs: Array, total: number}>}
   */
  async getLogs(id, limit = null, offset = null) {
    const params = new URLSearchParams()
    if (limit != null) params.append('limit', limit.toString())
    if (offset != null) params.append('offset', offset.toString())
    const query = params.toString()
    const url = `${API_BASE_PATH}/runs/${id}/logs${query ? '?' + query : ''}`
    const response = await api.get(url)
    return { logs: response.data.data.logs || [], total: response.data.data.total || 0 }
  }
}

export default runsService
