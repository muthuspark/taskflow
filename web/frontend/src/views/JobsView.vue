<script setup>
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useJobsStore } from '../stores/jobs'
import JobCard from '../components/JobCard.vue'

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
</script>

<template>
  <div class="jobs-view">
    <div class="page-header">
      <h1>Jobs</h1>
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

    <div v-else class="jobs-grid">
      <JobCard
        v-for="job in jobs"
        :key="job.id"
        :job="job"
        @click="goToJob(job.id)"
        @run="handleRun(job.id)"
        @delete="handleDelete(job.id)"
      />
    </div>
  </div>
</template>

<style scoped>
.jobs-view {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  margin: 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.loading-container {
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
  text-align: center;
  padding: 3rem;
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
  color: var(--black);
  font-weight: 900;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: var(--white);
  border-radius: 0;
  border: 1px solid var(--gray-light);
}

.empty-state h2 {
  margin: 0 0 0.5rem 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.empty-state p {
  margin: 0 0 1.5rem 0;
  color: var(--gray-medium);
}

.jobs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}
</style>
