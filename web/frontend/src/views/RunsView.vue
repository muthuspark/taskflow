<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRunsStore } from '../stores/runs'
import { useJobsStore } from '../stores/jobs'
import StatusBadge from '../components/StatusBadge.vue'

const route = useRoute()
const router = useRouter()
const runsStore = useRunsStore()
const jobsStore = useJobsStore()

// State
const runs = computed(() => runsStore.runs)
const loading = computed(() => runsStore.loading)
const error = computed(() => runsStore.error)
const jobs = computed(() => jobsStore.jobs)
const selectedJobId = ref(route.query.job_id || '')
const limit = ref(50)
const offset = ref(0)

// Fetch data on mount
onMounted(async () => {
  await jobsStore.fetchJobs()
  await loadRuns()
})

// Watch for filter changes
watch(selectedJobId, () => {
  offset.value = 0
  loadRuns()
})

async function loadRuns() {
  await runsStore.fetchRuns(
    selectedJobId.value || null,
    limit.value,
    offset.value
  )
}

function goToRun(runId) {
  router.push(`/runs/${runId}`)
}

function goToJob(jobId) {
  router.push(`/jobs/${jobId}`)
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

function formatDuration(ms) {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds}s`
}

function getJobName(jobId) {
  const job = jobs.value.find(j => j.id === jobId)
  return job?.name || jobId
}

function loadMore() {
  offset.value += limit.value
  loadRuns()
}

function refresh() {
  offset.value = 0
  loadRuns()
}
</script>

<template>
  <div class="runs-view">
    <div class="page-header">
      <h1>Run History</h1>
      <button @click="refresh" class="btn btn-secondary" :disabled="loading">
        Refresh
      </button>
    </div>

    <!-- Filters -->
    <div class="filters">
      <div class="filter-group">
        <label for="jobFilter">Filter by Job:</label>
        <select id="jobFilter" v-model="selectedJobId" :disabled="loading">
          <option value="">All Jobs</option>
          <option v-for="job in jobs" :key="job.id" :value="job.id">
            {{ job.name }}
          </option>
        </select>
      </div>
    </div>

    <div v-if="loading && !runs.length" class="loading-container">
      <div class="spinner-large"></div>
      <p>Loading runs...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <p>{{ error }}</p>
      <button @click="loadRuns" class="btn btn-primary">Retry</button>
    </div>

    <div v-else-if="!runs.length" class="empty-state">
      <h2>No runs found</h2>
      <p v-if="selectedJobId">No runs found for this job. Try running it!</p>
      <p v-else>No job runs yet. Create a job and run it to see history here.</p>
      <router-link to="/jobs" class="btn btn-primary">Go to Jobs</router-link>
    </div>

    <template v-else>
      <div class="runs-table-container">
        <table class="runs-table">
          <thead>
            <tr>
              <th>Job</th>
              <th>Status</th>
              <th>Started</th>
              <th>Finished</th>
              <th>Duration</th>
              <th>Exit Code</th>
              <th>Trigger</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in runs" :key="run.id">
              <td>
                <a @click.prevent="goToJob(run.job_id)" href="#" class="link">
                  {{ getJobName(run.job_id) }}
                </a>
              </td>
              <td><StatusBadge :status="run.status" /></td>
              <td>{{ formatDate(run.started_at) }}</td>
              <td>{{ formatDate(run.finished_at) }}</td>
              <td>{{ formatDuration(run.duration_ms) }}</td>
              <td>{{ run.exit_code ?? '-' }}</td>
              <td>
                <span class="trigger-badge" :class="run.trigger_type || 'manual'">
                  {{ run.trigger_type || 'manual' }}
                </span>
              </td>
              <td>
                <button @click="goToRun(run.id)" class="btn btn-small">
                  View Logs
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="runs.length >= limit" class="load-more">
        <button @click="loadMore" class="btn btn-secondary" :disabled="loading">
          Load More
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.runs-view {
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

.filters {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  padding: 1rem 1.5rem;
  box-shadow: none;
  margin-bottom: 1.5rem;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-weight: 900;
  color: var(--black);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.filter-group select {
  padding: 0;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-size: 0.875rem;
  min-width: 200px;
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
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  color: var(--black);
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  box-shadow: none;
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
  color: var(--gray-dark);
}

.runs-table-container {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  box-shadow: none;
  overflow: auto;
}

.runs-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 800px;
}

.runs-table th,
.runs-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--gray-light);
}

.runs-table th {
  font-weight: 900;
  color: var(--black);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: var(--gray-lighter);
  position: sticky;
  top: 0;
}

.runs-table tbody tr {
  background: var(--white);
}

.runs-table tbody tr:nth-child(even) {
  background: var(--gray-lighter);
}

.runs-table tbody tr:hover {
  background: var(--gray-light);
}

.runs-table tbody tr:last-child td {
  border-bottom: 1px solid var(--gray-light);
}

.link {
  color: var(--black);
  text-decoration: underline;
  cursor: pointer;
  font-weight: 700;
}

.link:hover {
  text-decoration: underline;
}

.trigger-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: var(--white);
  color: var(--black);
  font-weight: 700;
}

.trigger-badge.manual {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}

.trigger-badge.scheduled {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}

.trigger-badge.api {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}

.load-more {
  text-align: center;
  padding: 1.5rem;
}
</style>
