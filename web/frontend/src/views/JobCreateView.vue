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
  { label: 'Daily at Midnight', minutes: [0], hours: [0], days: [], months: [], weekdays: [] },
  { label: 'Weekdays at 9 AM', minutes: [0], hours: [9], days: [], months: [], weekdays: [1, 2, 3, 4, 5] },
  { label: 'Weekly on Monday', minutes: [0], hours: [0], days: [], months: [], weekdays: [1] },
  { label: 'Monthly on 1st', minutes: [0], hours: [0], days: [1], months: [], weekdays: [] },
]

const selectedPreset = ref('Daily at 9 AM')

function applyPreset(preset) {
  selectedPreset.value = preset.label
  scheduleMinutes.value = [...preset.minutes]
  scheduleHours.value = [...preset.hours]
  scheduleDays.value = [...preset.days]
  scheduleMonths.value = [...preset.months]
  scheduleWeekdays.value = [...preset.weekdays]
}

function toggleCustomSchedule() {
  showCustomSchedule.value = !showCustomSchedule.value
  if (showCustomSchedule.value) {
    selectedPreset.value = 'Custom'
  }
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
  selectedPreset.value = 'Custom'
}

function toggleHour(value) {
  const index = scheduleHours.value.indexOf(value)
  if (index === -1) {
    scheduleHours.value.push(value)
    scheduleHours.value.sort((a, b) => a - b)
  } else {
    scheduleHours.value.splice(index, 1)
  }
  selectedPreset.value = 'Custom'
}

function toggleWeekday(value) {
  const index = scheduleWeekdays.value.indexOf(value)
  if (index === -1) {
    scheduleWeekdays.value.push(value)
    scheduleWeekdays.value.sort((a, b) => a - b)
  } else {
    scheduleWeekdays.value.splice(index, 1)
  }
  selectedPreset.value = 'Custom'
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
  <div class="main-container">
    <div class="content-area">
      <h1>Create New Job</h1>
      <p>Configure a new scheduled task with script, execution settings, and schedule.</p>

      <form @submit.prevent="handleSubmit">
        <!-- Basic Information -->
        <div class="card">
          <div class="card-header">Basic Information</div>
          <div class="card-body">
            <div class="form-group">
              <label for="name">Job Name <span class="required">*</span></label>
              <input
                id="name"
                v-model="name"
                type="text"
                placeholder="e.g., Daily Backup"
                :disabled="loading"
              />
              <div v-if="errors.name" class="field-error">{{ errors.name }}</div>
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

            <div class="form-row-2col">
              <div class="form-group">
                <label for="timezone">Timezone</label>
                <select id="timezone" v-model="timezone" :disabled="loading">
                  <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
                </select>
              </div>

              <div class="form-group">
                <label>&nbsp;</label>
                <label class="checkbox-label">
                  <input type="checkbox" v-model="enabled" :disabled="loading" />
                  Enabled
                </label>
                <div class="field-hint">Disabled jobs won't be scheduled to run</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Script -->
        <div class="card">
          <div class="card-header">Script</div>
          <div class="card-body">
            <div class="form-group">
              <label for="script">Bash Script <span class="required">*</span></label>
              <textarea
                id="script"
                v-model="script"
                class="script-editor"
                rows="10"
                spellcheck="false"
                :disabled="loading"
              ></textarea>
              <div v-if="errors.script" class="field-error">{{ errors.script }}</div>
              <div class="field-hint">The script will be executed using bash</div>
            </div>

            <div class="form-group">
              <label for="workingDir">Working Directory</label>
              <input
                id="workingDir"
                v-model="workingDir"
                type="text"
                placeholder="/tmp"
                :disabled="loading"
              />
              <div class="field-hint">Directory where the script will be executed (default: /tmp)</div>
            </div>
          </div>
        </div>

        <!-- Execution Settings -->
        <div class="card">
          <div class="card-header">Execution Settings</div>
          <div class="card-body">
            <div class="form-row-3col">
              <div class="form-group">
                <label for="timeout">Timeout (seconds)</label>
                <input
                  id="timeout"
                  v-model.number="timeout"
                  type="number"
                  min="1"
                  max="86400"
                  :disabled="loading"
                />
                <div v-if="errors.timeout" class="field-error">{{ errors.timeout }}</div>
                <div class="field-hint">Max: 86400 (24 hours)</div>
              </div>

              <div class="form-group">
                <label for="retryCount">Retry Count</label>
                <input
                  id="retryCount"
                  v-model.number="retryCount"
                  type="number"
                  min="0"
                  max="10"
                  :disabled="loading"
                />
                <div v-if="errors.retryCount" class="field-error">{{ errors.retryCount }}</div>
                <div class="field-hint">0-10 retries on failure</div>
              </div>

              <div class="form-group">
                <label for="retryDelay">Retry Delay (sec)</label>
                <input
                  id="retryDelay"
                  v-model.number="retryDelay"
                  type="number"
                  min="0"
                  max="3600"
                  :disabled="loading"
                />
                <div v-if="errors.retryDelay" class="field-error">{{ errors.retryDelay }}</div>
                <div class="field-hint">Delay between retries</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Schedule -->
        <div class="card">
          <div class="card-header">Schedule</div>
          <div class="card-body">
            <div class="form-group">
              <label>Quick Presets</label>
              <div class="preset-buttons">
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
                <button
                  type="button"
                  @click="toggleCustomSchedule"
                  :class="['preset-btn', { active: showCustomSchedule }]"
                  :disabled="loading"
                >
                  Custom...
                </button>
              </div>
            </div>

            <div v-if="showCustomSchedule" class="custom-schedule">
              <hr class="section-divider">

              <div class="form-group">
                <label>Minutes (0-59)</label>
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
                <div class="field-hint">Click to select. Empty = every minute.</div>
              </div>

              <div class="form-group">
                <label>Hours (0-23)</label>
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
                <div class="field-hint">Click to select. Empty = every hour.</div>
              </div>

              <div class="form-group">
                <label>Days of Week</label>
                <div class="weekday-buttons">
                  <button
                    v-for="d in weekdayOptions"
                    :key="d.value"
                    type="button"
                    @click="toggleWeekday(d.value)"
                    :class="['weekday-btn', { active: scheduleWeekdays.includes(d.value) }]"
                    :disabled="loading"
                  >
                    {{ d.label }}
                  </button>
                </div>
                <div class="field-hint">Click to select. Empty = any day.</div>
              </div>
            </div>

            <div class="cron-preview">
              <strong>Cron Expression:</strong>
              <code>{{ cronPreview }}</code>
            </div>
          </div>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <!-- Form Actions -->
        <div class="form-actions">
          <button type="button" @click="cancel" class="btn" :disabled="loading">
            Cancel
          </button>
          <button type="submit" class="btn btn-primary" :disabled="loading">
            <span v-if="loading">Creating...</span>
            <span v-else>Create Job</span>
          </button>
        </div>
      </form>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">Help</div>
        <div class="sidebar-box-content">
          <p class="text-small"><strong>Job Name:</strong> A unique identifier for your task.</p>
          <p class="text-small"><strong>Script:</strong> Bash script to execute. Must start with shebang.</p>
          <p class="text-small"><strong>Timeout:</strong> Maximum execution time before termination.</p>
          <p class="text-small mb-0"><strong>Schedule:</strong> When the job should run (cron-style).</p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">Cron Format</div>
        <div class="sidebar-box-content">
          <pre class="cron-format">minute hour day month weekday
  *     *    *    *      *
  |     |    |    |      |
  |     |    |    |      +-- Day of week (0-6)
  |     |    |    +--------- Month (1-12)
  |     |    +-------------- Day of month (1-31)
  |     +------------------- Hour (0-23)
  +------------------------- Minute (0-59)</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.required {
  color: #cc0000;
}

.field-error {
  color: #cc0000;
  font-size: 11px;
  margin-top: 4px;
}

.field-hint {
  color: #666666;
  font-size: 11px;
  margin-top: 4px;
}

.form-row-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
}

