<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useJobsStore } from '../stores/jobs'

const router = useRouter()
const jobsStore = useJobsStore()

// Form state
const name = ref('')
const description = ref('')
const script = ref('#!/bin/bash\n\n# Your script here\necho "Hello, TaskFlow!"')
const workingDir = ref('')
const timeout = ref(300)
const retryCount = ref(0)
const retryDelay = ref(60)
const enabled = ref(true)
const timezone = ref('UTC')

// Schedule state
const scheduleMinutes = ref([0])
const scheduleHours = ref([9])
const scheduleDays = ref([])
const scheduleMonths = ref([])
const scheduleWeekdays = ref([])
const showCustomSchedule = ref(false)

const loading = ref(false)
const error = ref('')
const errors = ref({})

// Schedule presets
const schedulePresets = [
  { label: 'Daily at 9 AM', minutes: [0], hours: [9], days: [], months: [], weekdays: [] },
  { label: 'Hourly', minutes: [0], hours: [], days: [], months: [], weekdays: [] },
  { label: 'Daily at midnight', minutes: [0], hours: [0], days: [], months: [], weekdays: [] },
  { label: 'Weekdays at 9 AM', minutes: [0], hours: [9], days: [], months: [], weekdays: [1, 2, 3, 4, 5] },
  { label: 'Weekly on Monday', minutes: [0], hours: [0], days: [], months: [], weekdays: [1] },
  { label: 'Monthly on 1st', minutes: [0], hours: [0], days: [1], months: [], weekdays: [] },
  { label: 'Custom...', custom: true }
]

const selectedPreset = ref('Daily at 9 AM')

function applyPreset(preset) {
  if (preset.custom) {
    showCustomSchedule.value = true
    selectedPreset.value = 'Custom...'
    return
  }
  showCustomSchedule.value = false
  selectedPreset.value = preset.label
  scheduleMinutes.value = [...preset.minutes]
  scheduleHours.value = [...preset.hours]
  scheduleDays.value = [...preset.days]
  scheduleMonths.value = [...preset.months]
  scheduleWeekdays.value = [...preset.weekdays]
}

// Cron preview
const cronPreview = computed(() => {
  const min = scheduleMinutes.value.length ? scheduleMinutes.value.join(',') : '*'
  const hour = scheduleHours.value.length ? scheduleHours.value.join(',') : '*'
  const dom = scheduleDays.value.length ? scheduleDays.value.join(',') : '*'
  const month = scheduleMonths.value.length ? scheduleMonths.value.join(',') : '*'
  const dow = scheduleWeekdays.value.length ? scheduleWeekdays.value.join(',') : '*'
  return `${min} ${hour} ${dom} ${month} ${dow}`
})

// Custom schedule helpers
const minuteOptions = Array.from({ length: 60 }, (_, i) => i)
const hourOptions = Array.from({ length: 24 }, (_, i) => i)
const weekdayOptions = [
  { value: 0, label: 'Sun' },
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' }
]

function toggleMinute(value) {
  const index = scheduleMinutes.value.indexOf(value)
  if (index === -1) {
    scheduleMinutes.value.push(value)
    scheduleMinutes.value.sort((a, b) => a - b)
  } else {
    scheduleMinutes.value.splice(index, 1)
  }
}

function toggleHour(value) {
  const index = scheduleHours.value.indexOf(value)
  if (index === -1) {
    scheduleHours.value.push(value)
    scheduleHours.value.sort((a, b) => a - b)
  } else {
    scheduleHours.value.splice(index, 1)
  }
}

