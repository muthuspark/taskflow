import api from './api'

/**
 * Jobs service for handling job operations
 */
const jobsService = {
  /**
   * List all jobs
   * @returns {Promise<Array>}
   */
  async list() {
    const response = await api.get('/api/jobs')
    return response.data.data
  },

  /**
   * Get a single job by ID
   * @param {string} id
   * @returns {Promise<object>}
   */
  async get(id) {
    const response = await api.get(`/api/jobs/${id}`)
    return response.data.data
  },

  /**
   * Create a new job
   * @param {object} job - Job data
   * @returns {Promise<object>}
   */
  async create(job) {
    const response = await api.post('/api/jobs', job)
    return response.data.data
  },

  /**
   * Update an existing job
   * @param {string} id
   * @param {object} job - Updated job data
   * @returns {Promise<object>}
   */
  async update(id, job) {
    const response = await api.put(`/api/jobs/${id}`, job)
    return response.data.data
  },

  /**
   * Delete a job
   * @param {string} id
   * @returns {Promise<void>}
   */
  async delete(id) {
    await api.delete(`/api/jobs/${id}`)
  },

  /**
   * Trigger a job to run immediately
   * @param {string} id
   * @returns {Promise<object>} - Run object
   */
  async run(id) {
    const response = await api.post(`/api/jobs/${id}/run`)
    return response.data.data
  },

  /**
   * Get job schedule
   * @param {string} id
   * @returns {Promise<object>}
   */
  async getSchedule(id) {
    const response = await api.get(`/api/jobs/${id}/schedule`)
    return response.data.data
  },

  /**
   * Set job schedule
   * @param {string} id
   * @param {object} schedule - Schedule data
   * @returns {Promise<object>}
   */
  async setSchedule(id, schedule) {
    const response = await api.put(`/api/jobs/${id}/schedule`, schedule)
    return response.data.data
  }
}

export default jobsService
