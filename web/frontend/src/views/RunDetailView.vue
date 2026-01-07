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
</script>

<template>
  <div class="run-detail">
    <div v-if="loading && !run" class="loading-container">
      <div class="spinner-large"></div>
      <p>Loading run details...</p>
    </div>

    <div v-else-if="error && !run" class="error-container">
      <p>{{ error }}</p>
      <button @click="router.push('/runs')" class="btn btn-secondary">Back to Runs</button>
    </div>

    <template v-else-if="run">
      <!-- Header -->
      <div class="page-header">
        <div class="header-left">
          <button @click="router.push('/runs')" class="btn btn-link">
            &larr; Back to Runs
          </button>
          <h1>
            Run Details
            <StatusBadge :status="run.status" />
          </h1>
        </div>
        <div class="header-actions">
          <button @click="goToJob" class="btn btn-secondary">View Job</button>
          <button @click="refreshLogs" class="btn btn-secondary" :disabled="loading">
            Refresh
          </button>
        </div>
      </div>

      <!-- Run Info -->
      <div class="info-card">
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Job ID</span>
            <span class="value">
              <a @click.prevent="goToJob" href="#" class="link">{{ run.job_id }}</a>
            </span>
          </div>
          <div class="info-item">
            <span class="label">Status</span>
            <span class="value"><StatusBadge :status="run.status" /></span>
          </div>
          <div class="info-item">
            <span class="label">Started</span>
            <span class="value">{{ formatDate(run.started_at) }}</span>
          </div>
          <div class="info-item">
            <span class="label">Finished</span>
            <span class="value">{{ formatDate(run.finished_at) }}</span>
          </div>
          <div class="info-item">
            <span class="label">Duration</span>
            <span class="value">{{ formatDuration(run.duration_ms) }}</span>
          </div>
          <div class="info-item">
            <span class="label">Exit Code</span>
            <span class="value" :class="{ 'exit-error': run.exit_code !== 0 }">
              {{ run.exit_code ?? '-' }}
            </span>
          </div>
          <div class="info-item">
            <span class="label">Trigger</span>
            <span class="value trigger-badge" :class="run.trigger_type || 'manual'">
              {{ run.trigger_type || 'manual' }}
            </span>
          </div>
          <div v-if="run.error_message" class="info-item full-width">
            <span class="label">Error</span>
            <span class="value error-text">{{ run.error_message }}</span>
          </div>
        </div>
      </div>

      <!-- Logs -->
      <div class="logs-card">
        <div class="logs-header">
          <h2>Logs</h2>
          <div class="logs-controls">
            <label class="checkbox-label">
              <input type="checkbox" v-model="autoScroll" />
              <span>Auto-scroll</span>
            </label>
            <span v-if="wsConnected" class="ws-status connected">
              Live streaming
            </span>
            <span v-else-if="run.status === 'running'" class="ws-status disconnected">
              Reconnecting...
            </span>
          </div>
        </div>

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
              <span v-if="log.timestamp" class="log-timestamp">
                {{ new Date(log.timestamp).toLocaleTimeString() }}
              </span>
              <span v-if="log.stream" class="log-stream">[{{ log.stream }}]</span>
              <span class="log-content">{{ log.content }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.run-detail {
  padding: 0;
}

.loading-container,
.error-container {
  text-align: center;
  padding: 3rem;
}

.spinner-large {
  width: 48px;
  height: 48px;
  border: 4px solid var(--gray-light);
  border-top-color: var(--black);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-container {
  background: var(--gray-lighter);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  color: var(--black);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.header-left h1 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.btn-link {
  background: none;
  border: none;
  color: var(--black);
  padding: 0;
  cursor: pointer;
  font-size: 0.875rem;
  text-decoration: underline;
  font-weight: 700;
}

.btn-link:hover {
  text-decoration: underline;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.info-card {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  padding: 1.5rem;
  box-shadow: none;
  margin-bottom: 1.5rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-item.full-width {
  grid-column: 1 / -1;
}

.info-item .label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--black);
  font-weight: 900;
}

.info-item .value {
  font-size: 0.875rem;
  color: var(--black);
}

.exit-error {
  color: var(--black);
  font-weight: 900;
}

.error-text {
  color: var(--black);
  font-weight: 900;
}

.trigger-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  width: fit-content;
  background: var(--white);
  color: var(--black);
  font-weight: 700;
}

.trigger-badge.manual {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}

.trigger-badge.scheduled {
  background: var(--white);
  color: var(--black);
  border: 1px solid var(--gray-light);
}

.link {
  color: var(--black);
  text-decoration: underline;
  cursor: pointer;
  font-weight: 700;
}

.link:hover {
  text-decoration: underline;
}

.logs-card {
  background: var(--white);
  border: 1px solid var(--gray-light);
  border-radius: 0;
  box-shadow: none;
  overflow: hidden;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--gray-light);
}

.logs-header h2 {
  margin: 0;
  font-size: 1rem;
  color: var(--black);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.logs-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--black);
  cursor: pointer;
  font-weight: 700;
}

.ws-status {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 700;
}

.ws-status.connected {
  background: var(--white);
  color: var(--black);
}

.ws-status.disconnected {
  background: var(--white);
  color: var(--black);
}

.logs-container {
  background: var(--gray-dark);
  min-height: 400px;
  max-height: 600px;
  overflow: auto;
  border-top: 2px solid var(--black);
}

.empty-logs {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: var(--gray-light);
}

.log-entries {
  padding: 1rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
}

.log-entry {
  display: flex;
  gap: 0.5rem;
  padding: 0.125rem 0;
  color: var(--white);
}

.log-entry.log-stderr {
  color: var(--white);
  font-weight: 700;
}

.log-entry.log-system {
  color: var(--white);
  font-style: italic;
}

.log-timestamp {
  color: var(--gray-light);
  flex-shrink: 0;
}

.log-stream {
  color: var(--gray-light);
  flex-shrink: 0;
  min-width: 60px;
}

.log-content {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
