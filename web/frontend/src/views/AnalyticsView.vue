<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAnalyticsStore } from '../stores/analytics'
import ExecutionTrendsChart from '../components/ExecutionTrendsChart.vue'
import JobDurationChart from '../components/JobDurationChart.vue'
import JobStatsTable from '../components/JobStatsTable.vue'

const analyticsStore = useAnalyticsStore()

// State
const loading = computed(() => analyticsStore.loading)
const error = computed(() => analyticsStore.error)
const overview = computed(() => analyticsStore.overview)
const executionTrends = computed(() => analyticsStore.executionTrends)
const jobStats = computed(() => analyticsStore.jobStats)
const selectedJobTrends = computed(() => analyticsStore.selectedJobTrends)

const selectedDays = ref(30)
const showSuccessRate = ref(true)
const selectedJob = ref(null)

// Load data on mount
onMounted(async () => {
  await Promise.all([
    analyticsStore.fetchOverview(),
    analyticsStore.fetchExecutionTrends(selectedDays.value),
    analyticsStore.fetchJobStats()
  ])
})

// Functions
async function changeDays(days) {
  selectedDays.value = days
  await analyticsStore.fetchExecutionTrends(days)
  if (selectedJob.value) {
    await analyticsStore.fetchJobDurationTrends(selectedJob.value.job_id, days)
  }
}

async function selectJob(job) {
  selectedJob.value = job
  await analyticsStore.fetchJobDurationTrends(job.job_id, selectedDays.value)
}

function clearSelectedJob() {
  selectedJob.value = null
  analyticsStore.clearSelectedJobTrends()
}

