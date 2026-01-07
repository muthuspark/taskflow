import { ref } from 'vue'
import { defineStore } from 'pinia'
import runsService from '../services/runs'

export const useRunsStore = defineStore('runs', () => {
  // State
  const runs = ref([])
  const currentRun = ref(null)
  const logs = ref([])
  const loading = ref(false)
  const error = ref(null)

  // Actions
  async function fetchRuns(jobId = null, limit = 100, offset = 0) {
    loading.value = true
    error.value = null
    try {
      runs.value = await runsService.list(jobId, limit, offset)
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch runs'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchRun(id) {
    loading.value = true
    error.value = null
    try {
      currentRun.value = await runsService.get(id)
      return currentRun.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch run'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchLogs(id) {
    loading.value = true
    error.value = null
    try {
      logs.value = await runsService.getLogs(id)
      return logs.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch logs'
      throw e
    } finally {
      loading.value = false
    }
  }

  function addLog(log) {
    logs.value.push(log)
  }

  function clearLogs() {
    logs.value = []
  }

  function clearCurrentRun() {
    currentRun.value = null
    logs.value = []
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    runs,
    currentRun,
    logs,
    loading,
    error,
    // Actions
    fetchRuns,
    fetchRun,
    fetchLogs,
    addLog,
    clearLogs,
    clearCurrentRun,
    clearError
  }
})
