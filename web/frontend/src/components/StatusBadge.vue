<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: true
  }
})

const statusClass = computed(() => {
  const status = props.status.toLowerCase()
  switch (status) {
    case 'success':
    case 'completed':
    case 'enabled':
      return 'status-success'
    case 'failure':
    case 'failed':
    case 'error':
      return 'status-failure'
    case 'running':
    case 'in_progress':
      return 'status-running'
    case 'pending':
    case 'queued':
      return 'status-pending'
    case 'timeout':
    case 'timed_out':
      return 'status-timeout'
    case 'cancelled':
    case 'canceled':
    case 'disabled':
      return 'status-cancelled'
    default:
      return 'status-default'
  }
})

const displayText = computed(() => {
  return props.status.replace(/_/g, ' ')
})
</script>

<template>
  <span class="status-badge" :class="statusClass">
    {{ displayText }}
  </span>
</template>

<style scoped>
/* Status-specific overrides using global .status-badge base */
.status-failure,
.status-timeout {
  font-weight: 900;
}

.status-running {
  position: relative;
}

.status-running::before {
  content: '';
  display: inline-block;
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
  margin-right: 0.375rem;
  animation: none;
}

.status-cancelled {
  background: var(--gray-lighter);
  color: var(--black);
  border: 1px solid var(--gray-light);
}
</style>
