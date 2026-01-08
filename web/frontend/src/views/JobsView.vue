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
  <div>
    <div class="page-header flex justify-between items-center mb-6">
      <h1 class="m-0 text-black font-black uppercase tracking-tight">Jobs</h1>
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

    <div v-else class="bg-white border border-gray-light overflow-x-auto">
      <table class="w-full border-collapse text-sm">
        <thead class="bg-gray-lighter border-b border-gray-light">
          <tr>
            <th class="col-name px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Job Name</th>
            <th class="col-description px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Description</th>
            <th class="col-status px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Status</th>
            <th class="col-timeout px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Timeout</th>
            <th class="col-retries px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Retries</th>
            <th class="col-created px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs border-r border-gray-light">Created</th>
            <th class="col-actions px-4 py-4 text-left font-black text-black uppercase tracking-tight text-xs">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in jobs" :key="job.id" class="border-b border-gray-light cursor-pointer hover:bg-gray-lighter transition-none" @click="goToJob(job.id)">
            <td class="col-name px-4 py-4 text-black font-black uppercase tracking-tight border-r border-gray-light">
              {{ job.name }}
            </td>
            <td class="col-description px-4 py-4 text-gray-medium border-r border-gray-light">
              <span v-if="job.description" class="description-text">{{ job.description }}</span>
              <span v-else class="italic">No description</span>
            </td>
            <td class="col-status px-4 py-4 border-r border-gray-light">
              <StatusBadge :status="getStatusText(job)" />
            </td>
            <td class="col-timeout px-4 py-4 text-black border-r border-gray-light">{{ job.timeout_seconds }}s</td>
            <td class="col-retries px-4 py-4 text-black border-r border-gray-light">{{ job.retry_count }}</td>
            <td class="col-created px-4 py-4 text-gray-medium text-[0.8125rem] border-r border-gray-light">{{ formatDate(job.created_at) }}</td>
            <td class="col-actions px-4 py-4 whitespace-nowrap" @click.stop>
              <button @click="handleRun(job.id)" class="btn btn-primary btn-small mr-2">
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
/* Column widths for table */
.col-name { min-width: 150px; }
.col-description { min-width: 200px; }
.col-status { min-width: 100px; }
.col-timeout { min-width: 80px; }
.col-retries { min-width: 80px; }
.col-created { min-width: 100px; }
.col-actions { min-width: 140px; }

/* Line clamping for description (1 line max) */
.description-text {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
