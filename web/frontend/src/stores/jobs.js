import { ref } from 'vue'
import { defineStore } from 'pinia'
import jobsService from '../services/jobs'

export const useJobsStore = defineStore('jobs', () => {
  // State
  const jobs = ref([])
  const currentJob = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Actions
  async function fetchJobs() {
    loading.value = true
    error.value = null
    try {
      jobs.value = await jobsService.list()
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch jobs'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchJob(id) {
    loading.value = true
    error.value = null
    try {
      currentJob.value = await jobsService.get(id)
      return currentJob.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to fetch job'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function createJob(job) {
    loading.value = true
    error.value = null
    try {
      const newJob = await jobsService.create(job)
      jobs.value.push(newJob)
      return newJob
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to create job'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateJob(id, job) {
    loading.value = true
    error.value = null
    try {
      const updatedJob = await jobsService.update(id, job)
      const index = jobs.value.findIndex(j => j.id === id)
      if (index !== -1) {
        jobs.value[index] = updatedJob
      }
      if (currentJob.value?.id === id) {
        currentJob.value = updatedJob
      }
      return updatedJob
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to update job'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteJob(id) {
    loading.value = true
    error.value = null
    try {
      await jobsService.delete(id)
      jobs.value = jobs.value.filter(j => j.id !== id)
      if (currentJob.value?.id === id) {
        currentJob.value = null
      }
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to delete job'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function triggerJob(id) {
    loading.value = true
    error.value = null
    try {
      const run = await jobsService.run(id)
      return run
    } catch (e) {
      error.value = e.response?.data?.error || e.message || 'Failed to trigger job'
      throw e
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function clearCurrentJob() {
    currentJob.value = null
  }

  return {
    // State
    jobs,
    currentJob,
    loading,
    error,
    // Actions
    fetchJobs,
    fetchJob,
    createJob,
    updateJob,
    deleteJob,
    triggerJob,
    clearError,
    clearCurrentJob
  }
})
