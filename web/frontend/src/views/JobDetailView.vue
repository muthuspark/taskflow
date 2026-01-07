<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useJobsStore } from '../stores/jobs'
import { useRunsStore } from '../stores/runs'
import jobsService from '../services/jobs'
import StatusBadge from '../components/StatusBadge.vue'
import ScheduleViewer from '../components/ScheduleViewer.vue'
import ScheduleEditor from '../components/ScheduleEditor.vue'
import JobEditForm from '../components/JobEditForm.vue'

const route = useRoute()
const router = useRouter()
const jobsStore = useJobsStore()
const runsStore = useRunsStore()

// State
const job = computed(() => jobsStore.currentJob)
const runs = computed(() => runsStore.runs)
const loading = computed(() => jobsStore.loading)
const error = computed(() => jobsStore.error)
const schedule = ref(null)
const scheduleLoading = ref(false)
const showEditForm = ref(false)
const showScheduleEditor = ref(false)
const runningJob = ref(false)

// Fetch job and runs
onMounted(async () => {
  const jobId = route.params.id
  await Promise.all([
    jobsStore.fetchJob(jobId),
    runsStore.fetchRuns(jobId, 20, 0),
    loadSchedule(jobId)
  ])
})

onUnmounted(() => {
  jobsStore.clearCurrentJob()
})

async function loadSchedule(jobId) {
  scheduleLoading.value = true
  try {
    schedule.value = await jobsService.getSchedule(jobId)
  } catch (e) {
    // Schedule might not exist
    schedule.value = null
  } finally {
    scheduleLoading.value = false
  }
}

async function handleRun() {
  if (runningJob.value) return
  runningJob.value = true
  try {
    const run = await jobsStore.triggerJob(job.value.id)
    router.push(`/runs/${run.id}`)
  } catch (e) {
    // Error handled by store
  } finally {
    runningJob.value = false
  }
}

async function handleDelete() {
  if (!confirm('Are you sure you want to delete this job? This action cannot be undone.')) {
    return
  }
  try {
    await jobsStore.deleteJob(job.value.id)
    router.push('/jobs')
  } catch (e) {
    // Error handled by store
  }
}

function handleEditSave() {
  showEditForm.value = false
  jobsStore.fetchJob(route.params.id)
}

async function handleScheduleSave(newSchedule) {
  try {
    await jobsService.setSchedule(job.value.id, newSchedule)
    schedule.value = newSchedule
    showScheduleEditor.value = false
  } catch (e) {
    alert(e.response?.data?.error || e.message || 'Failed to save schedule')
  }
}