function formatDuration(ms) {
  if (!ms || ms === 0) return '-'
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

function formatSuccessRate(rate) {
  if (rate === undefined || rate === null) return '-'
  return (rate * 100).toFixed(1) + '%'
}
</script>

<template>
  <div class="main-container">
    <div class="content-area">
      <h1>Analytics</h1>
      <p>View execution trends, job performance statistics, and duration analysis.</p>

      <div v-if="loading && !overview" class="loading-container">
        <div class="spinner"></div>
        <p>Loading analytics...</p>
      </div>

      <div v-else-if="error && !overview" class="error-message">
        {{ error }}
        <button @click="$router.go()" class="btn btn-small" style="margin-left: 10px;">Retry</button>
      </div>

      <template v-else>
        <!-- Overview Stats -->
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ overview?.total_runs || 0 }}</div>
            <div class="stat-label">Total Runs</div>
          </div>
          <div class="stat-card">
            <div class="stat-value success">{{ formatSuccessRate(overview?.success_rate) }}</div>
            <div class="stat-label">Success Rate</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ overview?.runs_last_24h || 0 }}</div>
            <div class="stat-label">Last 24 Hours</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ formatDuration(overview?.avg_duration_ms) }}</div>
            <div class="stat-label">Avg Duration</div>
          </div>
        </div>

        <hr class="section-divider">

        <!-- Execution Trends -->
        <div class="section-header">
          <h2>Execution Trends</h2>
          <div class="section-controls">
            <div class="toggle-group">
              <button
                :class="['toggle-btn', { active: showSuccessRate }]"
                @click="showSuccessRate = true"
              >
                Success Rate
              </button>
              <button
                :class="['toggle-btn', { active: !showSuccessRate }]"
                @click="showSuccessRate = false"
              >
                Run Counts
              </button>
            </div>
            <div class="days-selector">
              <button
                v-for="days in [7, 14, 30, 90]"
                :key="days"
                :class="['btn', 'btn-small', { active: selectedDays === days }]"
                @click="changeDays(days)"
              >
                {{ days }}d
              </button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <ExecutionTrendsChart
              :trends="executionTrends"
              :showSuccessRate="showSuccessRate"
            />
          </div>
        </div>

        <hr class="section-divider">

        <!-- Job Statistics -->
        <h2>Job Statistics</h2>
        <p class="table-title">Performance statistics for all jobs (click "Trends" to view duration history)</p>

        <div class="card">
          <div class="card-body" style="padding: 0;">
            <JobStatsTable
              :jobs="jobStats"
              @select-job="selectJob"
            />
          </div>
        </div>

        <!-- Job Duration Trends (shown when a job is selected) -->
        <template v-if="selectedJob">
          <hr class="section-divider">

          <div class="section-header">
            <h2>Duration Trends: {{ selectedJob.job_name }}</h2>
            <button class="btn btn-small" @click="clearSelectedJob">Close</button>
          </div>
          <p class="table-title">Execution duration over time for this job</p>

          <div class="card">
            <div class="card-body">
              <JobDurationChart
                :trends="selectedJobTrends?.trends || []"
                :jobName="selectedJob.job_name"
              />
            </div>
          </div>

          <div class="duration-summary" v-if="selectedJob.total_runs > 0">
            <div class="summary-item">
              <span class="summary-label">Total Runs:</span>
              <span class="summary-value">{{ selectedJob.total_runs }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">Success Rate:</span>
              <span class="summary-value">{{ formatSuccessRate(selectedJob.success_rate) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">Avg Duration:</span>
              <span class="summary-value">{{ formatDuration(selectedJob.avg_duration_ms) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">Min Duration:</span>
              <span class="summary-value">{{ formatDuration(selectedJob.min_duration_ms) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">Max Duration:</span>
              <span class="summary-value">{{ formatDuration(selectedJob.max_duration_ms) }}</span>
            </div>
          </div>
        </template>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">Analytics Overview</div>
        <div class="sidebar-box-content">
          <p class="text-small">
            <strong>Execution Trends:</strong> Shows success rate or run counts over time.
          </p>
          <p class="text-small">
            <strong>Job Statistics:</strong> Per-job performance metrics including success rate and duration.
          </p>
          <p class="text-small mb-0">
            <strong>Duration Trends:</strong> Click "Trends" on any job to see how its execution time changes over time.
          </p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">Quick Stats</div>
        <div class="sidebar-box-content">
          <p class="text-small">
            <strong>Active Jobs:</strong> {{ overview?.active_jobs || 0 }}
          </p>
          <p class="text-small">
            <strong>Last 7 Days:</strong> {{ overview?.runs_last_7d || 0 }} runs
          </p>
          <p class="text-small mb-0">
            <strong>Failures:</strong> {{ overview?.failure_count || 0 }}
          </p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">Navigation</div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <router-link to="/" class="btn" style="width: 100%; text-align: center;">
              Dashboard
            </router-link>
          </p>
          <p class="mb-10">
            <router-link to="/jobs" class="btn" style="width: 100%; text-align: center;">
              Manage Jobs
            </router-link>
          </p>
          <p class="mb-0">
            <router-link to="/runs" class="btn" style="width: 100%; text-align: center;">
              Run History
            </router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin-bottom: 20px;
}

.stat-card {
  background: #fff;
  border: 1px solid #ccc;
  padding: 15px;
  text-align: center;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  line-height: 1.2;
}

.stat-value.success {
  color: #4CAF50;
}

.stat-label {
  font-size: 12px;
  color: #666;
  margin-top: 5px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.section-header h2 {
  margin: 0;
}

.section-controls {
  display: flex;
  gap: 15px;
  align-items: center;
}

.toggle-group {
  display: flex;
  border: 1px solid #ccc;
}

.toggle-btn {
  padding: 4px 12px;
  font-size: 12px;
  border: none;
  background: #fff;
  cursor: pointer;
}

.toggle-btn:first-child {
  border-right: 1px solid #ccc;
}

.toggle-btn.active {
  background: #2196F3;
  color: #fff;
}

.days-selector {
  display: flex;
  gap: 5px;
}

.days-selector .btn.active {
  background: #2196F3;
  color: #fff;
  border-color: #2196F3;
}

.duration-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  margin-top: 15px;
  padding: 15px;
  background: #f9f9f9;
  border: 1px solid #ddd;
}

.summary-item {
  display: flex;
  flex-direction: column;
}

.summary-label {
  font-size: 11px;
  color: #666;
  text-transform: uppercase;
}

.summary-value {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

@media (max-width: 800px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .section-controls {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
