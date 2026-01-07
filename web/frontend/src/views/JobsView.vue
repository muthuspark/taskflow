<script setup>
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useJobsStore } from '../stores/jobs'
import StatusBadge from '../components/StatusBadge.vue'

const router = useRouter()
const jobsStore = useJobsStore()

// Computed
const jobs = computed(() => jobsStore.jobs)
const loading = computed(() => jobsStore.loading)
const error = computed(() => jobsStore.error)

// Fetch jobs on mount
onMounted(() => {
  jobsStore.fetchJobs()
})

function goToCreate() {
  router.push('/jobs/new')
}

function goToJob(id) {
  router.push(`/jobs/${id}`)
}

async function handleRun(id) {
  try {
    const run = await jobsStore.triggerJob(id)
    router.push(`/runs/${run.id}`)
  } catch (e) {
    // Error is handled by store
  }
}

async function handleDelete(id) {
  if (confirm('Are you sure you want to delete this job?')) {
    try {
      await jobsStore.deleteJob(id)
    } catch (e) {
      // Error is handled by store
    }
  }
}

function retry() {
  jobsStore.clearError()
  jobsStore.fetchJobs()
}

function formatDate(dateStr) {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleDateString()
}

function getStatusText(job) {
  return job.enabled ? 'enabled' : 'disabled'
}
</script>

<template>
  <div class="jobs-view">
    <div class="page-header">
      <h1>Jobs</h1>
      <button @click="goToCreate" class="btn btn-primary">
        Create New Job
      </button>
    </div>

    <div v-if="loading && !jobs.length" class="loading-container">
      <div class="spinner-large"></div>
      <p>Loading jobs...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <p>{{ error }}</p>
      <button @click="retry" class="btn btn-primary">Retry</button>
    </div>

    <div v-else-if="!jobs.length" class="empty-state">
      <h2>No jobs yet</h2>
      <p>Create your first job to get started with task scheduling.</p>
      <button @click="goToCreate" class="btn btn-primary">
        Create Your First Job
      </button>
    </div>

    <div v-else class="jobs-table-container">
      <table class="jobs-table">
        <thead>
          <tr>
            <th class="col-name">Job Name</th>
            <th class="col-description">Description</th>
            <th class="col-status">Status</th>
            <th class="col-timeout">Timeout</th>
            <th class="col-retries">Retries</th>
            <th class="col-created">Created</th>
            <th class="col-actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in jobs" :key="job.id" class="job-row" @click="goToJob(job.id)">
            <td class="col-name">
              <span class="job-name-text">{{ job.name }}</span>
            </td>
            <td class="col-description">
              <span v-if="job.description" class="description-text">{{ job.description }}</span>
              <span v-else class="description-empty">No description</span>
            </td>
            <td class="col-status">
              <StatusBadge :status="getStatusText(job)" />
            </td>
            <td class="col-timeout">{{ job.timeout_seconds }}s</td>
            <td class="col-retries">{{ job.retry_count }}</td>
            <td class="col-created">{{ formatDate(job.created_at) }}</td>
            <td class="col-actions" @click.stop>
              <button @click="handleRun(job.id)" class="btn btn-primary btn-small">
                Run
              </button>
              <button @click="handleDelete(job.id)" class="btn btn-danger btn-small">
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.jobs-view {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  margin: 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.loading-container {
  text-align: center;
  padding: 3rem;
}

.spinner-large {
  width: 48px;
  height: 48px;
  border: 4px solid var(--gray-light);
  border-top-color: var(--black);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-container {
  text-align: center;
  padding: 3rem;
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  color: var(--black);
  font-weight: 900;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
}

.empty-state h2 {
  margin: 0 0 0.5rem 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.empty-state p {
  margin: 0 0 1.5rem 0;
  color: var(--gray-medium);
}

.jobs-table-container {
  background: var(--white);
  border: 1px solid var(--gray-light);
  overflow-x: auto;
}

.jobs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.jobs-table thead {
  background: var(--gray-lighter);
  border-bottom: 1px solid var(--gray-light);
}

.jobs-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 900;
  color: var(--black);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.75rem;
  border-right: 1px solid var(--gray-light);
}

.jobs-table th:last-child {
  border-right: none;
}

.jobs-table tbody tr {
  border-bottom: 1px solid var(--gray-light);
  cursor: pointer;
  transition: background-color 0.2s;
}

.jobs-table tbody tr:hover {
  background: var(--gray-lighter);
}

.jobs-table tbody tr:last-child {
  border-bottom: none;
}

.jobs-table td {
  padding: 1rem;
  border-right: 1px solid var(--gray-light);
  color: var(--black);
}

.jobs-table td:last-child {
  border-right: none;
}

.col-name {
  min-width: 150px;
  font-weight: 900;
}

.job-name-text {
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.col-description {
  min-width: 200px;
}

.description-text {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: var(--gray-medium);
}

.description-empty {
  color: var(--gray-medium);
  font-style: italic;
}

.col-status {
  min-width: 100px;
}

.col-timeout {
  min-width: 80px;
}

.col-retries {
  min-width: 80px;
}

.col-created {
  min-width: 100px;
  color: var(--gray-medium);
  font-size: 0.8125rem;
}

.col-actions {
  min-width: 140px;
  white-space: nowrap;
}

.btn-small {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  margin-right: 0.5rem;
}

.btn-small:last-child {
  margin-right: 0;
}
</style>
