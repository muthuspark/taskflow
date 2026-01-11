<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRunsStore } from '../stores/runs'
import { LogStreamClient } from '../services/websocket'
import StatusBadge from '../components/StatusBadge.vue'

const route = useRoute()
const router = useRouter()
const runsStore = useRunsStore()

// State
const run = computed(() => runsStore.currentRun)
const logs = computed(() => runsStore.logs)
const loading = computed(() => runsStore.loading)
const error = computed(() => runsStore.error)
const wsClient = ref(null)
const wsConnected = ref(false)
const autoScroll = ref(true)
const logsContainer = ref(null)

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
    scrollToBottom()
  } else if (message.type === 'status') {
    // Refresh run status
    runsStore.fetchRun(route.params.id)
    if (message.data?.status && !['pending', 'running'].includes(message.data.status)) {
      disconnectWebSocket()
    }
  }
}

function scrollToBottom() {
  if (autoScroll.value && logsContainer.value) {
    nextTick(() => {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight
    })
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

function refreshLogs() {
  runsStore.fetchLogs(route.params.id)
}

function goToJob() {
  if (run.value?.job_id) {
    router.push(`/jobs/${run.value.job_id}`)
  }
}

function openLogsInNewTab() {
  const url = router.resolve(`/runs/${route.params.id}/logs`).href
  window.open(url, '_blank')
}
</script>

<template>
  <div class="main-container">
    <div class="content-area">
      <div v-if="loading && !run" class="loading-container">
        <div class="spinner"></div>
        <p>Loading run details...</p>
      </div>

      <div v-else-if="error && !run" class="error-message">
        {{ error }}
        <button @click="router.push('/runs')" class="btn btn-small" style="margin-left: 10px;">Back to Runs</button>
      </div>

      <template v-else-if="run">
        <!-- Back Link -->
        <p class="back-link">
          <a href="#" @click.prevent="router.push('/runs')">&larr; Back to Runs</a>
        </p>

        <!-- Header -->
        <div class="page-header">
          <div>
            <h1 style="margin: 0;">Run Details</h1>
          </div>
          <div class="header-actions">
            <button @click="goToJob" class="btn">View Job</button>
            <button @click="openLogsInNewTab" class="btn">Open Logs in New Tab</button>
            <button @click="refreshLogs" class="btn" :disabled="loading">Refresh</button>
          </div>
        </div>

        <!-- Run Info -->
        <div class="card">
          <div class="card-header">
            Run Information
            <StatusBadge :status="run.status" />
          </div>
          <div class="card-body">
            <table class="details-table">
              <tbody>
                <tr>
                  <th>Run ID</th>
                  <td><code>{{ run.id }}</code></td>
                </tr>
                <tr>
                  <th>Job</th>
                  <td><a href="#" @click.prevent="goToJob">{{ run.job_id }}</a></td>
                </tr>
                <tr>
                  <th>Status</th>
                  <td><StatusBadge :status="run.status" /></td>
                </tr>
                <tr>
                  <th>Trigger</th>
                  <td><span class="badge">{{ run.trigger_type || 'manual' }}</span></td>
                </tr>
                <tr>
                  <th>Started</th>
                  <td>{{ formatDate(run.started_at) }}</td>
                </tr>
                <tr>
                  <th>Finished</th>
                  <td>{{ formatDate(run.finished_at) }}</td>
                </tr>
                <tr>
                  <th>Duration</th>
                  <td>{{ formatDuration(run.duration_ms) }}</td>
                </tr>
                <tr>
                  <th>Exit Code</th>
                  <td><strong>{{ run.exit_code ?? '-' }}</strong></td>
                </tr>
                <tr v-if="run.error_message">
                  <th>Error</th>
                  <td class="error-text">{{ run.error_message }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Logs -->
        <div class="card">
          <div class="card-header">
            <span>Logs</span>
            <div class="log-controls">
              <label class="autoscroll-label">
                <input type="checkbox" v-model="autoScroll" />
                Auto-scroll
              </label>
              <span v-if="wsConnected" class="ws-status connected">Live</span>
              <span v-else-if="run.status === 'running'" class="ws-status disconnected">Reconnecting...</span>
            </div>
          </div>
          <div class="card-body" style="padding: 0;">
            <div ref="logsContainer" class="logs-container">
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
          </div>
        </div>
      </template>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">Quick Actions</div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <button @click="goToJob" class="btn" style="width: 100%;">View Job</button>
          </p>
          <p class="mb-10">
            <button @click="refreshLogs" class="btn" style="width: 100%;" :disabled="loading">Refresh Logs</button>
          </p>
          <p class="mb-10">
            <button @click="openLogsInNewTab" class="btn" style="width: 100%;">Full Page Logs</button>
          </p>
          <p class="mb-0">
            <router-link to="/runs" class="btn" style="width: 100%; text-align: center;">All Runs</router-link>
          </p>
        </div>
      </div>

      <div class="sidebar-box" v-if="run">
        <div class="sidebar-box-header">Run Status</div>
        <div class="sidebar-box-content">
          <p class="text-small"><strong>Status:</strong> <StatusBadge :status="run.status" /></p>
          <p class="text-small"><strong>Exit Code:</strong> {{ run.exit_code ?? 'N/A' }}</p>
          <p class="text-small mb-0"><strong>Duration:</strong> {{ formatDuration(run.duration_ms) }}</p>
        </div>
      </div>

      <div class="sidebar-box">
        <div class="sidebar-box-header">Legend</div>
        <div class="sidebar-box-content">
          <p class="text-small"><span class="legend-stdout">[stdout]</span> Standard output</p>
          <p class="text-small"><span class="legend-stderr">[stderr]</span> Error output</p>
          <p class="text-small mb-0"><span class="legend-system">[system]</span> System messages</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.back-link {
  margin-bottom: 10px;
}

.back-link a {
  font-size: 12px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.details-table {
  width: 100%;
  margin: 0;
}

.details-table th {
  width: 100px;
  background: #f4f4f4;
  font-weight: bold;
  text-align: left;
}

.details-table td code {
  background: #f4f4f4;
  padding: 2px 6px;
  border: 1px solid #cccccc;
  font-family: "Courier New", monospace;
  font-size: 11px;
}

.error-text {
  color: #cc0000;
}

.log-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.autoscroll-label {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  cursor: pointer;
}

.autoscroll-label input {
  width: auto;
}

.ws-status {
  font-size: 10px;
  padding: 2px 8px;
  border: 1px solid;
}

.ws-status.connected {
  background: #ccffcc;
  border-color: #99cc99;
  color: #006600;
}

.ws-status.disconnected {
  background: #ffffcc;
  border-color: #cccc99;
  color: #666600;
}

.logs-container {
  background: #1e1e1e;
  min-height: 300px;
  max-height: 500px;
  overflow: auto;
}

.empty-logs {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: #666666;
}

.empty-logs p {
  margin: 0;
}

.log-entries {
  padding: 10px;
  font-family: "Courier New", monospace;
  font-size: 11px;
  line-height: 1.5;
}

.log-entry {
  display: flex;
  gap: 8px;
  padding: 1px 0;
  color: #cccccc;
}

.log-entry.log-stderr {
  color: #ff6b6b;
}

.log-entry.log-system {
  color: #888888;
  font-style: italic;
}

.log-timestamp {
  color: #666666;
  flex-shrink: 0;
}

.log-stream {
  color: #888888;
  flex-shrink: 0;
  min-width: 60px;
}

.log-content {
  white-space: pre-wrap;
  word-break: break-all;
}

.legend-stdout {
  color: #cccccc;
  font-family: "Courier New", monospace;
  font-size: 11px;
}

.legend-stderr {
  color: #ff6b6b;
  font-family: "Courier New", monospace;
  font-size: 11px;
}

.legend-system {
  color: #888888;
  font-family: "Courier New", monospace;
  font-size: 11px;
  font-style: italic;
}
</style>
