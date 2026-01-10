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
  showSuccessRate: { type: Boolean, default: true }
})

const chartData = computed(() => {
  if (!props.trends || props.trends.length === 0) {
    return { labels: [], datasets: [] }
  }

  const labels = props.trends.map(t => {
    const date = new Date(t.date)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  })

  if (props.showSuccessRate) {
    return {
      labels,
      datasets: [
        {
          label: 'Success Rate',
          data: props.trends.map(t => (t.success_rate * 100).toFixed(1)),
          borderColor: '#4CAF50',
          backgroundColor: 'rgba(76, 175, 80, 0.1)',
          fill: true,
          tension: 0.3,
          yAxisID: 'y'
        }
      ]
    }
  }

  return {
    labels,
    datasets: [
      {
        label: 'Success',
        data: props.trends.map(t => t.success_count),
        borderColor: '#4CAF50',
        backgroundColor: '#4CAF50',
        tension: 0.3
      },
      {
        label: 'Failure',
        data: props.trends.map(t => t.failure_count),
        borderColor: '#f44336',
        backgroundColor: '#f44336',
        tension: 0.3
      },
      {
        label: 'Timeout',
        data: props.trends.map(t => t.timeout_count),
        borderColor: '#FF9800',
        backgroundColor: '#FF9800',
        tension: 0.3
      }
    ]
  }
})

const chartOptions = computed(() => {
  const baseOptions = {
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
        titleFont: { size: 13 },
        bodyFont: { size: 12 }
      }
    },
    scales: {
      x: {
        grid: {
          display: false
        }
      }
    }
  }

  if (props.showSuccessRate) {
    baseOptions.scales.y = {
      beginAtZero: true,
      max: 100,
      ticks: {
        callback: (value) => value + '%'
      }
    }
  } else {
    baseOptions.scales.y = {
      beginAtZero: true,
      ticks: {
        stepSize: 1
      }
    }
  }

  return baseOptions
})
</script>

<template>
  <div class="chart-container">
    <div v-if="!trends || trends.length === 0" class="no-data">
      <p>No execution data available</p>
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
