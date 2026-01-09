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
      return 'success'
    case 'failure':
    case 'failed':
    case 'error':
      return 'failed'
    case 'running':
    case 'in_progress':
      return 'running'
    case 'pending':
    case 'queued':
      return 'pending'
    case 'timeout':
    case 'timed_out':
      return 'failed'
    case 'cancelled':
    case 'canceled':
    case 'disabled':
      return 'pending'
    default:
      return 'pending'
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
/* Using global W3Techs-style CSS */
</style>