function goToRun(runId) {
  router.push(`/runs/${runId}`)
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
</script>

<template>
  <div class="job-detail">
    <div v-if="loading && !job" class="loading-container">
      <div class="spinner-large"></div>
      <p>Loading job...</p>
    </div>

    <div v-else-if="error && !job" class="error-container">
      <p>{{ error }}</p>
      <button @click="router.push('/jobs')" class="btn btn-secondary">Back to Jobs</button>
    </div>

    <template v-else-if="job">
      <!-- Header -->
      <div class="page-header">
        <div class="header-left">
          <button @click="router.push('/jobs')" class="btn btn-link">
            &larr; Back to Jobs
          </button>
          <h1>{{ job.name }}</h1>
          <StatusBadge :status="job.enabled ? 'enabled' : 'disabled'" />
        </div>
        <div class="header-actions">
          <button @click="handleRun" class="btn btn-primary" :disabled="runningJob">
            <span v-if="runningJob">Running...</span>
            <span v-else>Run Now</span>
          </button>
          <button @click="showEditForm = true" class="btn btn-secondary">Edit</button>
          <button @click="handleDelete" class="btn btn-danger">Delete</button>
        </div>
      </div>

      <!-- Job Info Cards -->
      <div class="info-grid">
        <div class="info-card">
          <h3>Details</h3>
          <div class="info-row">
            <span class="label">Description:</span>
            <span class="value">{{ job.description || 'No description' }}</span>
          </div>
          <div class="info-row">
            <span class="label">Timezone:</span>
            <span class="value">{{ job.timezone || 'UTC' }}</span>
          </div>
          <div class="info-row">
            <span class="label">Timeout:</span>
            <span class="value">{{ job.timeout_seconds }}s</span>
          </div>
          <div class="info-row">
            <span class="label">Retries:</span>
            <span class="value">{{ job.retry_count }} (delay: {{ job.retry_delay_seconds }}s)</span>
          </div>
          <div class="info-row">
            <span class="label">Created:</span>
            <span class="value">{{ formatDate(job.created_at) }}</span>
          </div>
        </div>

        <div class="info-card">
          <div class="card-header">
            <h3>Schedule</h3>
            <button @click="showScheduleEditor = true" class="btn btn-small">
              {{ schedule ? 'Edit' : 'Add' }}
            </button>
          </div>
          <div v-if="scheduleLoading" class="loading-small">Loading...</div>
          <div v-else-if="schedule">
            <ScheduleViewer :schedule="schedule" />
          </div>
          <div v-else class="empty-schedule">
            <p>No schedule configured</p>
            <p class="hint">Job will only run when triggered manually</p>
          </div>
        </div>
      </div>

      <!-- Script Preview -->
      <div class="script-card">
        <h3>Script</h3>
        <pre class="script-preview">{{ job.script }}</pre>
      </div>

      <!-- Recent Runs -->
      <div class="runs-card">
        <div class="card-header">
          <h3>Recent Runs</h3>
          <router-link :to="`/runs?job_id=${job.id}`" class="btn btn-small">
            View All
          </router-link>
        </div>
        <div v-if="!runs.length" class="empty-runs">
          <p>No runs yet</p>
        </div>
        <table v-else class="runs-table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Started</th>
              <th>Duration</th>
              <th>Exit Code</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in runs" :key="run.id">
              <td><StatusBadge :status="run.status" /></td>
              <td>{{ formatDate(run.started_at) }}</td>
              <td>{{ formatDuration(run.duration_ms) }}</td>
              <td>{{ run.exit_code ?? '-' }}</td>
              <td>
                <button @click="goToRun(run.id)" class="btn btn-small">View</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- Edit Form Modal -->
    <JobEditForm
      v-if="showEditForm && job"
      :job="job"
      @save="handleEditSave"
      @cancel="showEditForm = false"
    />

    <!-- Schedule Editor Modal -->
    <ScheduleEditor
      v-if="showScheduleEditor && job"
      :schedule="schedule"
      @save="handleScheduleSave"
      @cancel="showScheduleEditor = false"
    />
  </div>
</template>

<style scoped>
.job-detail {
  padding: 0;
}

.loading-container,
.error-container {
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
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  color: var(--black);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.header-left h1 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.btn-link {
  background: none;
  border: none;
  color: var(--black);
  padding: 0;
  cursor: pointer;
  font-size: 0.875rem;
  text-decoration: underline;
  font-weight: 700;
}

.btn-link:hover {
  text-decoration: underline;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.info-card,
.script-card,
.runs-card {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  padding: 1.5rem;
  box-shadow: none;
}

.info-card h3,
.script-card h3,
.runs-card h3 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
}

.info-row {
  display: flex;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--gray-light);
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .label {
  font-weight: 900;
  color: var(--black);
  width: 100px;
  flex-shrink: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.75rem;
}

.info-row .value {
  color: var(--black);
}

.loading-small {
  color: var(--gray-dark);
  font-size: 0.875rem;
}

.empty-schedule,
.empty-runs {
  text-align: center;
  padding: 1rem;
  color: var(--gray-dark);
}

.empty-schedule .hint {
  font-size: 0.75rem;
  color: var(--gray-medium);
}

.script-card {
  margin-bottom: 1.5rem;
}

.script-preview {
  background: var(--gray-dark);
  color: var(--white);
  padding: 1rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  overflow-x: auto;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.runs-table {
  width: 100%;
  border-collapse: collapse;
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
</style>
