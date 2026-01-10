<script setup>
import { computed } from 'vue'

const props = defineProps({
  jobs: { type: Array, default: () => [] }
})

const emit = defineEmits(['select-job'])

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

function getSuccessRateClass(rate) {
  if (rate === undefined || rate === null) return ''
  if (rate >= 0.9) return 'rate-good'
  if (rate >= 0.7) return 'rate-warning'
  return 'rate-bad'
}

function formatDate(dateStr) {
  if (!dateStr) return 'Never'
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const sortedJobs = computed(() => {
  return [...props.jobs].sort((a, b) => b.total_runs - a.total_runs)
})
</script>

<template>
  <div class="job-stats-table">
    <div v-if="!jobs || jobs.length === 0" class="no-data">
      <p>No job statistics available</p>
    </div>
    <table v-else>
      <thead>
        <tr>
          <th>Job Name</th>
          <th class="text-right">Runs</th>
          <th class="text-right">Success Rate</th>
          <th class="text-right">Avg Duration</th>
          <th class="text-right">Min / Max</th>
          <th>Last Run</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="job in sortedJobs" :key="job.job_id">
          <td class="job-name">
            <router-link :to="`/jobs/${job.job_id}`">{{ job.job_name }}</router-link>
          </td>
          <td class="text-right">{{ job.total_runs }}</td>
          <td class="text-right">
            <span :class="getSuccessRateClass(job.success_rate)">
              {{ formatSuccessRate(job.success_rate) }}
            </span>
          </td>
          <td class="text-right">{{ formatDuration(job.avg_duration_ms) }}</td>
          <td class="text-right text-muted">
            {{ formatDuration(job.min_duration_ms) }} / {{ formatDuration(job.max_duration_ms) }}
          </td>
          <td class="text-muted">{{ formatDate(job.last_run_at) }}</td>
          <td>
            <button
              v-if="job.total_runs > 0"
              class="btn btn-small"
              @click="emit('select-job', job)"
            >
              Trends
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.job-stats-table {
  overflow-x: auto;
}

.job-stats-table table {
  width: 100%;
  margin: 0;
}

.job-stats-table th {
  background: #f4f4f4;
  font-weight: bold;
  text-align: left;
  white-space: nowrap;
}

.job-stats-table td {
  white-space: nowrap;
}

.job-name {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.job-name a {
  color: #2196F3;
  text-decoration: none;
}

.job-name a:hover {
  text-decoration: underline;
}

.text-right {
  text-align: right;
}

.text-muted {
  color: #666;
  font-size: 12px;
}

.rate-good {
  color: #4CAF50;
  font-weight: bold;
}

.rate-warning {
  color: #FF9800;
  font-weight: bold;
}

.rate-bad {
  color: #f44336;
  font-weight: bold;
}

.btn-small {
  padding: 2px 8px;
  font-size: 11px;
}

.no-data {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100px;
  color: #666;
}

.no-data p {
  margin: 0;
}
</style>
