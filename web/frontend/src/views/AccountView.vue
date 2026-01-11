<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import authService from '../services/auth'

const router = useRouter()
const authStore = useAuthStore()

const currentUser = computed(() => authStore.user)
const isAdmin = computed(() => currentUser.value?.role === 'admin')

// Password form state
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordLoading = ref(false)
const passwordError = ref('')
const passwordSuccess = ref('')

// SMTP form state
const smtpServer = ref('')
const smtpPort = ref(587)
const smtpUsername = ref('')
const smtpPassword = ref('')
const smtpFromName = ref('')
const smtpFromEmail = ref('')
const smtpLoading = ref(false)
const smtpError = ref('')
const smtpSuccess = ref('')

onMounted(async () => {
  if (isAdmin.value) {
    await loadSMTPSettings()
  }
})

async function loadSMTPSettings() {
  try {
    const settings = await authService.getSMTPSettings()
    smtpServer.value = settings.server || ''
    smtpPort.value = settings.port || 587
    smtpUsername.value = settings.username || ''
    smtpPassword.value = settings.password || ''
    smtpFromName.value = settings.from_name || ''
    smtpFromEmail.value = settings.from_email || ''
  } catch (err) {
    console.error('Failed to load SMTP settings:', err)
  }
}

async function changePassword() {
  passwordError.value = ''
  passwordSuccess.value = ''

  // Validate
  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    passwordError.value = 'All fields are required'
    return
  }

  if (newPassword.value.length < 6) {
    passwordError.value = 'New password must be at least 6 characters'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'New passwords do not match'
    return
  }

  passwordLoading.value = true

  try {
    await authService.changePassword(currentPassword.value, newPassword.value)
    passwordSuccess.value = 'Password changed successfully'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err) {
    if (err.response?.data?.error) {
      passwordError.value = err.response.data.error
    } else {
      passwordError.value = 'Failed to change password'
    }
  } finally {
    passwordLoading.value = false
  }
}

async function saveSMTPSettings() {
  smtpError.value = ''
  smtpSuccess.value = ''

  smtpLoading.value = true

  try {
    await authService.updateSMTPSettings({
      server: smtpServer.value,
      port: smtpPort.value,
      username: smtpUsername.value,
      password: smtpPassword.value,
      from_name: smtpFromName.value,
      from_email: smtpFromEmail.value
    })
    smtpSuccess.value = 'SMTP settings saved successfully'
  } catch (err) {
    if (err.response?.data?.error) {
      smtpError.value = err.response.data.error
    } else {
      smtpError.value = 'Failed to save SMTP settings'
    }
  } finally {
    smtpLoading.value = false
  }
}
</script>