.form-row-3col {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 15px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: bold;
  padding: 6px 0;
}

.checkbox-label input {
  width: auto;
}

.script-editor {
  font-family: "Courier New", monospace;
  font-size: 12px;
  line-height: 1.4;
  resize: vertical;
}

.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.preset-btn {
  padding: 6px 12px;
  font-size: 11px;
  border: 1px solid #999999;
  background: #ffffff;
  cursor: pointer;
  color: #333333;
  font-weight: bold;
}

.preset-btn:hover {
  background: #bcd4ec;
}

.preset-btn.active {
  background: #6699cc;
  color: #ffffff;
  border-color: #6699cc;
}

.preset-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.custom-schedule {
  margin-top: 15px;
}

.value-grid {
  display: grid;
  gap: 3px;
}

.minute-grid {
  grid-template-columns: repeat(15, 1fr);
}

.hour-grid {
  grid-template-columns: repeat(12, 1fr);
}

.value-btn {
  padding: 4px 2px;
  font-size: 10px;
  border: 1px solid #cccccc;
  background: #ffffff;
  cursor: pointer;
  color: #333333;
  font-weight: bold;
  text-align: center;
}

.value-btn:hover {
  background: #bcd4ec;
}

.value-btn.active {
  background: #6699cc;
  color: #ffffff;
  border-color: #6699cc;
}

.value-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.weekday-buttons {
  display: flex;
  gap: 6px;
}

.weekday-btn {
  padding: 6px 12px;
  font-size: 11px;
  border: 1px solid #cccccc;
  background: #ffffff;
  cursor: pointer;
  color: #333333;
  font-weight: bold;
}

.weekday-btn:hover {
  background: #bcd4ec;
}

.weekday-btn.active {
  background: #6699cc;
  color: #ffffff;
  border-color: #6699cc;
}

.weekday-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cron-preview {
  margin-top: 15px;
  padding: 10px;
  background: #f4f4f4;
  border: 1px solid #cccccc;
}

.cron-preview code {
  font-family: "Courier New", monospace;
  background: #ffffff;
  padding: 2px 8px;
  border: 1px solid #cccccc;
  margin-left: 10px;
}

.cron-format {
  font-family: "Courier New", monospace;
  font-size: 10px;
  line-height: 1.4;
  margin: 0;
  white-space: pre;
  overflow-x: auto;
}

.form-actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .form-row-2col,
  .form-row-3col {
    grid-template-columns: 1fr;
  }

  .minute-grid {
    grid-template-columns: repeat(10, 1fr);
  }

  .hour-grid {
    grid-template-columns: repeat(8, 1fr);
  }

  .weekday-buttons {
    flex-wrap: wrap;
  }
}
</style>
