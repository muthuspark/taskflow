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

async function handleEnable() {
  try {
    await jobsStore.updateJob(job.value.id, { ...job.value, enabled: true })
  } catch (e) {
    // Error handled by store
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
  <div class="main-container">
    <div class="content-area">
      <div v-if="loading && !job" class="loading-container">
        <div class="spinner"></div>
        <p>Loading job...</p>
      </div>

      <div v-else-if="error && !job" class="error-message">
        {{ error }}
        <button @click="router.push('/jobs')" class="btn btn-small" style="margin-left: 10px;">Back to Jobs</button>
      </div>

      <template v-else-if="job">
        <!-- Back Link -->
        <p class="back-link">
          <a href="#" @click.prevent="router.push('/jobs')">&larr; Back to Jobs</a>
        </p>

        <!-- Header -->
        <div class="page-header">
          <div>
            <h1 style="margin: 0;">{{ job.name }}</h1>
            <p v-if="job.description" class="job-description">{{ job.description }}</p>
          </div>
          <div class="header-actions">
            <button
              v-if="job.enabled"
              @click="handleRun"
              class="btn btn-primary"
              :disabled="runningJob"
            >
              <span v-if="runningJob">Running...</span>
              <span v-else>Run Now</span>
            </button>
            <button v-else @click="handleEnable" class="btn btn-primary">
              Enable
            </button>
            <button @click="showEditForm = true" class="btn">Edit</button>
            <button @click="handleDelete" class="btn btn-danger">Delete</button>
          </div>
        </div>

        <!-- Job Details Table -->
        <div class="card">
          <div class="card-header">
            Job Details
            <StatusBadge :status="job.enabled ? 'enabled' : 'disabled'" />
          </div>
          <div class="card-body">
            <table class="details-table">
              <tbody>
                <tr>
                  <th>Status</th>
                  <td><StatusBadge :status="job.enabled ? 'enabled' : 'disabled'" /></td>
                </tr>
                <tr>
                  <th>Timezone</th>
                  <td>{{ job.timezone || 'UTC' }}</td>
                </tr>
                <tr>
                  <th>Timeout</th>
                  <td>{{ job.timeout_seconds }} seconds</td>
                </tr>
                <tr>
                  <th>Retries</th>
                  <td>{{ job.retry_count }} attempts ({{ job.retry_delay_seconds }}s delay)</td>
                </tr>
                <tr>
                  <th>Working Dir</th>
                  <td><code>{{ job.working_dir || '/tmp' }}</code></td>
                </tr>
                <tr>
                  <th>Created</th>
                  <td>{{ formatDate(job.created_at) }}</td>
                </tr>
                <tr>
                  <th>Updated</th>
                  <td>{{ formatDate(job.updated_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Schedule -->
        <div class="card">
          <div class="card-header">
            Schedule
            <button @click="showScheduleEditor = true" class="btn btn-small">
              {{ schedule ? 'Edit' : 'Add Schedule' }}
            </button>
          </div>
          <div class="card-body">
            <div v-if="scheduleLoading" class="text-muted">Loading...</div>
            <div v-else-if="schedule">
              <ScheduleViewer :schedule="schedule" />
            </div>
            <div v-else class="empty-schedule">
              <p>No schedule configured</p>
              <p class="text-small text-muted">Job will only run when triggered manually</p>
            </div>
          </div>
        </div>

        <!-- Script -->
        <div class="card">
          <div class="card-header">Script</div>
          <div class="card-body" style="padding: 0;">
            <pre class="script-preview">{{ job.script }}</pre>
          </div>
        </div>

        <!-- Recent Runs -->
        <div class="card">
          <div class="card-header">
            Recent Runs
            <router-link :to="`/runs?job_id=${job.id}`" class="btn btn-small">View All</router-link>
          </div>
          <div class="card-body" style="padding: 0;">
            <div v-if="!runs.length" class="empty-state" style="margin: 0; border: 0;">
              <p>No runs yet. Click "Run Now" to execute this job.</p>
            </div>
            <table v-else class="full-width">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Status</th>
                  <th>Started</th>
                  <th>Duration</th>
                  <th>Exit Code</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(run, index) in runs" :key="run.id">
                  <td>{{ index + 1 }}.</td>
                  <td><StatusBadge :status="run.status" /></td>
                  <td>{{ formatDate(run.started_at) }}</td>
                  <td class="number">{{ formatDuration(run.duration_ms) }}</td>
                  <td class="number">{{ run.exit_code ?? '-' }}</td>
                  <td>
                    <a href="#" @click.prevent="goToRun(run.id)">view logs</a>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">Quick Actions</div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <button
              v-if="job?.enabled"
              @click="handleRun"
              class="btn btn-primary"
              style="width: 100%;"
              :disabled="runningJob"
            >
              {{ runningJob ? 'Running...' : 'Run Now' }}
            </button>
            <button
              v-else
              @click="handleEnable"
              class="btn btn-primary"
              style="width: 100%;"
            >
              Enable
            </button>
          </p>
          <p class="mb-10">
            <button @click="showEditForm = true" class="btn" style="width: 100%;">
              Edit Job
            </button>
          </p>
          <p class="mb-0">
            <router-link :to="`/runs?job_id=${job?.id}`" class="btn" style="width: 100%; text-align: center;">
              View All Runs
            </router-link>
          </p>
        </div>
      </div>

      <div class="sidebar-box" v-if="job">
        <div class="sidebar-box-header">Job Info</div>
        <div class="sidebar-box-content">
          <p class="text-small"><strong>ID:</strong> {{ job.id }}</p>
          <p class="text-small"><strong>Status:</strong> {{ job.enabled ? 'Enabled' : 'Disabled' }}</p>
          <p class="text-small mb-0"><strong>Timezone:</strong> {{ job.timezone || 'UTC' }}</p>
        </div>
      </div>
    </div>

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
.back-link {
  margin-bottom: 10px;
}

.back-link a {
  font-size: 12px;
}

.job-description {
  color: #666666;
  margin: 5px 0 0 0;
  font-size: 13px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.details-table {
  width: 100%;
  margin: 0;
}

.details-table th {
  width: 120px;
  background: #f4f4f4;
  font-weight: bold;
  text-align: left;
}

.details-table td code {
  background: #f4f4f4;
  padding: 2px 6px;
  border: 1px solid #cccccc;
  font-family: "Courier New", monospace;
  font-size: 11px;
}

.empty-schedule {
  text-align: center;
  padding: 20px;
  color: #666666;
}

.empty-schedule p {
  margin: 0;
}

.script-preview {
  background: #333333;
  color: #ffffff;
  padding: 15px;
  margin: 0;
  font-family: "Courier New", monospace;
  font-size: 12px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border: none;
}
</style>
