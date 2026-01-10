import { ref } from 'vue'
import { defineStore } from 'pinia'
import analyticsService from '../services/analytics'

export const useAnalyticsStore = defineStore('analytics', () => {
  // State
  const overview = ref(null)
  const executionTrends = ref([])
  const jobStats = ref([])
  const selectedJobTrends = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Actions
  async function fetchOverview() {
    loading.value = true
    error.value = null
    try {
      overview.value = await analyticsService.getOverview()
      return overview.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch overview'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchExecutionTrends(days = 30) {
    loading.value = true
    error.value = null
    try {
      const data = await analyticsService.getExecutionTrends(days)
      executionTrends.value = data.trends || []
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch execution trends'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchJobStats() {
    loading.value = true
    error.value = null
    try {
      const data = await analyticsService.getJobStats()
      jobStats.value = data.jobs || []
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch job stats'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchJobDurationTrends(jobId, days = 30) {
    loading.value = true
    error.value = null
    try {
      const data = await analyticsService.getJobDurationTrends(jobId, days)
      selectedJobTrends.value = data
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch duration trends'
      throw e
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function clearSelectedJobTrends() {
    selectedJobTrends.value = null
  }

  return {
    // State
    overview,
    executionTrends,
    jobStats,
    selectedJobTrends,
    loading,
    error,
    // Actions
    fetchOverview,
    fetchExecutionTrends,
    fetchJobStats,
    fetchJobDurationTrends,
    clearError,
    clearSelectedJobTrends
  }
})
