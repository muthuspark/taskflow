<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useRunsStore } from '../stores/runs'
import { LogStreamClient } from '../services/websocket'

const route = useRoute()
const runsStore = useRunsStore()

// State
const run = computed(() => runsStore.currentRun)
const logs = computed(() => runsStore.logs)
const loading = computed(() => runsStore.loading)
const error = computed(() => runsStore.error)
const wsClient = ref(null)
const wsConnected = ref(false)

// Fetch run and logs
onMounted(async () => {
  const runId = route.params.id
  await Promise.all([
    runsStore.fetchRun(runId),
    runsStore.fetchLogs(runId)
  ])

  // Connect WebSocket for live streaming if run is in progress
  if (run.value && (run.value.status === 'pending' || run.value.status === 'running')) {
    connectWebSocket(runId)
  }

  // Update page title
  if (run.value) {
    document.title = `Logs - Run ${run.value.id} - TaskFlow`
  }
})

onUnmounted(() => {
  disconnectWebSocket()
  runsStore.clearCurrentRun()
})

function connectWebSocket(runId) {
  wsClient.value = new LogStreamClient(runId)
  wsClient.value.onMessage(handleWsMessage)

  wsClient.value.connect()
    .then(() => {
      wsConnected.value = true
    })
    .catch(err => {
      console.error('WebSocket connection failed:', err)
      wsConnected.value = false
    })
}

function disconnectWebSocket() {
  if (wsClient.value) {
    wsClient.value.disconnect()
    wsClient.value = null
    wsConnected.value = false
  }
}

