<script setup>
import { ref, onMounted } from 'vue'
import { useJobsStore } from '../stores/jobs'

const props = defineProps({
  job: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['save', 'cancel'])

const jobsStore = useJobsStore()

// Form state
const name = ref('')
const description = ref('')
const script = ref('')
const workingDir = ref('')
const timeout = ref(300)
const retryCount = ref(0)
const retryDelay = ref(60)
const enabled = ref(true)
const timezone = ref('UTC')

const loading = ref(false)
const error = ref('')
const errors = ref({})

// Initialize from job
onMounted(() => {
  name.value = props.job.name || ''
  description.value = props.job.description || ''
  script.value = props.job.script || ''
  workingDir.value = props.job.working_dir || ''
  timeout.value = props.job.timeout_seconds || 300
  retryCount.value = props.job.retry_count || 0
  retryDelay.value = props.job.retry_delay_seconds || 60
  enabled.value = props.job.enabled !== false
  timezone.value = props.job.timezone || 'UTC'
})

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
    await jobsStore.updateJob(props.job.id, {
      name: name.value.trim(),
      description: description.value.trim(),
      script: script.value,
      working_dir: workingDir.value.trim() || '/tmp',
      timeout_seconds: parseInt(timeout.value),
      retry_count: parseInt(retryCount.value),
      retry_delay_seconds: parseInt(retryDelay.value),
      enabled: enabled.value,
      timezone: timezone.value
    })
    emit('save')
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to update job'
  } finally {
    loading.value = false
  }
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
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Edit Job</h2>
        <button @click="emit('cancel')" class="close-btn">&times;</button>
      </div>

      <form @submit.prevent="handleSubmit" class="modal-body">
        <div class="form-group">
          <label for="edit-name">Job Name *</label>
          <input
            id="edit-name"
            v-model="name"
            type="text"
            :class="{ 'has-error': errors.name }"
            :disabled="loading"
          />
          <span v-if="errors.name" class="field-error">{{ errors.name }}</span>
        </div>

        <div class="form-group">
          <label for="edit-description">Description</label>
          <textarea
            id="edit-description"
            v-model="description"
            rows="2"
            :disabled="loading"
          ></textarea>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="edit-timezone">Timezone</label>
            <select id="edit-timezone" v-model="timezone" :disabled="loading">
              <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
            </select>
          </div>

          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="enabled" :disabled="loading" />
              <span>Enabled</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <label for="edit-script">Bash Script *</label>
          <textarea
            id="edit-script"
            v-model="script"
            class="script-editor"
            rows="10"
            spellcheck="false"
            :class="{ 'has-error': errors.script }"
            :disabled="loading"
          ></textarea>
          <span v-if="errors.script" class="field-error">{{ errors.script }}</span>
        </div>

        <div class="form-group">
          <label for="edit-workingDir">Working Directory</label>
          <input
            id="edit-workingDir"
            v-model="workingDir"
            type="text"
            placeholder="/tmp"
            :disabled="loading"
          />
          <span class="field-hint">Directory where the script will be executed (default: /tmp)</span>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="edit-timeout">Timeout (seconds)</label>
            <input
              id="edit-timeout"
              v-model.number="timeout"
              type="number"
              min="1"
              max="86400"
              :class="{ 'has-error': errors.timeout }"
              :disabled="loading"
            />
            <span v-if="errors.timeout" class="field-error">{{ errors.timeout }}</span>
          </div>

          <div class="form-group">
            <label for="edit-retryCount">Retry Count</label>
            <input
              id="edit-retryCount"
              v-model.number="retryCount"
              type="number"
              min="0"
              max="10"
              :class="{ 'has-error': errors.retryCount }"
              :disabled="loading"
            />
            <span v-if="errors.retryCount" class="field-error">{{ errors.retryCount }}</span>
          </div>

          <div class="form-group">
            <label for="edit-retryDelay">Retry Delay (s)</label>
            <input
              id="edit-retryDelay"
              v-model.number="retryDelay"
              type="number"
              min="0"
              max="3600"
              :class="{ 'has-error': errors.retryDelay }"
              :disabled="loading"
            />
            <span v-if="errors.retryDelay" class="field-error">{{ errors.retryDelay }}</span>
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </form>

      <div class="modal-footer">
        <button @click="emit('cancel')" class="btn btn-secondary" :disabled="loading">
          Cancel
        </button>
        <button @click="handleSubmit" class="btn btn-primary" :disabled="loading">
          <span v-if="loading">
            <span class="spinner"></span>
            Saving...
          </span>
          <span v-else>Save Changes</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Uses global color variables and classes from style.css */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  width: 100%;
  max-width: 700px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--gray-light);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--black);
  cursor: pointer;
  padding: 0;
  line-height: 1;
  font-weight: 900;
}

.close-btn:hover {
  color: var(--black);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-group label {
  font-weight: 900;
  color: var(--black);
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
  font-size: 0.9375rem;
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
  font-size: 0.8125rem;
  line-height: 1.5;
  resize: vertical;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
}

.checkbox-group {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 700;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
}

.field-error {
  font-size: 0.75rem;
  color: var(--black);
  font-weight: 900;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--gray-dark);
}

.error-message {
  background: var(--gray-lighter);
  color: var(--black);
  padding: 0.75rem 1rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-size: 0.875rem;
  text-align: center;
  font-weight: 900;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 3px solid var(--black);
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
