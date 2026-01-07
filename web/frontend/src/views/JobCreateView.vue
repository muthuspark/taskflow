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
  <div class="job-create">
    <div class="page-header">
      <h1>Create New Job</h1>
    </div>

    <form @submit.prevent="handleSubmit" class="job-form">
      <div class="form-card">
        <h2>Basic Information</h2>

        <div class="form-group">
          <label for="name">Job Name *</label>
          <input
            id="name"
            v-model="name"
            type="text"
            placeholder="e.g., Daily Backup"
            :class="{ 'has-error': errors.name }"
            :disabled="loading"
          />
          <span v-if="errors.name" class="field-error">{{ errors.name }}</span>
        </div>

        <div class="form-group">
          <label for="description">Description</label>
          <textarea
            id="description"
            v-model="description"
            placeholder="Describe what this job does..."
            rows="2"
            :disabled="loading"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="timezone">Timezone</label>
          <select id="timezone" v-model="timezone" :disabled="loading">
            <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
          </select>
        </div>

        <div class="form-group checkbox-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="enabled" :disabled="loading" />
            <span>Enabled</span>
          </label>
          <span class="help-text">Disabled jobs won't be scheduled to run</span>
        </div>
      </div>

      <div class="form-card">
        <h2>Script</h2>

        <div class="form-group">
          <label for="script">Bash Script *</label>
          <textarea
            id="script"
            v-model="script"
            class="script-editor"
            rows="12"
            spellcheck="false"
            :class="{ 'has-error': errors.script }"
            :disabled="loading"
          ></textarea>
          <span v-if="errors.script" class="field-error">{{ errors.script }}</span>
          <span class="help-text">The script will be executed using bash</span>
        </div>
      </div>

      <div class="form-card">
        <h2>Execution Settings</h2>

        <div class="form-row">
          <div class="form-group">
            <label for="timeout">Timeout (seconds)</label>
            <input
              id="timeout"
              v-model.number="timeout"
              type="number"
              min="1"
              max="86400"
              :class="{ 'has-error': errors.timeout }"
              :disabled="loading"
            />
            <span v-if="errors.timeout" class="field-error">{{ errors.timeout }}</span>
            <span class="help-text">Max: 86400 (24 hours)</span>
          </div>

          <div class="form-group">
            <label for="retryCount">Retry Count</label>
            <input
              id="retryCount"
              v-model.number="retryCount"
              type="number"
              min="0"
              max="10"
              :class="{ 'has-error': errors.retryCount }"
              :disabled="loading"
            />
            <span v-if="errors.retryCount" class="field-error">{{ errors.retryCount }}</span>
            <span class="help-text">Number of retries on failure</span>
          </div>

          <div class="form-group">
            <label for="retryDelay">Retry Delay (seconds)</label>
            <input
              id="retryDelay"
              v-model.number="retryDelay"
              type="number"
              min="0"
              max="3600"
              :class="{ 'has-error': errors.retryDelay }"
              :disabled="loading"
            />
            <span v-if="errors.retryDelay" class="field-error">{{ errors.retryDelay }}</span>
            <span class="help-text">Delay between retries</span>
          </div>
        </div>
      </div>

      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <div class="form-actions">
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
.job-create {
  max-width: 800px;
}

.page-header h1 {
  margin: 0 0 1.5rem 0;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.job-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-card {
  background: var(--white);
  border: 1px solid var(--gray-light);
  padding: 1.5rem;
  box-shadow: none;
}

.form-card h2 {
  margin: 0 0 1rem 0;
  font-size: 1.125rem;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--gray-light);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-weight: 900;
  color: var(--black);
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 0;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-size: 1rem;
  transition: none;
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
  outline: none;
  border: 1px solid var(--gray-light);
  box-shadow: none;
}

.form-group input.has-error,
.form-group textarea.has-error {
  border: 1px solid var(--gray-light);
}

.form-group input:disabled,
.form-group textarea:disabled,
.form-group select:disabled {
  background: var(--gray-lighter);
  cursor: not-allowed;
}

.script-editor {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  resize: vertical;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: 700;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
}

.help-text {
  font-size: 0.75rem;
  color: var(--gray-dark);
  margin-top: 0.25rem;
}

.field-error {
  display: block;
  font-size: 0.75rem;
  color: var(--black);
  margin-top: 0.25rem;
  font-weight: 900;
}

.error-message {
  background: var(--gray-lighter);
  color: var(--black);
  padding: 1rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  text-align: center;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

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
