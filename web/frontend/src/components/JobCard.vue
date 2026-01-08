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
  <div class="job-card cursor-pointer flex flex-col gap-3 hover:bg-gray-lighter transition-colors" @click="handleClick">
    <div class="flex justify-between items-start gap-3 pb-3 border-b border-gray-light">
      <h3 class="m-0 text-xl text-black font-black uppercase tracking-tight break-words">{{ job.name }}</h3>
      <StatusBadge :status="statusText" />
    </div>

    <p v-if="job.description" class="job-description m-0 text-sm text-gray-medium leading-relaxed">
      {{ job.description }}
    </p>
    <p v-else class="m-0 text-sm text-gray-medium italic">No description</p>

    <div class="flex flex-wrap gap-4 text-xs py-3 border-t border-black font-medium uppercase tracking-tight">
      <div class="flex gap-1">
        <span class="text-gray-dark font-black">Timeout:</span>
        <span class="text-black font-black">{{ job.timeout_seconds }}s</span>
      </div>
      <div class="flex gap-1">
        <span class="text-gray-dark font-black">Retries:</span>
        <span class="text-black font-black">{{ job.retry_count }}</span>
      </div>
      <div class="flex gap-1">
        <span class="text-gray-dark font-black">Created:</span>
        <span class="text-black font-black">{{ formatDate(job.created_at) }}</span>
      </div>
    </div>

    <div class="flex gap-2 mt-auto pt-3">
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
/* Line clamping for description (2 lines max) */
.job-description {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
