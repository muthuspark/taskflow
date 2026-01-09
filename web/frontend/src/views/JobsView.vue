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
  <div class="main-container">
    <div class="content-area">
      <div class="page-header">
        <h1 style="margin: 0;">Jobs</h1>
        <button @click="goToCreate" class="btn btn-primary">
          Create New Job
        </button>
      </div>

      <p>Manage your scheduled tasks. Each job can be configured with a script, schedule, timeout, and retry settings.</p>

      <div v-if="loading && !jobs.length" class="loading-container">
        <div class="spinner"></div>
        <p>Loading jobs...</p>
      </div>

      <div v-else-if="error" class="error-message">
        {{ error }}
        <button @click="retry" class="btn btn-small" style="margin-left: 10px;">Retry</button>
      </div>

      <div v-else-if="!jobs.length" class="empty-state">
        <h2>No jobs yet</h2>
        <p>Create your first job to get started with task scheduling.</p>
        <button @click="goToCreate" class="btn btn-primary">
          Create Your First Job
        </button>
      </div>

      <template v-else>
        <h2>All Jobs</h2>
        <p class="table-title">List of all configured jobs</p>

        <table class="full-width">
          <thead>
            <tr>
              <th>#</th>
              <th>Job Name</th>
              <th>Description</th>
              <th>Status</th>
              <th>Timeout</th>
              <th>Retries</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(job, index) in jobs" :key="job.id">
              <td>{{ index + 1 }}.</td>
              <td>
                <a href="#" @click.prevent="goToJob(job.id)">{{ job.name }}</a>
              </td>
              <td>
                <span v-if="job.description">{{ job.description }}</span>
                <span v-else class="text-muted" style="font-style: italic;">No description</span>
              </td>
              <td>
                <StatusBadge :status="getStatusText(job)" />
              </td>
              <td class="number">{{ job.timeout_seconds }}s</td>
              <td class="number">{{ job.retry_count }}</td>
              <td>{{ formatDate(job.created_at) }}</td>
              <td>
                <button @click="handleRun(job.id)" class="btn btn-small btn-primary" style="margin-right: 4px;">
                  Run
                </button>
                <button @click="handleDelete(job.id)" class="btn btn-small btn-danger">
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="table-note">showing all {{ jobs.length }} job(s)</div>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Quick Actions
        </div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <button @click="goToCreate" class="btn btn-primary" style="width: 100%;">
              Create New Job
            </button>
          </p>
          <p class="mb-0">
            <router-link to="/runs" class="btn" style="width: 100%; text-align: center;">
              View Run History
            </router-link>
          </p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">
          Job Settings
        </div>
        <div class="sidebar-box-content">
          <p class="text-small">
            <strong>Timeout:</strong> Maximum execution time before the job is terminated.
          </p>
          <p class="text-small">
            <strong>Retries:</strong> Number of retry attempts if the job fails.
          </p>
          <p class="text-small mb-0">
            <strong>Status:</strong> Enable or disable job scheduling.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Using global W3Techs-style CSS */
</style>
