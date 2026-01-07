<script setup>
import { ref, computed, onMounted } from 'vue'

const props = defineProps({
  schedule: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['save', 'cancel'])

// State
const minutes = ref([])
const hours = ref([])
const daysOfMonth = ref([])
const months = ref([])
const daysOfWeek = ref([])
const saving = ref(false)
const error = ref('')

// Initialize from existing schedule
onMounted(() => {
  if (props.schedule) {
    minutes.value = [...(props.schedule.minutes || [])]
    hours.value = [...(props.schedule.hours || [])]
    daysOfMonth.value = [...(props.schedule.days_of_month || [])]
    months.value = [...(props.schedule.months || [])]
    daysOfWeek.value = [...(props.schedule.days_of_week || [])]
  }
})

// Quick presets
const presets = [
  { label: 'Every minute', minutes: [], hours: [], days: [], months: [], weekdays: [] },
  { label: 'Every hour', minutes: [0], hours: [], days: [], months: [], weekdays: [] },
  { label: 'Daily at midnight', minutes: [0], hours: [0], days: [], months: [], weekdays: [] },
  { label: 'Daily at 9 AM', minutes: [0], hours: [9], days: [], months: [], weekdays: [] },
  { label: 'Weekdays at 9 AM', minutes: [0], hours: [9], days: [], months: [], weekdays: [1, 2, 3, 4, 5] },
  { label: 'Weekly on Monday', minutes: [0], hours: [0], days: [], months: [], weekdays: [1] },
  { label: 'Monthly on 1st', minutes: [0], hours: [0], days: [1], months: [], weekdays: [] }
]

function applyPreset(preset) {
  minutes.value = [...preset.minutes]
  hours.value = [...preset.hours]
  daysOfMonth.value = [...preset.days]
  months.value = [...preset.months]
  daysOfWeek.value = [...preset.weekdays]
}

// Helpers
const minuteOptions = Array.from({ length: 60 }, (_, i) => i)
const hourOptions = Array.from({ length: 24 }, (_, i) => i)
const dayOptions = Array.from({ length: 31 }, (_, i) => i + 1)
const monthOptions = [
  { value: 1, label: 'January' },
  { value: 2, label: 'February' },
  { value: 3, label: 'March' },
  { value: 4, label: 'April' },
  { value: 5, label: 'May' },
  { value: 6, label: 'June' },
  { value: 7, label: 'July' },
  { value: 8, label: 'August' },
  { value: 9, label: 'September' },
  { value: 10, label: 'October' },
  { value: 11, label: 'November' },
  { value: 12, label: 'December' }
]
const weekdayOptions = [
  { value: 0, label: 'Sunday' },
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' }
]

function toggleValue(array, value) {
  const index = array.indexOf(value)
  if (index === -1) {
    array.push(value)
    array.sort((a, b) => a - b)
  } else {
    array.splice(index, 1)
  }
}

function selectAll(array, options) {
  if (Array.isArray(options[0])) {
    array.splice(0, array.length, ...options.map(o => o.value))
  } else {
    array.splice(0, array.length, ...options)
  }
}

function clearAll(array) {
  array.splice(0, array.length)
}

// Preview
const cronPreview = computed(() => {
  const min = minutes.value.length ? minutes.value.join(',') : '*'
  const hour = hours.value.length ? hours.value.join(',') : '*'
  const dom = daysOfMonth.value.length ? daysOfMonth.value.join(',') : '*'
  const month = months.value.length ? months.value.join(',') : '*'
  const dow = daysOfWeek.value.length ? daysOfWeek.value.join(',') : '*'
  return `${min} ${hour} ${dom} ${month} ${dow}`
})

// Validation
function validate() {
  // At least one constraint should be set to avoid running every minute
  const hasConstraint = minutes.value.length > 0 ||
    hours.value.length > 0 ||
    daysOfMonth.value.length > 0 ||
    months.value.length > 0 ||
    daysOfWeek.value.length > 0

  if (!hasConstraint) {
    error.value = 'Warning: This schedule will run every minute. Add at least one constraint.'
    return true // Allow but warn
  }

  error.value = ''
  return true
}

async function handleSave() {
  if (!validate()) return

  saving.value = true
  try {
    const schedule = {
      minutes: minutes.value.length ? minutes.value : null,
      hours: hours.value.length ? hours.value : null,
      days_of_month: daysOfMonth.value.length ? daysOfMonth.value : null,
      months: months.value.length ? months.value : null,
      days_of_week: daysOfWeek.value.length ? daysOfWeek.value : null
    }
    emit('save', schedule)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content">
      <div class="modal-header">
        <h2>{{ schedule ? 'Edit Schedule' : 'Add Schedule' }}</h2>
        <button @click="emit('cancel')" class="close-btn">&times;</button>
      </div>

      <div class="modal-body">
        <!-- Presets -->
        <div class="presets">
          <label>Quick Presets:</label>
          <div class="preset-buttons">
            <button
              v-for="preset in presets"
              :key="preset.label"
              @click="applyPreset(preset)"
              class="preset-btn"
              type="button"
            >
              {{ preset.label }}
            </button>
          </div>
        </div>

        <!-- Minutes -->
        <div class="field-group">
          <div class="field-header">
            <label>Minutes</label>
            <div class="field-actions">
              <button @click="clearAll(minutes)" type="button" class="text-btn">Clear</button>
            </div>
          </div>
          <div class="value-grid minute-grid">
            <button
              v-for="m in minuteOptions"
              :key="m"
              @click="toggleValue(minutes, m)"
              :class="['value-btn', { active: minutes.includes(m) }]"
              type="button"
            >
              {{ m }}
            </button>
          </div>
          <span class="field-hint">Empty = every minute</span>
        </div>

        <!-- Hours -->
        <div class="field-group">
          <div class="field-header">
            <label>Hours</label>
            <div class="field-actions">
              <button @click="clearAll(hours)" type="button" class="text-btn">Clear</button>
            </div>
          </div>
          <div class="value-grid hour-grid">
            <button
              v-for="h in hourOptions"
              :key="h"
              @click="toggleValue(hours, h)"
              :class="['value-btn', { active: hours.includes(h) }]"
              type="button"
            >
              {{ h }}
            </button>
          </div>
          <span class="field-hint">Empty = every hour</span>
        </div>

        <!-- Days of Month -->
        <div class="field-group">
          <div class="field-header">
            <label>Days of Month</label>
            <div class="field-actions">
              <button @click="clearAll(daysOfMonth)" type="button" class="text-btn">Clear</button>
            </div>
          </div>
          <div class="value-grid day-grid">
            <button
              v-for="d in dayOptions"
              :key="d"
              @click="toggleValue(daysOfMonth, d)"
              :class="['value-btn', { active: daysOfMonth.includes(d) }]"
              type="button"
            >
              {{ d }}
            </button>
          </div>
          <span class="field-hint">Empty = every day</span>
        </div>

        <!-- Months -->
        <div class="field-group">
          <div class="field-header">
            <label>Months</label>
            <div class="field-actions">
              <button @click="clearAll(months)" type="button" class="text-btn">Clear</button>
            </div>
          </div>
          <div class="value-grid month-grid">
            <button
              v-for="m in monthOptions"
              :key="m.value"
              @click="toggleValue(months, m.value)"
              :class="['value-btn', { active: months.includes(m.value) }]"
              type="button"
            >
              {{ m.label.slice(0, 3) }}
            </button>
          </div>
          <span class="field-hint">Empty = every month</span>
        </div>

        <!-- Days of Week -->
        <div class="field-group">
          <div class="field-header">
            <label>Days of Week</label>
            <div class="field-actions">
              <button @click="clearAll(daysOfWeek)" type="button" class="text-btn">Clear</button>
            </div>
          </div>
          <div class="value-grid weekday-grid">
            <button
              v-for="d in weekdayOptions"
              :key="d.value"
              @click="toggleValue(daysOfWeek, d.value)"
              :class="['value-btn', { active: daysOfWeek.includes(d.value) }]"
              type="button"
            >
              {{ d.label.slice(0, 3) }}
            </button>
          </div>
          <span class="field-hint">Empty = any day of week</span>
        </div>

        <!-- Preview -->
        <div class="preview">
          <label>Cron Expression:</label>
          <code>{{ cronPreview }}</code>
        </div>

        <div v-if="error" class="warning-message">
          {{ error }}
        </div>
      </div>

      <div class="modal-footer">
        <button @click="emit('cancel')" class="btn btn-secondary" :disabled="saving">
          Cancel
        </button>
        <button @click="handleSave" class="btn btn-primary" :disabled="saving">
          <span v-if="saving">Saving...</span>
          <span v-else>Save Schedule</span>
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
  max-width: 600px;
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
  gap: 1.25rem;
}

.presets {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.presets label {
  font-weight: 900;
  font-size: 0.875rem;
  color: var(--black);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.preset-btn {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  background: var(--white);
  border-radius: 0;
  cursor: pointer;
  transition: none;
  color: var(--black);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preset-btn:hover {
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.field-header label {
  font-weight: 900;
  font-size: 0.875rem;
  color: var(--black);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.field-actions {
  display: flex;
  gap: 0.5rem;
}

.text-btn {
  background: none;
  border: none;
  color: var(--black);
  font-size: 0.75rem;
  cursor: pointer;
  padding: 0;
  font-weight: 700;
  text-decoration: underline;
  transition: none;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.text-btn:hover {
  text-decoration: underline;
}

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

.day-grid {
  grid-template-columns: repeat(10, 1fr);
}

.month-grid {
  grid-template-columns: repeat(6, 1fr);
}

.weekday-grid {
  grid-template-columns: repeat(7, 1fr);
}

.value-btn {
  padding: 0.375rem 0.25rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  background: var(--white);
  border-radius: 0;
  cursor: pointer;
  transition: none;
  color: var(--black);
  font-weight: 700;
}

.value-btn:hover {
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
}

.value-btn.active {
  background: var(--black);
  border: 1px solid var(--gray-light);
  color: var(--white);
}

.field-hint {
  font-size: 0.7rem;
  color: var(--gray-medium);
  font-weight: 700;
}

.preview {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
  border-radius: 0;
}

.preview label {
  font-weight: 900;
  font-size: 0.875rem;
  color: var(--black);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preview code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.875rem;
  background: var(--gray-lighter);
  color: var(--black);
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-weight: 700;
}

.warning-message {
  background: var(--gray-lighter);
  color: var(--black);
  padding: 0.75rem 1rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  font-size: 0.875rem;
  font-weight: 900;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 3px solid var(--black);
}
</style>
