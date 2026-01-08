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
  <div>
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
      <div class="flex justify-between items-start flex-wrap gap-4 mb-6">
        <div class="flex flex-col gap-2">
          <button @click="router.push('/jobs')" class="bg-none border-none text-black p-0 cursor-pointer text-sm underline font-bold text-left">
            &larr; Back to Jobs
          </button>
          <div class="flex items-center gap-3">
            <h1 class="m-0 text-black font-black uppercase tracking-tight">{{ job.name }}</h1>
            <StatusBadge :status="job.enabled ? 'enabled' : 'disabled'" />
          </div>
        </div>
        <div class="flex gap-2">
          <button @click="handleRun" class="btn btn-primary" :disabled="runningJob">
            <span v-if="runningJob">Running...</span>
            <span v-else>Run Now</span>
          </button>
          <button @click="showEditForm = true" class="btn btn-secondary">Edit</button>
          <button @click="handleDelete" class="btn btn-danger">Delete</button>
        </div>
      </div>

      <!-- Job Info Cards -->
      <div class="grid grid-cols-[repeat(auto-fit,minmax(300px,1fr))] gap-4 mb-6">
        <div class="bg-white border border-gray-light p-6">
          <h3 class="m-0 mb-4 text-black font-black uppercase tracking-tight text-sm pb-2 border-b border-gray-light">Details</h3>
          <div class="flex py-2 border-b border-gray-light last:border-b-0">
            <span class="font-black text-black text-xs uppercase tracking-tight w-[100px] flex-shrink-0">Description:</span>
            <span class="text-black">{{ job.description || 'No description' }}</span>
          </div>
          <div class="flex py-2 border-b border-gray-light last:border-b-0">
            <span class="font-black text-black text-xs uppercase tracking-tight w-[100px] flex-shrink-0">Timezone:</span>
            <span class="text-black">{{ job.timezone || 'UTC' }}</span>
          </div>
          <div class="flex py-2 border-b border-gray-light last:border-b-0">
            <span class="font-black text-black text-xs uppercase tracking-tight w-[100px] flex-shrink-0">Timeout:</span>
            <span class="text-black">{{ job.timeout_seconds }}s</span>
          </div>
          <div class="flex py-2 border-b border-gray-light last:border-b-0">
            <span class="font-black text-black text-xs uppercase tracking-tight w-[100px] flex-shrink-0">Retries:</span>
            <span class="text-black">{{ job.retry_count }} (delay: {{ job.retry_delay_seconds }}s)</span>
          </div>
          <div class="flex py-2 border-b border-gray-light last:border-b-0">
            <span class="font-black text-black text-xs uppercase tracking-tight w-[100px] flex-shrink-0">Created:</span>
            <span class="text-gray-medium text-[0.8125rem]">{{ formatDate(job.created_at) }}</span>
          </div>
        </div>

        <div class="bg-white border border-gray-light p-6">
          <div class="flex justify-between items-center mb-4 pb-2 border-b border-gray-light">
            <h3 class="m-0 text-black font-black uppercase tracking-tight text-sm">Schedule</h3>
            <button @click="showScheduleEditor = true" class="btn btn-small">
              {{ schedule ? 'Edit' : 'Add' }}
            </button>
          </div>
          <div v-if="scheduleLoading" class="text-gray-dark text-sm">Loading...</div>
          <div v-else-if="schedule">
            <ScheduleViewer :schedule="schedule" />
          </div>
          <div v-else class="text-center py-4 text-gray-dark">
            <p class="m-0">No schedule configured</p>
            <p class="m-0 text-xs text-gray-medium">Job will only run when triggered manually</p>
          </div>
        </div>
      </div>

      <!-- Script Preview -->
      <div class="bg-white border border-gray-light p-6 mb-6">
        <h3 class="m-0 mb-4 text-black font-black uppercase tracking-tight text-sm pb-2 border-b border-gray-light">Script</h3>
        <pre class="script-preview">{{ job.script }}</pre>
      </div>

      <!-- Recent Runs -->
      <div class="bg-white border border-gray-light p-6">
        <div class="flex justify-between items-center mb-4 pb-2 border-b border-gray-light">
          <h3 class="m-0 text-black font-black uppercase tracking-tight text-sm">Recent Runs</h3>
          <router-link :to="`/runs?job_id=${job.id}`" class="btn btn-small">
            View All
          </router-link>
        </div>
        <div v-if="!runs.length" class="text-center py-4 text-gray-dark">
          <p>No runs yet</p>
        </div>
        <table v-else class="w-full border-collapse text-sm">
          <thead class="bg-gray-lighter border-b border-gray-light">
            <tr>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Status</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Started</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Duration</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Exit Code</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(run, index) in runs" :key="run.id" :class="index % 2 === 0 ? 'bg-white' : 'bg-gray-lighter'" class="border-b border-gray-light hover:bg-gray-light">
              <td class="px-4 py-3 border-r border-gray-light"><StatusBadge :status="run.status" /></td>
              <td class="px-4 py-3 text-gray-medium text-[0.8125rem] border-r border-gray-light">{{ formatDate(run.started_at) }}</td>
              <td class="px-4 py-3 text-black border-r border-gray-light">{{ formatDuration(run.duration_ms) }}</td>
              <td class="px-4 py-3 text-black border-r border-gray-light">{{ run.exit_code ?? '-' }}</td>
              <td class="px-4 py-3">
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
/* Script preview - specialized dark theme styling */
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
</style>
