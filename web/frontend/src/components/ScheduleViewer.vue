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
  <div class="flex flex-col gap-4">
    <div class="grid grid-cols-[repeat(auto-fit,minmax(120px,1fr))] gap-3">
      <div class="flex flex-col gap-1">
        <span class="text-xs uppercase tracking-tight text-black font-black">Minutes</span>
        <span class="text-sm text-black font-bold">{{ minutesDisplay }}</span>
      </div>
      <div class="flex flex-col gap-1">
        <span class="text-xs uppercase tracking-tight text-black font-black">Hours</span>
        <span class="text-sm text-black font-bold">{{ hoursDisplay }}</span>
      </div>
      <div class="flex flex-col gap-1">
        <span class="text-xs uppercase tracking-tight text-black font-black">Days</span>
        <span class="text-sm text-black font-bold">{{ daysDisplay }}</span>
      </div>
      <div class="flex flex-col gap-1">
        <span class="text-xs uppercase tracking-tight text-black font-black">Months</span>
        <span class="text-sm text-black font-bold">{{ monthsDisplay }}</span>
      </div>
      <div class="flex flex-col gap-1">
        <span class="text-xs uppercase tracking-tight text-black font-black">Weekdays</span>
        <span class="text-sm text-black font-bold">{{ weekdaysDisplay }}</span>
      </div>
    </div>
    <div class="flex items-center gap-2 pt-3 border-t-2 border-black">
      <span class="text-xs uppercase text-black font-black">Cron:</span>
      <code class="font-mono text-[0.8125rem] bg-gray-lighter px-2 py-1 border border-gray-light text-black font-bold">{{ cronExpression }}</code>
    </div>
  </div>
</template>

<style scoped>
/* All styles handled by Tailwind utilities */
</style>
