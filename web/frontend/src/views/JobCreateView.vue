<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useJobsStore } from '../stores/jobs'

const router = useRouter()
const jobsStore = useJobsStore()

// Form state
const name = ref('')
const description = ref('')
const script = ref('#!/bin/bash\n\n# Your script here\necho "Hello, TaskFlow!"')
const timeout = ref(300)
const retryCount = ref(0)
const retryDelay = ref(60)
const enabled = ref(true)
const timezone = ref('UTC')

const loading = ref(false)
const error = ref('')
const errors = ref({})

// Validation
function validate() {
  errors.value = {}

  if (!name.value.trim()) {
    errors.value.name = 'Name is required'
  } else if (name.value.length > 100) {
    errors.value.name = 'Name must be less than 100 characters'
  }

  if (!script.value.trim()) {
    errors.value.script = 'Script is required'
  }

  if (timeout.value < 1 || timeout.value > 86400) {
    errors.value.timeout = 'Timeout must be between 1 and 86400 seconds'
  }

  if (retryCount.value < 0 || retryCount.value > 10) {
    errors.value.retryCount = 'Retry count must be between 0 and 10'
  }

  if (retryDelay.value < 0 || retryDelay.value > 3600) {
    errors.value.retryDelay = 'Retry delay must be between 0 and 3600 seconds'
  }

  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return

  error.value = ''
  loading.value = true

  try {
    const job = await jobsStore.createJob({
      name: name.value.trim(),
      description: description.value.trim(),
      script: script.value,
      timeout_seconds: parseInt(timeout.value),
      retry_count: parseInt(retryCount.value),
      retry_delay_seconds: parseInt(retryDelay.value),
      enabled: enabled.value,
      timezone: timezone.value
    })
    router.push(`/jobs/${job.id}`)
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to create job'
  } finally {
    loading.value = false
  }
}

function cancel() {
  router.back()
}

// Common timezones
const timezones = [
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Asia/Tokyo',
  'Asia/Shanghai',
  'Asia/Kolkata',
  'Australia/Sydney'
]
</script>

<template>
  <div class="max-w-[800px]">
    <h1 class="m-0 mb-6 text-black font-black uppercase tracking-tight">Create New Job</h1>

    <form @submit.prevent="handleSubmit" class="flex flex-col gap-6">
      <div class="bg-white border border-gray-light p-6">
        <h2 class="m-0 mb-4 pb-3 border-b border-gray-light text-black font-black uppercase tracking-tight text-lg">Basic Information</h2>

        <div class="form-group">
          <label for="name" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Job Name *</label>
          <input
            id="name"
            v-model="name"
            type="text"
            placeholder="e.g., Daily Backup"
            :class="{ 'has-error': errors.name }"
            :disabled="loading"
            class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
          />
          <span v-if="errors.name" class="text-xs font-black mt-1 block">{{ errors.name }}</span>
        </div>

        <div class="form-group">
          <label for="description" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Description</label>
          <textarea
            id="description"
            v-model="description"
            placeholder="Describe what this job does..."
            rows="2"
            :disabled="loading"
            class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="timezone" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Timezone</label>
          <select id="timezone" v-model="timezone" :disabled="loading" class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed">
            <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
          </select>
        </div>

        <div class="flex flex-col gap-1">
          <label class="flex items-center gap-2 cursor-pointer font-bold">
            <input type="checkbox" v-model="enabled" :disabled="loading" class="w-auto" />
            <span>Enabled</span>
          </label>
          <span class="text-xs text-gray-dark">Disabled jobs won't be scheduled to run</span>
        </div>
      </div>

      <div class="bg-white border border-gray-light p-6">
        <h2 class="m-0 mb-4 pb-3 border-b border-gray-light text-black font-black uppercase tracking-tight text-lg">Script</h2>

        <div class="form-group">
          <label for="script" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Bash Script *</label>
          <textarea
            id="script"
            v-model="script"
            class="script-editor w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
            rows="12"
            spellcheck="false"
            :class="{ 'has-error': errors.script }"
            :disabled="loading"
          ></textarea>
          <span v-if="errors.script" class="text-xs font-black mt-1 block">{{ errors.script }}</span>
          <span class="text-xs text-gray-dark mt-1 block">The script will be executed using bash</span>
        </div>
      </div>

      <div class="bg-white border border-gray-light p-6">
        <h2 class="m-0 mb-4 pb-3 border-b border-gray-light text-black font-black uppercase tracking-tight text-lg">Execution Settings</h2>

        <div class="grid grid-cols-[repeat(auto-fit,minmax(150px,1fr))] gap-4">
          <div class="form-group">
            <label for="timeout" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Timeout (seconds)</label>
            <input
              id="timeout"
              v-model.number="timeout"
              type="number"
              min="1"
              max="86400"
              :class="{ 'has-error': errors.timeout }"
              :disabled="loading"
              class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
            />
            <span v-if="errors.timeout" class="text-xs font-black mt-1 block">{{ errors.timeout }}</span>
            <span class="text-xs text-gray-dark mt-1 block">Max: 86400 (24 hours)</span>
          </div>

          <div class="form-group">
            <label for="retryCount" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Retry Count</label>
            <input
              id="retryCount"
              v-model.number="retryCount"
              type="number"
              min="0"
              max="10"
              :class="{ 'has-error': errors.retryCount }"
              :disabled="loading"
              class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
            />
            <span v-if="errors.retryCount" class="text-xs font-black mt-1 block">{{ errors.retryCount }}</span>
            <span class="text-xs text-gray-dark mt-1 block">Number of retries on failure</span>
          </div>

          <div class="form-group">
            <label for="retryDelay" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Retry Delay (seconds)</label>
            <input
              id="retryDelay"
              v-model.number="retryDelay"
              type="number"
              min="0"
              max="3600"
              :class="{ 'has-error': errors.retryDelay }"
              :disabled="loading"
              class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
            />
            <span v-if="errors.retryDelay" class="text-xs font-black mt-1 block">{{ errors.retryDelay }}</span>
            <span class="text-xs text-gray-dark mt-1 block">Delay between retries</span>
          </div>
        </div>
      </div>

      <div v-if="error" class="bg-gray-lighter border border-gray-light p-4 text-center text-black">
        {{ error }}
      </div>

      <div class="flex justify-end gap-4">
        <button type="button" @click="cancel" class="btn btn-secondary" :disabled="loading">
          Cancel
        </button>
        <button type="submit" class="btn btn-primary" :disabled="loading">
          <span v-if="loading">
            <span class="spinner"></span>
            Creating...
          </span>
          <span v-else>Create Job</span>
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
/* Form validation and input styling */
.form-group input.has-error,
.form-group textarea.has-error {
  border-color: var(--gray-light);
}

/* Specialized script editor styling - monospace font */
.script-editor {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  resize: vertical;
  padding: 0.75rem !important;
}

/* Inline spinner */
.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid var(--gray-light);
  border-top-color: var(--white);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 0.5rem;
  vertical-align: middle;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