function toggleWeekday(value) {
  const index = scheduleWeekdays.value.indexOf(value)
  if (index === -1) {
    scheduleWeekdays.value.push(value)
    scheduleWeekdays.value.sort((a, b) => a - b)
  } else {
    scheduleWeekdays.value.splice(index, 1)
  }
}

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
      working_dir: workingDir.value.trim() || '/tmp',
      timeout_seconds: parseInt(timeout.value),
      retry_count: parseInt(retryCount.value),
      retry_delay_seconds: parseInt(retryDelay.value),
      enabled: enabled.value,
      timezone: timezone.value,
      schedule: {
        minutes: scheduleMinutes.value.length ? scheduleMinutes.value : null,
        hours: scheduleHours.value.length ? scheduleHours.value : null,
        days: scheduleDays.value.length ? scheduleDays.value : null,
        months: scheduleMonths.value.length ? scheduleMonths.value : null,
        weekdays: scheduleWeekdays.value.length ? scheduleWeekdays.value : null
      }
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

        <div class="form-group">
          <label for="workingDir" class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Working Directory</label>
          <input
            id="workingDir"
            v-model="workingDir"
            type="text"
            placeholder="/tmp"
            :disabled="loading"
            class="w-full px-3 py-2 border border-gray-light focus:border-gray-light focus:outline-none disabled:bg-gray-lighter disabled:cursor-not-allowed"
          />
          <span class="text-xs text-gray-dark mt-1 block">Directory where the script will be executed (default: /tmp)</span>
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

      <div class="bg-white border border-gray-light p-6">
        <h2 class="m-0 mb-4 pb-3 border-b border-gray-light text-black font-black uppercase tracking-tight text-lg">Schedule</h2>

        <div class="form-group mb-4">
          <label class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Quick Presets</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="preset in schedulePresets"
              :key="preset.label"
              type="button"
              @click="applyPreset(preset)"
              :class="['preset-btn', { active: selectedPreset === preset.label }]"
              :disabled="loading"
            >
              {{ preset.label }}
            </button>
          </div>
        </div>

        <div v-if="showCustomSchedule" class="custom-schedule">
          <div class="form-group mb-4">
            <label class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Minutes</label>
            <div class="value-grid minute-grid">
              <button
                v-for="m in minuteOptions"
                :key="m"
                type="button"
                @click="toggleMinute(m)"
                :class="['value-btn', { active: scheduleMinutes.includes(m) }]"
                :disabled="loading"
              >
                {{ m }}
              </button>
            </div>
            <span class="text-xs text-gray-dark mt-1 block">Empty = every minute</span>
          </div>

          <div class="form-group mb-4">
            <label class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Hours</label>
            <div class="value-grid hour-grid">
              <button
                v-for="h in hourOptions"
                :key="h"
                type="button"
                @click="toggleHour(h)"
                :class="['value-btn', { active: scheduleHours.includes(h) }]"
                :disabled="loading"
              >
                {{ h }}
              </button>
            </div>
            <span class="text-xs text-gray-dark mt-1 block">Empty = every hour</span>
          </div>

          <div class="form-group mb-4">
            <label class="block font-black text-black mb-2 text-sm uppercase tracking-tight">Days of Week</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="d in weekdayOptions"
                :key="d.value"
                type="button"
                @click="toggleWeekday(d.value)"
                :class="['value-btn weekday-btn', { active: scheduleWeekdays.includes(d.value) }]"
                :disabled="loading"
              >
                {{ d.label }}
              </button>
            </div>
            <span class="text-xs text-gray-dark mt-1 block">Empty = any day</span>
          </div>
        </div>

        <div class="cron-preview">
          <label class="font-black text-black text-sm uppercase tracking-tight">Cron Expression:</label>
          <code class="ml-2 font-mono text-sm bg-gray-lighter px-2 py-1 border border-gray-light">{{ cronPreview }}</code>
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

/* Schedule preset buttons */
.preset-btn {
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  background: var(--white);
  cursor: pointer;
  color: var(--black);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preset-btn:hover {
  background: var(--gray-lighter);
}

.preset-btn.active {
  background: var(--black);
  color: var(--white);
}

.preset-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Schedule value grids */
.value-grid {
  display: grid;
  gap: 4px;
}

.minute-grid {
  grid-template-columns: repeat(12, 1fr);
}

.hour-grid {
  grid-template-columns: repeat(12, 1fr);
}

.value-btn {
  padding: 0.375rem 0.25rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  background: var(--white);
  cursor: pointer;
  color: var(--black);
  font-weight: 700;
}

.value-btn:hover {
  background: var(--gray-lighter);
}

.value-btn.active {
  background: var(--black);
  color: var(--white);
}

.value-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.weekday-btn {
  padding: 0.5rem 0.75rem;
}

.cron-preview {
  display: flex;
  align-items: center;
  padding: 1rem;
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
  margin-top: 1rem;
}

.custom-schedule {
  border-top: 1px solid var(--gray-light);
  padding-top: 1rem;
  margin-top: 1rem;
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
