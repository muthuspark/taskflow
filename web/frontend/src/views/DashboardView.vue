<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import dashboardService from '../services/dashboard'
import StatusBadge from '../components/StatusBadge.vue'

const router = useRouter()

// State
const stats = ref(null)
const loading = ref(true)
const error = ref(null)

// Computed
const successRatePercent = computed(() => {
  if (!stats.value) return '0'
  return (stats.value.success_rate * 100).toFixed(1)
})

// Fetch dashboard stats
onMounted(async () => {
  try {
    stats.value = await dashboardService.getStats()
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load dashboard'
  } finally {
    loading.value = false
  }
})

function formatDuration(ms) {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds}s`
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

function goToRun(runId) {
  router.push(`/runs/${runId}`)
}

function goToJob(jobId) {
  router.push(`/jobs/${jobId}`)
}
</script>

<template>
  <div>
    <h1 class="mb-6 text-black font-black uppercase tracking-tight">Dashboard</h1>

    <div v-if="loading" class="loading-container">
      <div class="spinner-large"></div>
      <p>Loading dashboard...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <p>{{ error }}</p>
      <button @click="$router.go()" class="btn btn-primary">Retry</button>
    </div>

    <template v-else>
      <!-- Stats Cards -->
      <div class="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-4 mb-8">
        <div class="stat-card bg-white border border-gray-light p-6 text-center hover:bg-gray-lighter transition-none">
          <div class="text-5xl font-black leading-none mb-2 uppercase">{{ stats?.total_jobs || 0 }}</div>
          <div class="text-sm text-gray-medium uppercase tracking-tight font-black">Total Jobs</div>
        </div>
        <div class="stat-card bg-white border border-gray-light p-6 text-center hover:bg-gray-lighter transition-none">
          <div class="text-5xl font-black leading-none mb-2 uppercase">{{ stats?.active_jobs || 0 }}</div>
          <div class="text-sm text-gray-medium uppercase tracking-tight font-black">Active Jobs</div>
        </div>
        <div class="stat-card bg-white border border-gray-light p-6 text-center hover:bg-gray-lighter transition-none">
          <div class="text-5xl font-black leading-none mb-2 uppercase">{{ successRatePercent }}%</div>
          <div class="text-sm text-gray-medium uppercase tracking-tight font-black">Success Rate</div>
        </div>
        <div class="stat-card bg-black text-white border border-gray-light p-6 text-center hover:bg-black transition-none">
          <div class="text-5xl font-black leading-none mb-2 uppercase">{{ stats?.running_now || 0 }}</div>
          <div class="text-sm text-white uppercase tracking-tight font-black">Running Now</div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="flex gap-4 mb-8 flex-wrap">
        <router-link to="/jobs/new" class="btn btn-primary">
          Create New Job
        </router-link>
        <router-link to="/jobs" class="btn btn-secondary">
          View All Jobs
        </router-link>
        <router-link to="/runs" class="btn btn-secondary">
          View All Runs
        </router-link>
      </div>

      <!-- Recent Runs Table -->
      <div class="bg-white border border-gray-light p-6">
        <h2 class="m-0 mb-4 text-2xl text-black font-black uppercase tracking-tight pb-4 border-b border-gray-light">Recent Runs</h2>
        <div v-if="!stats?.recent_runs?.length" class="empty-state">
          <p>No recent runs found</p>
        </div>
        <table v-else class="w-full border-collapse">
          <thead>
            <tr>
              <th>Job</th>
              <th>Status</th>
              <th>Started</th>
              <th>Duration</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in stats.recent_runs" :key="run.id">
              <td>
                <a @click.prevent="goToJob(run.job_id)" href="#" class="link">
                  {{ run.job_name || run.job_id }}
                </a>
              </td>
              <td>
                <StatusBadge :status="run.status" />
              </td>
              <td>{{ formatDate(run.started_at) }}</td>
              <td>{{ formatDuration(run.duration_ms) }}</td>
              <td>
                <button @click="goToRun(run.id)" class="btn btn-small">
                  View Logs
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
/* All styles handled by Tailwind utilities and global CSS classes */
</style>
