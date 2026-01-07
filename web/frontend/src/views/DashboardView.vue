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
  <div class="dashboard">
    <h1>Dashboard</h1>

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
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-value">{{ stats?.total_jobs || 0 }}</div>
          <div class="stat-label">Total Jobs</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats?.active_jobs || 0 }}</div>
          <div class="stat-label">Active Jobs</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ successRatePercent }}%</div>
          <div class="stat-label">Success Rate</div>
        </div>
        <div class="stat-card running">
          <div class="stat-value">{{ stats?.running_now || 0 }}</div>
          <div class="stat-label">Running Now</div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="quick-actions">
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
      <div class="recent-runs">
        <h2>Recent Runs</h2>
        <div v-if="!stats?.recent_runs?.length" class="empty-state">
          <p>No recent runs found</p>
        </div>
        <table v-else class="runs-table">
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
.dashboard {
  padding: 0;
}

.dashboard h1 {
  margin: 0 0 1.5rem 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
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
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  color: var(--black);
  font-weight: 900;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  padding: 1.5rem;
  text-align: center;
  transition: none;
}

.stat-card:hover {
  transform: none;
  box-shadow: none;
  background: var(--gray-lighter);
}

.stat-card.running {
  background: var(--black);
  color: var(--white);
  border: 1px solid var(--gray-light);
}

.stat-card.running:hover {
  background: var(--black);
  box-shadow: none;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 900;
  line-height: 1;
  margin-bottom: 0.5rem;
  text-transform: uppercase;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--gray-medium);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 900;
}

.stat-card.running .stat-label {
  color: var(--white);
}

.quick-actions {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  flex-wrap: wrap;
}

.recent-runs {
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  padding: 1.5rem;
  box-shadow: none;
}

.recent-runs h2 {
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--gray-light);
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--gray-medium);
}

.runs-table {
  width: 100%;
  border-collapse: collapse;
}

.runs-table th,
.runs-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border: 1px solid var(--gray-light);
  font-size: 0.875rem;
}

.runs-table th {
  font-weight: 900;
  color: var(--white);
  background-color: var(--black);
  text-transform: uppercase;
  letter-spacing: 0.05em;
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

.link {
  color: var(--black);
  text-decoration: underline;
  cursor: pointer;
  font-weight: 700;
}

.link:hover {
  background: var(--black);
  color: var(--white);
  text-decoration: none;
}
</style>
