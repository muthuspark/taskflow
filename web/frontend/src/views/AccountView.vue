<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import authService from '../services/auth'

const router = useRouter()
const authStore = useAuthStore()

const currentUser = computed(() => authStore.user)

// Form state
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

async function changePassword() {
  error.value = ''
  success.value = ''

  // Validate
  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    error.value = 'All fields are required'
    return
  }

  if (newPassword.value.length < 6) {
    error.value = 'New password must be at least 6 characters'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    error.value = 'New passwords do not match'
    return
  }

  loading.value = true

  try {
    await authService.changePassword(currentPassword.value, newPassword.value)
    success.value = 'Password changed successfully'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err) {
    if (err.response?.data?.error?.message) {
      error.value = err.response.data.error.message
    } else {
      error.value = 'Failed to change password'
    }
  } finally {
    loading.value = false
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
          <div v-if="error" class="error-message">{{ error }}</div>
          <div v-if="success" class="success-message">{{ success }}</div>

          <form @submit.prevent="changePassword">
            <div class="form-group">
              <label for="currentPassword">Current Password</label>
              <input
                type="password"
                id="currentPassword"
                v-model="currentPassword"
                :disabled="loading"
              />
            </div>

            <div class="form-group">
              <label for="newPassword">New Password</label>
              <input
                type="password"
                id="newPassword"
                v-model="newPassword"
                :disabled="loading"
              />
              <span class="form-hint">Minimum 6 characters</span>
            </div>

            <div class="form-group">
              <label for="confirmPassword">Confirm New Password</label>
              <input
                type="password"
                id="confirmPassword"
                v-model="confirmPassword"
                :disabled="loading"
              />
            </div>

            <button type="submit" class="btn btn-primary" :disabled="loading">
              {{ loading ? 'Changing...' : 'Change Password' }}
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
</style>