function handleWsMessage(message) {
  if (message.type === 'log' && message.data) {
    runsStore.addLog({
      stream: message.data.stream,
      content: message.data.content,
      timestamp: message.timestamp
    })
  } else if (message.type === 'status') {
    runsStore.fetchRun(route.params.id)
    if (message.data?.status && !['pending', 'running'].includes(message.data.status)) {
      disconnectWebSocket()
    }
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

function formatDuration(ms) {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds}s`
}

function getStreamClass(stream) {
  switch (stream) {
    case 'stderr': return 'log-stderr'
    case 'system': return 'log-system'
    default: return 'log-stdout'
  }
}

function printPage() {
  window.print()
}
</script>

<template>
  <div class="print-page">
    <!-- Print Controls (hidden when printing) -->
    <div class="print-controls no-print">
      <button @click="printPage" class="btn">Print</button>
      <span v-if="wsConnected" class="live-badge">Live</span>
    </div>

    <!-- Loading State -->
    <div v-if="loading && !run" class="loading">
      Loading run details...
    </div>

    <!-- Error State -->
    <div v-else-if="error && !run" class="error">
      {{ error }}
    </div>

    <!-- Content -->
    <template v-else-if="run">
      <!-- Header (hidden when printing) -->
      <div class="header no-print">
        <h1>TaskFlow - Run Logs</h1>
        <div class="run-info">
          <div class="info-row">
            <span class="label">Run ID:</span>
            <span class="value">{{ run.id }}</span>
          </div>
          <div class="info-row">
            <span class="label">Job ID:</span>
            <span class="value">{{ run.job_id }}</span>
          </div>
          <div class="info-row">
            <span class="label">Status:</span>
            <span class="value status" :class="run.status">{{ run.status }}</span>
          </div>
          <div class="info-row">
            <span class="label">Started:</span>
            <span class="value">{{ formatDate(run.started_at) }}</span>
          </div>
          <div class="info-row">
            <span class="label">Finished:</span>
            <span class="value">{{ formatDate(run.finished_at) }}</span>
          </div>
          <div class="info-row">
            <span class="label">Duration:</span>
            <span class="value">{{ formatDuration(run.duration_ms) }}</span>
          </div>
          <div class="info-row">
            <span class="label">Exit Code:</span>
            <span class="value">{{ run.exit_code ?? '-' }}</span>
          </div>
          <div v-if="run.error_message" class="info-row">
            <span class="label">Error:</span>
            <span class="value error-text">{{ run.error_message }}</span>
          </div>
        </div>
      </div>

      <hr class="divider no-print" />

      <!-- Logs Section -->
      <div class="logs-section">
        <h2 class="no-print">Output</h2>
        <div v-if="!logs.length" class="empty-logs">
          <p v-if="run.status === 'pending'">Waiting for job to start...</p>
          <p v-else-if="run.status === 'running'">Waiting for output...</p>
          <p v-else>No output captured</p>
        </div>
        <div v-else class="log-entries">
          <div
            v-for="(log, index) in logs"
            :key="index"
            class="log-entry"
            :class="getStreamClass(log.stream)"
          >
            <span v-if="log.timestamp" class="log-timestamp">{{ new Date(log.timestamp).toLocaleTimeString() }}</span>
            <span v-if="log.stream" class="log-stream">[{{ log.stream }}]</span>
            <span class="log-content">{{ log.content }}</span>
          </div>
        </div>
      </div>

      <!-- Footer (hidden when printing) -->
      <div class="footer no-print">
        <p>Generated: {{ new Date().toLocaleString() }}</p>
      </div>
    </template>
  </div>
</template>

<style scoped>
/* Base styles */
.print-page {
  font-family: "Courier New", Courier, monospace;
  font-size: 12px;
  line-height: 1.4;
  background: #fff;
  color: #000;
  padding: 20px;
  min-height: 100vh;
}

/* Print controls - hidden when printing */
.print-controls {
  position: fixed;
  top: 10px;
  right: 10px;
  display: flex;
  gap: 10px;
  align-items: center;
  z-index: 1000;
}

.live-badge {
  background: #ccffcc;
  border: 1px solid #99cc99;
  color: #006600;
  padding: 2px 8px;
  font-size: 10px;
}

/* Loading and error */
.loading, .error {
  text-align: center;
  padding: 40px;
  font-size: 14px;
}

.error {
  color: #cc0000;
}

/* Header */
.header {
  margin-bottom: 20px;
}

.header h1 {
  font-size: 18px;
  margin: 0 0 15px 0;
  padding-bottom: 10px;
  border-bottom: 2px solid #000;
}

.run-info {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 5px 20px;
}

.info-row {
  display: flex;
  gap: 8px;
}

.info-row .label {
  font-weight: bold;
  min-width: 80px;
}

.info-row .value {
  flex: 1;
}

.info-row .value.status {
  text-transform: uppercase;
  font-weight: bold;
}

.info-row .value.status.success {
  color: #006600;
}

.info-row .value.status.failed {
  color: #cc0000;
}

.info-row .value.status.running {
  color: #0066cc;
}

.info-row .value.status.pending {
  color: #666666;
}

.info-row .value.status.timeout {
  color: #cc6600;
}

.error-text {
  color: #cc0000;
}

/* Divider */
.divider {
  border: none;
  border-top: 1px solid #ccc;
  margin: 20px 0;
}

/* Logs section */
.logs-section h2 {
  font-size: 14px;
  margin: 0 0 10px 0;
}

.empty-logs {
  color: #666;
  font-style: italic;
  padding: 20px 0;
}

.log-entries {
  background: #f8f8f8;
  border: 1px solid #ddd;
  padding: 10px;
}

.log-entry {
  display: flex;
  gap: 8px;
  padding: 1px 0;
  color: #333;
}

.log-entry.log-stderr {
  color: #cc0000;
}

.log-entry.log-system {
  color: #666;
  font-style: italic;
}

.log-timestamp {
  color: #888;
  flex-shrink: 0;
}

.log-stream {
  color: #666;
  flex-shrink: 0;
  min-width: 60px;
}

.log-content {
  white-space: pre-wrap;
  word-break: break-all;
}

/* Footer */
.footer {
  margin-top: 30px;
  padding-top: 10px;
  border-top: 1px solid #ccc;
  font-size: 10px;
  color: #888;
}

.footer p {
  margin: 0;
}

/* Print styles */
@media print {
  .no-print {
    display: none !important;
  }

  .print-page {
    padding: 0;
    font-size: 10px;
  }

  .logs-section {
    margin: 0;
  }

  .log-entries {
    border: none;
    background: #fff;
    padding: 0;
  }

  .log-entry {
    page-break-inside: avoid;
  }
}
</style>