<template>
  <div class="main-container">
    <div class="content-area">
      <!-- Back Link -->
      <p class="back-link">
        <a href="#" @click.prevent="router.push('/')">&larr; Back to Dashboard</a>
      </p>

      <!-- Header -->
      <div class="page-header">
        <h1 style="margin: 0;">Account Settings</h1>
      </div>

      <!-- User Info -->
      <div class="card">
        <div class="card-header">Account Information</div>
        <div class="card-body">
          <table class="details-table">
            <tbody>
              <tr>
                <th>Username</th>
                <td>{{ currentUser?.username }}</td>
              </tr>
              <tr>
                <th>Email</th>
                <td>{{ currentUser?.email || '-' }}</td>
              </tr>
              <tr>
                <th>Role</th>
                <td>
                  <span class="badge" :class="{ 'badge-info': currentUser?.role === 'admin' }">
                    {{ currentUser?.role }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Change Password -->
      <div class="card">
        <div class="card-header">Change Password</div>
        <div class="card-body">
          <div v-if="passwordError" class="error-message">{{ passwordError }}</div>
          <div v-if="passwordSuccess" class="success-message">{{ passwordSuccess }}</div>

          <form @submit.prevent="changePassword">
            <div class="form-group">
              <label for="currentPassword">Current Password</label>
              <input
                type="password"
                id="currentPassword"
                v-model="currentPassword"
                :disabled="passwordLoading"
              />
            </div>

            <div class="form-group">
              <label for="newPassword">New Password</label>
              <input
                type="password"
                id="newPassword"
                v-model="newPassword"
                :disabled="passwordLoading"
              />
              <span class="form-hint">Minimum 6 characters</span>
            </div>

            <div class="form-group">
              <label for="confirmPassword">Confirm New Password</label>
              <input
                type="password"
                id="confirmPassword"
                v-model="confirmPassword"
                :disabled="passwordLoading"
              />
            </div>

            <button type="submit" class="btn btn-primary" :disabled="passwordLoading">
              {{ passwordLoading ? 'Changing...' : 'Change Password' }}
            </button>
          </form>
        </div>
      </div>

      <!-- SMTP Settings (Admin Only) -->
      <div v-if="isAdmin" class="card">
        <div class="card-header">SMTP Configuration</div>
        <div class="card-body">
          <p class="text-muted text-small mb-15">Configure SMTP settings for email notifications. These settings override environment variables.</p>

          <div v-if="smtpError" class="error-message">{{ smtpError }}</div>
          <div v-if="smtpSuccess" class="success-message">{{ smtpSuccess }}</div>

          <form @submit.prevent="saveSMTPSettings">
            <div class="form-row">
              <div class="form-group form-group-inline">
                <label for="smtpServer">SMTP Server</label>
                <input
                  type="text"
                  id="smtpServer"
                  v-model="smtpServer"
                  placeholder="smtp.example.com"
                  :disabled="smtpLoading"
                />
              </div>

              <div class="form-group form-group-small">
                <label for="smtpPort">Port</label>
                <input
                  type="number"
                  id="smtpPort"
                  v-model.number="smtpPort"
                  placeholder="587"
                  :disabled="smtpLoading"
                />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group form-group-inline">
                <label for="smtpUsername">Username</label>
                <input
                  type="text"
                  id="smtpUsername"
                  v-model="smtpUsername"
                  placeholder="user@example.com"
                  :disabled="smtpLoading"
                />
              </div>

              <div class="form-group form-group-inline">
                <label for="smtpPassword">Password</label>
                <input
                  type="password"
                  id="smtpPassword"
                  v-model="smtpPassword"
                  placeholder="Leave blank to keep existing"
                  :disabled="smtpLoading"
                />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group form-group-inline">
                <label for="smtpFromName">From Name</label>
                <input
                  type="text"
                  id="smtpFromName"
                  v-model="smtpFromName"
                  placeholder="TaskFlow"
                  :disabled="smtpLoading"
                />
              </div>

              <div class="form-group form-group-inline">
                <label for="smtpFromEmail">From Email</label>
                <input
                  type="email"
                  id="smtpFromEmail"
                  v-model="smtpFromEmail"
                  placeholder="noreply@example.com"
                  :disabled="smtpLoading"
                />
              </div>
            </div>

            <button type="submit" class="btn btn-primary" :disabled="smtpLoading">
              {{ smtpLoading ? 'Saving...' : 'Save SMTP Settings' }}
            </button>
          </form>
        </div>
      </div>
    </div>

    <!-- Sidebar -->
    <div class="sidebar">
      <div class="sidebar-box">
        <div class="sidebar-box-header">Quick Actions</div>
        <div class="sidebar-box-content">
          <p class="mb-10">
            <router-link to="/" class="btn" style="width: 100%; text-align: center;">Dashboard</router-link>
          </p>
          <p class="mb-0">
            <router-link to="/jobs" class="btn" style="width: 100%; text-align: center;">Jobs</router-link>
          </p>
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

.form-hint {
  font-size: 11px;
  color: #666;
  display: block;
  margin-top: 4px;
}

.success-message {
  background-color: #ccffcc;
  border: 1px solid #99cc99;
  padding: 10px;
  color: #006600;
  margin-bottom: 15px;
}

.form-row {
  display: flex;
  gap: 15px;
}

.form-group-inline {
  flex: 1;
}

.form-group-small {
  width: 100px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .form-row {
    flex-direction: column;
    gap: 0;
  }

  .form-group-small {
    width: 100%;
  }
}
</style>
