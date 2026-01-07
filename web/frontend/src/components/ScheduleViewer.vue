<script setup>
import { computed } from 'vue'

const props = defineProps({
  schedule: {
    type: Object,
    required: true
  }
})

function formatField(values, allLabel, singular, plural) {
  if (!values || values.length === 0) {
    return allLabel
  }
  if (values.length === 1) {
    return `${singular} ${values[0]}`
  }
  if (values.length <= 5) {
    return `${plural} ${values.join(', ')}`
  }
  return `${values.length} ${plural.toLowerCase()}`
}

const minutesDisplay = computed(() => {
  return formatField(props.schedule?.minutes, 'Every minute', 'Minute', 'Minutes')
})

const hoursDisplay = computed(() => {
  return formatField(props.schedule?.hours, 'Every hour', 'Hour', 'Hours')
})

const daysDisplay = computed(() => {
  return formatField(props.schedule?.days_of_month, 'Every day', 'Day', 'Days')
})

const monthsDisplay = computed(() => {
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const months = props.schedule?.months
  if (!months || months.length === 0) {
    return 'Every month'
  }
  if (months.length === 1) {
    return monthNames[months[0] - 1] || `Month ${months[0]}`
  }
  if (months.length <= 5) {
    return months.map(m => monthNames[m - 1] || m).join(', ')
  }
  return `${months.length} months`
})

const weekdaysDisplay = computed(() => {
  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const days = props.schedule?.days_of_week
  if (!days || days.length === 0) {
    return 'Any day of week'
  }
  if (days.length === 7) {
    return 'Every day of week'
  }
  if (days.length === 5 && !days.includes(0) && !days.includes(6)) {
    return 'Weekdays'
  }
  if (days.length === 2 && days.includes(0) && days.includes(6)) {
    return 'Weekends'
  }
  return days.map(d => dayNames[d]).join(', ')
})

const cronExpression = computed(() => {
  const min = props.schedule?.minutes?.join(',') || '*'
  const hour = props.schedule?.hours?.join(',') || '*'
  const dom = props.schedule?.days_of_month?.join(',') || '*'
  const month = props.schedule?.months?.join(',') || '*'
  const dow = props.schedule?.days_of_week?.join(',') || '*'
  return `${min} ${hour} ${dom} ${month} ${dow}`
})
</script>

<template>
  <div class="schedule-viewer">
    <div class="schedule-grid">
      <div class="schedule-item">
        <span class="label">Minutes</span>
        <span class="value">{{ minutesDisplay }}</span>
      </div>
      <div class="schedule-item">
        <span class="label">Hours</span>
        <span class="value">{{ hoursDisplay }}</span>
      </div>
      <div class="schedule-item">
        <span class="label">Days</span>
        <span class="value">{{ daysDisplay }}</span>
      </div>
      <div class="schedule-item">
        <span class="label">Months</span>
        <span class="value">{{ monthsDisplay }}</span>
      </div>
      <div class="schedule-item">
        <span class="label">Weekdays</span>
        <span class="value">{{ weekdaysDisplay }}</span>
      </div>
    </div>
    <div class="cron-expression">
      <span class="label">Cron:</span>
      <code>{{ cronExpression }}</code>
    </div>
  </div>
</template>

<style scoped>
/* Uses global color variables and classes from style.css */
.schedule-viewer {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.schedule-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 0.75rem;
}

.schedule-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.schedule-item .label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--black);
  font-weight: 900;
}

.schedule-item .value {
  font-size: 0.875rem;
  color: var(--black);
  font-weight: 700;
}

.cron-expression {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-top: 0.75rem;
  border-top: 2px solid var(--black);
}

.cron-expression .label {
  font-size: 0.75rem;
  text-transform: uppercase;
  color: var(--black);
  font-weight: 900;
}

.cron-expression code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.8125rem;
  background: var(--gray-lighter);
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  color: var(--black);
  font-weight: 700;
}
</style>
