import { API_BASE_PATH } from './api'

/**
 * WebSocket client for streaming logs from a run
 */
export class LogStreamClient {
  /**
   * @param {string} runId - The run ID to stream logs for
   */
  constructor(runId) {
    this.runId = runId
    this.ws = null
    this.messageCallbacks = []
    this.connected = false
  }

  /**
   * Connect to the WebSocket server
   * @returns {Promise<void>}
   */
  connect() {
    return new Promise((resolve, reject) => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const token = localStorage.getItem('token')

      const url = `${protocol}//${host}${API_BASE_PATH}/ws/logs?run_id=${this.runId}&token=${encodeURIComponent(token || '')}`

      this.ws = new WebSocket(url)

      this.ws.onopen = () => {
        this.connected = true
        resolve()
      }

      this.ws.onerror = (error) => {
        this.connected = false
        reject(error)
      }

      this.ws.onclose = () => {
        this.connected = false
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          this.messageCallbacks.forEach(callback => callback(message))
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }
    })
  }

  /**
   * Register a callback for incoming messages
   * @param {function} callback - Function to call with parsed message
   */
  onMessage(callback) {
    this.messageCallbacks.push(callback)
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
      this.connected = false
      this.messageCallbacks = []
    }
  }

  /**
   * Check if connected
   * @returns {boolean}
   */
  isConnected() {
    return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

export default LogStreamClient
