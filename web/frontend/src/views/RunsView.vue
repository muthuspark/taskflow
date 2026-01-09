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
  <div class="main-container">
    <div class="content-area">
      <div class="page-header">
        <h1 style="margin: 0;">Run History</h1>
        <button @click="refresh" class="btn" :disabled="loading">
          Refresh
        </button>
      </div>

      <p>View the execution history of all jobs. Click on a run to see detailed logs and metrics.</p>

      <!-- Filters -->
      <div class="card">
        <div class="card-header">Filters</div>
        <div class="card-body">
          <div class="form-group mb-0">
            <label for="jobFilter">Filter by Job:</label>
            <select id="jobFilter" v-model="selectedJobId" :disabled="loading" style="max-width: 300px;">
              <option value="">All Jobs</option>
              <option v-for="job in jobs" :key="job.id" :value="job.id">
                {{ job.name }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <div v-if="loading && !runs.length" class="loading-container">
        <div class="spinner"></div>
        <p>Loading runs...</p>
      </div>

      <div v-else-if="error" class="error-message">
        {{ error }}
        <button @click="loadRuns" class="btn btn-small" style="margin-left: 10px;">Retry</button>
      </div>

      <div v-else-if="!runs.length" class="empty-state">
        <h2>No runs found</h2>
        <p v-if="selectedJobId">No runs found for this job. Try running it!</p>
        <p v-else>No job runs yet. Create a job and run it to see history here.</p>
        <router-link to="/jobs" class="btn btn-primary">Go to Jobs</router-link>
      </div>

      <template v-else>
        <h2>All Runs</h2>
        <p class="table-title">Job execution history</p>

        <table class="full-width">
          <thead>
            <tr>
              <th>#</th>
              <th>Job</th>
              <th>Status</th>
              <th>Started</th>
              <th>Finished</th>
              <th>Duration</th>
              <th>Exit</th>
              <th>Trigger</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(run, index) in runs" :key="run.id">
              <td>{{ index + 1 }}.</td>
              <td>
                <a href="#" @click.prevent="goToJob(run.job_id)">
                  {{ getJobName(run.job_id) }}
                </a>
              </td>
              <td><StatusBadge :status="run.status" /></td>
              <td>{{ formatDate(run.started_at) }}</td>
              <td>{{ formatDate(run.finished_at) }}</td>
              <td class="number">{{ formatDuration(run.duration_ms) }}</td>
              <td class="number">{{ run.exit_code ?? '-' }}</td>
              <td>
                <span class="badge">{{ run.trigger_type || 'manual' }}</span>
              </td>
              <td>
                <a href="#" @click.prevent="goToRun(run.id)">view logs</a>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="table-note">showing {{ runs.length }} run(s)</div>

        <div v-if="runs.length >= limit" class="text-center mt-15">
          <button @click="loadMore" class="btn" :disabled="loading">
            Load More
          </button>
        </div>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Quick Actions
        </div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <button @click="refresh" class="btn" style="width: 100%;" :disabled="loading">
              Refresh Data
            </button>
          </p>
          <p class="mb-0">
            <router-link to="/jobs" class="btn" style="width: 100%; text-align: center;">
              View All Jobs
            </router-link>
          </p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Run Status Legend
        </div>
        <div class="sidebar-box-content">
          <p class="text-small">
            <StatusBadge status="success" /> Job completed successfully
          </p>
          <p class="text-small">
            <StatusBadge status="failed" /> Job failed with error
          </p>
          <p class="text-small">
            <StatusBadge status="running" /> Job is currently executing
          </p>
          <p class="text-small mb-0">
            <StatusBadge status="pending" /> Job is queued for execution
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Using global W3Techs-style CSS */
</style>
