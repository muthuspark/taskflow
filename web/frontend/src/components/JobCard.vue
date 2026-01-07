<script setup>
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  job: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['click', 'run', 'delete'])

const statusText = computed(() => props.job.enabled ? 'enabled' : 'disabled')

function handleClick(e) {
  // Don't trigger click when clicking buttons
  if (e.target.closest('button')) return
  emit('click')
}

function handleRun(e) {
  e.stopPropagation()
  emit('run')
}

function handleDelete(e) {
  e.stopPropagation()
  emit('delete')
}

function formatDate(dateStr) {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleDateString()
}
</script>

<template>
  <div class="job-card" @click="handleClick">
    <div class="job-header">
      <h3 class="job-name">{{ job.name }}</h3>
      <StatusBadge :status="statusText" />
    </div>

    <p v-if="job.description" class="job-description">
      {{ job.description }}
    </p>
    <p v-else class="job-description empty">No description</p>

    <div class="job-meta">
      <div class="meta-item">
        <span class="meta-label">Timeout:</span>
        <span class="meta-value">{{ job.timeout_seconds }}s</span>
      </div>
      <div class="meta-item">
        <span class="meta-label">Retries:</span>
        <span class="meta-value">{{ job.retry_count }}</span>
      </div>
      <div class="meta-item">
        <span class="meta-label">Created:</span>
        <span class="meta-value">{{ formatDate(job.created_at) }}</span>
      </div>
    </div>

    <div class="job-actions">
      <button @click="handleRun" class="btn btn-primary btn-small">
        Run Now
      </button>
      <button @click="handleDelete" class="btn btn-danger btn-small">
        Delete
      </button>
    </div>
  </div>
</template>

<style scoped>
/* JobCard inherits card styling from global CSS */
.job-card {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.job-card:hover {
  background: var(--gray-lighter);
}

.job-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--gray-light);
}

.job-name {
  margin: 0;
  font-size: 1.125rem;
  color: var(--black);
  word-break: break-word;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.job-description {
  margin: 0;
  font-size: 0.875rem;
  color: var(--gray-medium);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.job-description.empty {
  font-style: italic;
  color: var(--gray-medium);
}

.job-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--black);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.meta-item {
  display: flex;
  gap: 0.25rem;
}

.meta-label {
  color: var(--gray-dark);
  font-weight: 900;
}

.meta-value {
  color: var(--black);
  font-weight: 900;
}

.job-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: auto;
  padding-top: 0.75rem;
}
</style>
