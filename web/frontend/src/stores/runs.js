import { ref } from 'vue'
import { defineStore } from 'pinia'
import runsService from '../services/runs'

const MAX_LOG_ENTRIES = 2000

export const useRunsStore = defineStore('runs', () => {
  // State
  const runs = ref([])
  const currentRun = ref(null)
  const logs = ref([])
  const totalLogs = ref(0)
  const logsOffset = ref(0)
  const loading = ref(false)
  const loadingMore = ref(false)
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
      // Load the last MAX_LOG_ENTRIES entries (tail of log)
      const { total } = await runsService.getLogs(id, 1, 0)
      totalLogs.value = total
      const offset = Math.max(0, total - MAX_LOG_ENTRIES)
      const result = await runsService.getLogs(id, MAX_LOG_ENTRIES, offset)
      logs.value = result.logs
      logsOffset.value = offset
      return logs.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch logs'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchEarlierLogs(id) {
    if (logsOffset.value <= 0 || loadingMore.value) return
    loadingMore.value = true
    try {
      const batchSize = MAX_LOG_ENTRIES
      const newOffset = Math.max(0, logsOffset.value - batchSize)
      const limit = logsOffset.value - newOffset
      const result = await runsService.getLogs(id, limit, newOffset)
      logs.value = [...result.logs, ...logs.value]
      logsOffset.value = newOffset
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch earlier logs'
    } finally {
      loadingMore.value = false
    }
  }

  function addLog(log) {
    logs.value.push(log)
    totalLogs.value++
    // Cap the in-memory log array to prevent unbounded growth
    if (logs.value.length > MAX_LOG_ENTRIES) {
      const excess = logs.value.length - MAX_LOG_ENTRIES
      logs.value.splice(0, excess)
      logsOffset.value += excess
    }
  }

  function clearLogs() {
    logs.value = []
    totalLogs.value = 0
    logsOffset.value = 0
  }

  function clearCurrentRun() {
    currentRun.value = null
    clearLogs()
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    runs,
    currentRun,
    logs,
    totalLogs,
    logsOffset,
    loading,
    loadingMore,
    error,
    // Actions
    fetchRuns,
    fetchRun,
    fetchLogs,
    fetchEarlierLogs,
    addLog,
    clearLogs,
    clearCurrentRun,
    clearError
  }
})
