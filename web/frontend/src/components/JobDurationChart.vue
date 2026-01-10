<script setup>
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const props = defineProps({
  trends: { type: Array, default: () => [] },
  jobName: { type: String, default: '' }
})

function formatDuration(ms) {
  if (!ms || ms === 0) return '0s'
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

const chartData = computed(() => {
  if (!props.trends || props.trends.length === 0) {
    return { labels: [], datasets: [] }
  }

  const labels = props.trends.map(t => {
    const date = new Date(t.date)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  })

  return {
    labels,
    datasets: [
      {
        label: 'Average Duration',
        data: props.trends.map(t => t.avg_duration_ms / 1000),
        borderColor: '#2196F3',
        backgroundColor: 'rgba(33, 150, 243, 0.1)',
        fill: true,
        tension: 0.3
      },
      {
        label: 'Min Duration',
        data: props.trends.map(t => t.min_duration_ms / 1000),
        borderColor: '#4CAF50',
        backgroundColor: 'transparent',
        borderDash: [5, 5],
        tension: 0.3,
        pointRadius: 2
      },
      {
        label: 'Max Duration',
        data: props.trends.map(t => t.max_duration_ms / 1000),
        borderColor: '#f44336',
        backgroundColor: 'transparent',
        borderDash: [5, 5],
        tension: 0.3,
        pointRadius: 2
      }
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index'
  },
  plugins: {
    legend: {
      position: 'top',
      labels: {
        usePointStyle: true,
        padding: 15
      }
    },
    tooltip: {
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      padding: 12,
      callbacks: {
        label: function(context) {
          const value = context.raw
          return `${context.dataset.label}: ${formatDuration(value * 1000)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        display: false
      }
    },
    y: {
      beginAtZero: true,
      ticks: {
        callback: (value) => formatDuration(value * 1000)
      }
    }
  }
}
</script>

<template>
  <div class="chart-container">
    <div v-if="!trends || trends.length === 0" class="no-data">
      <p>No duration data available for this job</p>
    </div>
    <Line v-else :data="chartData" :options="chartOptions" />
  </div>
</template>

<style scoped>
.chart-container {
  position: relative;
  height: 300px;
  width: 100%;
}

.no-data {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #666;
}

.no-data p {
  margin: 0;
}
</style>
