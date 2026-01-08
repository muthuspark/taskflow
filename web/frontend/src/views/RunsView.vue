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
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="m-0 text-black font-black uppercase tracking-tight">Run History</h1>
      <button @click="refresh" class="btn btn-secondary" :disabled="loading">
        Refresh
      </button>
    </div>

    <!-- Filters -->
    <div class="bg-white border border-gray-light p-6 mb-6">
      <div class="flex items-center gap-3">
        <label for="jobFilter" class="font-black text-black text-sm uppercase tracking-tight">Filter by Job:</label>
        <select id="jobFilter" v-model="selectedJobId" :disabled="loading" class="px-3 py-2 border border-gray-light text-sm">
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
      <div class="bg-white border border-gray-light overflow-auto">
        <table class="w-full border-collapse text-sm">
          <thead class="bg-gray-lighter border-b border-gray-light">
            <tr>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Job</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Status</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Started</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Finished</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Duration</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Exit Code</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight border-r border-gray-light">Trigger</th>
              <th class="px-4 py-3 text-left font-black text-black text-xs uppercase tracking-tight">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(run, index) in runs" :key="run.id" :class="index % 2 === 0 ? 'bg-white' : 'bg-gray-lighter'" class="border-b border-gray-light hover:bg-gray-light">
              <td class="px-4 py-3 border-r border-gray-light">
                <a @click.prevent="goToJob(run.job_id)" href="#" class="text-black underline font-bold cursor-pointer">
                  {{ getJobName(run.job_id) }}
                </a>
              </td>
              <td class="px-4 py-3 border-r border-gray-light"><StatusBadge :status="run.status" /></td>
              <td class="px-4 py-3 text-gray-medium text-[0.8125rem] border-r border-gray-light">{{ formatDate(run.started_at) }}</td>
              <td class="px-4 py-3 text-gray-medium text-[0.8125rem] border-r border-gray-light">{{ formatDate(run.finished_at) }}</td>
              <td class="px-4 py-3 text-black border-r border-gray-light">{{ formatDuration(run.duration_ms) }}</td>
              <td class="px-4 py-3 text-black border-r border-gray-light">{{ run.exit_code ?? '-' }}</td>
              <td class="px-4 py-3 border-r border-gray-light">
                <span class="trigger-badge" :class="run.trigger_type || 'manual'">
                  {{ run.trigger_type || 'manual' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <button @click="goToRun(run.id)" class="btn btn-small">
                  View Logs
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="runs.length >= limit" class="text-center py-6">
        <button @click="loadMore" class="btn btn-secondary" :disabled="loading">
          Load More
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
/* Trigger badge styling - component-specific states */
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

.trigger-badge.manual,
.trigger-badge.scheduled,
.trigger-badge.api {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}
</style>
