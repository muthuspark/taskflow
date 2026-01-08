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
  <div>
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
      <div class="flex justify-between items-start flex-wrap gap-4 mb-6">
        <div class="flex flex-col gap-2">
          <button @click="router.push('/runs')" class="bg-none border-none text-black p-0 cursor-pointer text-sm underline font-bold text-left">
            &larr; Back to Runs
          </button>
          <div class="flex items-center gap-3">
            <h1 class="m-0 text-black font-black uppercase tracking-tight">Run Details</h1>
            <StatusBadge :status="run.status" />
          </div>
        </div>
        <div class="flex gap-2">
          <button @click="goToJob" class="btn btn-secondary">View Job</button>
          <button @click="refreshLogs" class="btn btn-secondary" :disabled="loading">
            Refresh
          </button>
        </div>
      </div>

      <!-- Run Info -->
      <div class="bg-white border border-gray-light p-6 mb-6">
        <div class="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-4">
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Job ID</span>
            <span class="text-sm text-black">
              <a @click.prevent="goToJob" href="#" class="text-black underline font-bold cursor-pointer">{{ run.job_id }}</a>
            </span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Status</span>
            <span class="text-sm"><StatusBadge :status="run.status" /></span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Started</span>
            <span class="text-sm text-gray-medium">{{ formatDate(run.started_at) }}</span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Finished</span>
            <span class="text-sm text-gray-medium">{{ formatDate(run.finished_at) }}</span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Duration</span>
            <span class="text-sm text-black">{{ formatDuration(run.duration_ms) }}</span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Exit Code</span>
            <span class="text-sm text-black font-bold">{{ run.exit_code ?? '-' }}</span>
          </div>
          <div class="flex flex-col gap-1">
            <span class="text-xs uppercase tracking-tight text-black font-black">Trigger</span>
            <span class="trigger-badge" :class="run.trigger_type || 'manual'">
              {{ run.trigger_type || 'manual' }}
            </span>
          </div>
          <div v-if="run.error_message" class="flex flex-col gap-1 col-span-full">
            <span class="text-xs uppercase tracking-tight text-black font-black">Error</span>
            <span class="text-sm text-black font-bold">{{ run.error_message }}</span>
          </div>
        </div>
      </div>

      <!-- Logs -->
      <div class="bg-white border border-gray-light overflow-hidden">
        <div class="flex justify-between items-center p-6 border-b border-gray-light">
          <h2 class="m-0 text-black font-black uppercase tracking-tight text-sm">Logs</h2>
          <div class="flex items-center gap-4">
            <label class="flex items-center gap-2 text-sm text-black cursor-pointer font-bold">
              <input type="checkbox" v-model="autoScroll" />
              <span>Auto-scroll</span>
            </label>
            <span v-if="wsConnected" class="ws-status connected text-xs px-2 py-1 border border-gray-light font-bold uppercase tracking-tight">
              Live streaming
            </span>
            <span v-else-if="run.status === 'running'" class="ws-status disconnected text-xs px-2 py-1 border border-gray-light font-bold uppercase tracking-tight">
              Reconnecting...
            </span>
          </div>
        </div>

        <div ref="logsContainer" class="logs-container">
          <div v-if="!logs.length" class="flex items-center justify-center min-h-[200px] text-gray-light">
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
/* Trigger badge styling */
.trigger-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--gray-light);
  border-radius: 0;
  background: var(--white);
  color: var(--black);
  width: fit-content;
}

.ws-status {
  background: var(--white);
  color: var(--black);
}

/* Specialized log viewer styling - dark terminal theme */
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
