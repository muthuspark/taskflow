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
  <div class="main-container">
    <!-- Main Content Area -->
    <div class="content-area">
      <h1>TaskFlow - Task Scheduler Dashboard</h1>

      <p>TaskFlow provides a lightweight, self-hosted solution for scheduling and running tasks on your servers.</p>

      <div class="vision-box">
        <strong>Our goal:</strong> Provide the simplest, most reliable task scheduling solution for small to medium deployments.
      </div>

      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <p>Loading dashboard...</p>
      </div>

      <div v-else-if="error" class="error-message">
        {{ error }}
        <button @click="$router.go()" class="btn btn-small" style="margin-left: 10px;">Retry</button>
      </div>

      <template v-else>
        <hr class="section-divider">

        <h2>System Statistics</h2>
        <p class="table-title">Current system overview</p>

        <table>
          <thead>
            <tr>
              <th>Metric</th>
              <th>Value</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Total Jobs</td>
              <td class="number">{{ stats?.total_jobs || 0 }}</td>
              <td>Total number of configured jobs</td>
            </tr>
            <tr>
              <td>Active Jobs</td>
              <td class="number">{{ stats?.active_jobs || 0 }}</td>
              <td>Jobs currently enabled for scheduling</td>
            </tr>
            <tr>
              <td>Success Rate</td>
              <td class="number positive">{{ successRatePercent }}%</td>
              <td>Percentage of successful runs (last 7 days)</td>
            </tr>
            <tr>
              <td>Running Now</td>
              <td class="number" :class="{ positive: stats?.running_now > 0 }">{{ stats?.running_now || 0 }}</td>
              <td>Jobs currently executing</td>
            </tr>
          </tbody>
        </table>

        <hr class="section-divider">

        <h2>Recent Runs</h2>
        <p class="table-title">Most recent job executions</p>

        <div v-if="!stats?.recent_runs?.length" class="empty-state">
          <p class="mb-0">No recent runs found.</p>
        </div>

        <table v-else class="full-width">
          <thead>
            <tr>
              <th>#</th>
              <th>Job Name</th>
              <th>Status</th>
              <th>Started</th>
              <th>Duration</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(run, index) in stats.recent_runs" :key="run.id">
              <td>{{ index + 1 }}.</td>
              <td>
                <a href="#" @click.prevent="goToJob(run.job_id)">
                  {{ run.job_name || run.job_id }}
                </a>
              </td>
              <td>
                <StatusBadge :status="run.status" />
              </td>
              <td>{{ formatDate(run.started_at) }}</td>
              <td class="number">{{ formatDuration(run.duration_ms) }}</td>
              <td>
                <a href="#" @click.prevent="goToRun(run.id)">view logs</a>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="table-note">showing most recent executions</div>

        <p class="mt-15">
          Find more details in the <router-link to="/runs">run history</router-link>.
        </p>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <!-- Quick Actions Box -->
      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Quick Actions
        </div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <router-link to="/jobs/new" class="btn btn-primary" style="width: 100%; text-align: center;">
              Create New Job
            </router-link>
          </p>
          <p class="mb-10">
            <router-link to="/jobs" class="btn" style="width: 100%; text-align: center;">
              View All Jobs
            </router-link>
          </p>
          <p class="mb-0">
            <router-link to="/runs" class="btn" style="width: 100%; text-align: center;">
              View Run History
            </router-link>
          </p>
        </div>
      </div>

      <!-- System Info Box -->
      <div class="sidebar-box">
        <div class="sidebar-box-header">
          System Info
        </div>
        <div class="sidebar-box-content">
          <div class="news-item">
            <div class="news-title">TaskFlow v1.0</div>
            <div class="news-date">Self-hosted scheduler</div>
            <div class="news-text">
              Lightweight task scheduling with cron-like syntax, script execution, and real-time log streaming.
            </div>
          </div>
          <div class="news-item">
            <div class="news-title">Features</div>
            <div class="news-text">
              <ul style="margin: 0; padding-left: 20px;">
                <li>Cron-style scheduling</li>
                <li>Script execution</li>
                <li>Real-time WebSocket logs</li>
                <li>CPU/memory metrics</li>
                <li>JWT authentication</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Documentation Box -->
      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Documentation
        </div>
        <div class="sidebar-box-content">
          <p class="text-small">
            TaskFlow uses SQLite for storage and supports both manual and scheduled job execution.
          </p>
          <p class="text-small mb-0">
            Jobs run sequentially to avoid resource contention on single-server deployments.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Using global W3Techs-style CSS */
</style>
