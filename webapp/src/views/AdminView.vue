<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { auth, impersonate as doImpersonate, clearImpersonate } from '../authStore.js'
import { config } from '../config.js'
import {
    getServerStats, getAdminUsers, getAdminUploads,
    createUser as apiCreateUser, deleteUser as apiDeleteUser,
    updateUser, removeUpload, getFileURL, getVersion
} from '../api.js'
import {
    humanReadableSize, getUploadUrl, quotaLabel, ttlLabel,
    clampQuota, defaultSizeHint, defaultTTLHint, TTL_UNITS,
    formatDate, buildEditForm, buildEditPayload,
} from '../utils.js'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const router = useRouter()

// ── Display mode ──
const display = ref('stats') // 'stats' | 'users' | 'uploads'

// ── Version info ──
const version = ref(null)

// ── Stats ──
const stats = ref(null)
const statsLoading = ref(false)

// ── Users ──
const users = ref([])
const usersCursor = ref(null)
const usersLoading = ref(false)

// ── Uploads ──
const uploads = ref([])
const uploadsCursor = ref(null)
const uploadsLoading = ref(false)
const uploadsUserFilter = ref('')
const uploadsTokenFilter = ref('')
const uploadsSortBy = ref('date') // 'date' | 'size'
const uploadsSortOrder = ref('desc') // 'desc' | 'asc'

// ── Create user modal ──
const showCreateUser = ref(false)
const createForm = ref({ provider: 'local', login: '', password: '', name: '', email: '', admin: false, maxFileSize: 0, maxUserSize: 0, maxTTL: 0 })
const createTTLUnit = ref(60)
const createError = ref('')
const createSaving = ref(false)

// ── Edit user modal ──
const showEditUser = ref(false)
const editForm = ref({})
const editError = ref('')
const editSaving = ref(false)

// ── Confirm dialog ──
const confirm = ref(null)

// ── Helpers ──

// Edit form display helpers
const editTTLUnit = ref(60)

// ── Stats API ──
async function loadStats() {
    statsLoading.value = true
    try {
        stats.value = await getServerStats()
    } catch (err) {
        console.error('Failed to load stats', err)
    } finally {
        statsLoading.value = false
    }
}

// ── Users API ──
async function loadUsers(more = false) {
    usersLoading.value = true
    try {
        const opts = { limit: 50 }
        if (more && usersCursor.value) opts.after = usersCursor.value
        const data = await getAdminUsers(opts)
        if (more) {
            users.value = [...users.value, ...data.results]
        } else {
            users.value = data.results || []
        }
        usersCursor.value = data.after || null
    } catch (err) {
        console.error('Failed to load users', err)
    } finally {
        usersLoading.value = false
    }
}

// ── Uploads API ──
async function loadUploads(more = false) {
    uploadsLoading.value = true
    try {
        const opts = {
            limit: 50,
            sort: uploadsSortBy.value,
            order: uploadsSortOrder.value,
        }
        if (uploadsUserFilter.value) opts.user = uploadsUserFilter.value
        if (uploadsTokenFilter.value) opts.token = uploadsTokenFilter.value
        if (more && uploadsCursor.value) opts.after = uploadsCursor.value
        const data = await getAdminUploads(opts)
        if (more) {
            uploads.value = [...uploads.value, ...data.results]
        } else {
            uploads.value = data.results || []
        }
        uploadsCursor.value = data.after || null
    } catch (err) {
        console.error('Failed to load uploads', err)
    } finally {
        uploadsLoading.value = false
    }
}

function filterUploadsByUser(userId) {
    uploadsUserFilter.value = userId
    uploadsTokenFilter.value = ''
    display.value = 'uploads'
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

function filterUploadsByToken(token) {
    uploadsTokenFilter.value = token
    display.value = 'uploads'
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

function clearUserFilter() {
    uploadsUserFilter.value = ''
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

function clearTokenFilter() {
    uploadsTokenFilter.value = ''
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

function changeSortBy(val) {
    uploadsSortBy.value = val
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

function changeSortOrder(val) {
    uploadsSortOrder.value = val
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

// ── User management ──
function openCreateUser() {
    createForm.value = { provider: 'local', login: '', password: '', name: '', email: '', admin: false, maxFileSize: 0, maxUserSize: 0, maxTTL: 0 }
    createTTLUnit.value = 60
    createError.value = ''
    showCreateUser.value = true
}

async function submitCreateUser() {
    createSaving.value = true
    createError.value = ''
    try {
        const payload = buildEditPayload(createForm.value, createTTLUnit.value)
        const u = await apiCreateUser(payload)
        users.value = [u, ...users.value]
        showCreateUser.value = false
    } catch (err) {
        createError.value = err.message || 'Failed to create user'
    } finally {
        createSaving.value = false
    }
}

function openEditUser(user) {
    const { form, ttlUnit } = buildEditForm(user)
    editForm.value = form
    editTTLUnit.value = ttlUnit
    editError.value = ''
    showEditUser.value = true
}

async function submitEditUser() {
    editSaving.value = true
    editError.value = ''
    try {
        const payload = buildEditPayload(editForm.value, editTTLUnit.value)
        const updated = await updateUser(payload)
        const idx = users.value.findIndex(u => u.id === updated.id)
        if (idx >= 0) users.value[idx] = updated
        showEditUser.value = false
    } catch (err) {
        editError.value = err.message || 'Failed to update user'
    } finally {
        editSaving.value = false
    }
}

function handleDeleteUser(user) {
    confirm.value = {
        message: `Delete user "${user.login}" (${user.provider})? All their uploads will remain.`,
        action: async () => {
            try {
                await apiDeleteUser(user.id)
                users.value = users.value.filter(u => u.id !== user.id)
            } catch (err) {
                console.error('Failed to delete user', err)
            }
            confirm.value = null
        }
    }
}

function handleDeleteUpload(upload) {
    confirm.value = {
        message: `Delete upload ${upload.id}?`,
        action: async () => {
            try {
                await removeUpload(upload.id, upload.uploadToken)
                uploads.value = uploads.value.filter(u => u.id !== upload.id)
            } catch (err) {
                console.error('Failed to delete upload', err)
            }
            confirm.value = null
        }
    }
}

// ── Display switching ──
function showStatsView() {
    display.value = 'stats'
    loadStats()
}

function showUsersView() {
    display.value = 'users'
    users.value = []
    loadUsers()
}

function showUploadsView() {
    display.value = 'uploads'
    uploadsUserFilter.value = ''
    uploadsTokenFilter.value = ''
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

// ── Init ──
onMounted(async () => {
    if (!auth.user || !auth.user.admin) {
        router.push('/')
        return
    }
    try {
        version.value = await getVersion()
    } catch (err) {
        console.error('Failed to get version', err)
    }
    loadStats()
})
</script>

<template>
  <div class="w-full max-w-screen-2xl mx-auto px-4 sm:px-6 py-6">
    <div class="flex flex-col md:flex-row gap-6">

      <!-- ═══════ Sidebar ═══════ -->
      <aside class="w-full md:w-72 shrink-0 space-y-4">

        <!-- Server Info Card -->
        <div class="glass-card p-5 text-center space-y-2">
          <p class="text-surface-200 font-medium">Plik Server</p>
          <p v-if="version" class="text-xs text-surface-500 font-mono">
            v{{ version.version }}
          </p>
          <p v-if="version" class="text-xs text-surface-500">
            {{ version.goVersion }}
          </p>
          <div v-if="version" class="flex items-center justify-center gap-2 pt-1">
            <span :class="version.isRelease
              ? 'bg-emerald-500/20 text-emerald-400'
              : 'bg-red-500/20 text-red-400'"
                  class="text-xs px-2 py-0.5 rounded-full">release</span>
            <span :class="version.isMint
              ? 'bg-emerald-500/20 text-emerald-400'
              : 'bg-red-500/20 text-red-400'"
                  class="text-xs px-2 py-0.5 rounded-full">mint</span>
          </div>
        </div>

        <!-- Nav Buttons -->
        <div class="glass-card p-2 space-y-1">
          <button @click="showStatsView"
                  :class="display === 'stats'
                    ? 'bg-accent-500/10 text-accent-400 border-l-2 border-accent-400'
                    : 'text-surface-300 hover:text-white hover:bg-surface-700/50 border-l-2 border-transparent'"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            Stats
          </button>

          <button @click="showUploadsView"
                  :class="display === 'uploads'
                    ? 'bg-accent-500/10 text-accent-400 border-l-2 border-accent-400'
                    : 'text-surface-300 hover:text-white hover:bg-surface-700/50 border-l-2 border-transparent'"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            Uploads
          </button>

          <button @click="showUsersView"
                  :class="display === 'users'
                    ? 'bg-accent-500/10 text-accent-400 border-l-2 border-accent-400'
                    : 'text-surface-300 hover:text-white hover:bg-surface-700/50 border-l-2 border-transparent'"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            Users
          </button>
        </div>

        <!-- Create User -->
        <div class="glass-card p-2">
          <button @click="openCreateUser"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm
                         text-surface-300 hover:text-white hover:bg-surface-700/50 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
            </svg>
            Create User
          </button>
        </div>
      </aside>

      <!-- ═══════ Main Content ═══════ -->
      <main class="flex-1 min-w-0">

        <!-- ─── Stats View ─── -->
        <template v-if="display === 'stats'">
          <div v-if="statsLoading" class="text-center py-12 text-surface-500">Loading stats...</div>

          <div v-else-if="stats" class="space-y-4">
            <!-- Server Config -->
            <div class="glass-card p-5">
              <h3 class="text-sm text-surface-400 uppercase tracking-wider mb-4">Server Configuration</h3>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
                <div>
                  <p class="text-xs text-surface-500">Max File Size</p>
                  <p class="text-surface-200 font-medium">{{ quotaLabel(config.maxFileSize) }}</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Max User Size</p>
                  <p class="text-surface-200 font-medium">{{ quotaLabel(config.maxUserSize) }}</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Default TTL</p>
                  <p class="text-surface-200 font-medium">{{ ttlLabel(config.defaultTTL) }}</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Max TTL</p>
                  <p class="text-surface-200 font-medium">{{ ttlLabel(config.maxTTL) }}</p>
                </div>
              </div>
            </div>

            <!-- Server Stats -->
            <div class="glass-card p-5">
              <h3 class="text-sm text-surface-400 uppercase tracking-wider mb-4">Server Statistics</h3>
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-6 text-center">
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ stats.users }}</p>
                  <p class="text-xs text-surface-500">Users</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ stats.uploads }}</p>
                  <p class="text-xs text-surface-500">Uploads</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ stats.anonymousUploads }}</p>
                  <p class="text-xs text-surface-500">Anonymous Uploads</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ stats.files }}</p>
                  <p class="text-xs text-surface-500">Files</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ humanReadableSize(stats.totalSize) }}</p>
                  <p class="text-xs text-surface-500">Total Size</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ humanReadableSize(stats.anonymousTotalSize) }}</p>
                  <p class="text-xs text-surface-500">Anonymous Size</p>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ─── Users View ─── -->
        <template v-if="display === 'users'">
          <div v-if="usersLoading && users.length === 0" class="text-center py-12 text-surface-500">
            Loading users...
          </div>

          <div v-else-if="users.length === 0" class="text-center py-12 text-surface-500">
            No users yet
          </div>

          <div class="space-y-3">
            <div v-for="user in users" :key="user.id" class="glass-card p-4">
              <div class="flex flex-col sm:flex-row gap-4">
                <!-- User info -->
                <div class="sm:w-1/4 text-sm space-y-1">
                  <p class="text-surface-200 font-medium">{{ user.login }}</p>
                  <p class="text-surface-500">({{ user.provider }})</p>
                  <span v-if="user.admin"
                        class="inline-block text-xs bg-emerald-500/20 text-emerald-400 px-2 py-0.5 rounded-full">
                    admin
                  </span>
                </div>

                <!-- Name / Email -->
                <div class="sm:w-1/4 text-sm space-y-1">
                  <p v-if="user.name" class="text-surface-300">{{ user.name }}</p>
                  <p v-if="user.email" class="text-surface-400">{{ user.email }}</p>
                </div>

                <!-- Quotas -->
                <div class="sm:w-1/4 text-xs text-surface-500 space-y-1">
                  <p>max file size: {{ quotaLabel(user.maxFileSize) }}</p>
                  <p>max user size: {{ quotaLabel(user.maxUserSize) }}</p>
                  <p>max TTL: {{ ttlLabel(user.maxTTL) }}</p>
                </div>

                <!-- Actions -->
                <div class="sm:w-1/4 flex flex-wrap items-center justify-end gap-2">
                  <button @click="doImpersonate(user)"
                          :disabled="user.id === auth.originalUser?.id"
                          :class="user.id === auth.originalUser?.id ? 'opacity-30 cursor-not-allowed' : 'hover:text-green-300 hover:bg-green-500/10'"
                          class="text-xs text-green-400 border border-green-500/30
                                 rounded-lg px-3 py-1.5 transition-colors"
                          title="Impersonate">
                    👤
                  </button>
                  <button @click="openEditUser(user)"
                          class="text-xs text-accent-400 hover:text-accent-300 border border-accent-500/30
                                 rounded-lg px-3 py-1.5 hover:bg-accent-500/10 transition-colors">
                    Edit
                  </button>
                  <button @click="handleDeleteUser(user)"
                          :disabled="user.id === auth.user?.id"
                          :class="user.id === auth.user?.id ? 'opacity-30 cursor-not-allowed' : 'hover:text-red-300 hover:bg-red-500/10'"
                          class="text-xs text-red-400 border border-red-500/30
                                 rounded-lg px-3 py-1.5 transition-colors">
                    Delete
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Load more -->
          <div v-if="usersCursor" class="mt-4">
            <button @click="loadUsers(true)"
                    class="w-full glass-card p-3 text-sm text-surface-400 hover:text-white
                           hover:bg-surface-700/30 transition-colors text-center"
                    :disabled="usersLoading">
              {{ usersLoading ? 'Loading...' : 'Load more users' }}
            </button>
          </div>
        </template>

        <!-- ─── Uploads View ─── -->
        <template v-if="display === 'uploads'">

          <!-- Sort / filter controls -->
          <div class="glass-card p-3 mb-4 space-y-2 text-sm">
            <div class="flex flex-wrap items-center gap-4">
              <!-- Sort by -->
              <div class="flex items-center gap-2 text-surface-400">
                <span>Sort:</span>
                <button @click="changeSortBy('date')"
                        :class="uploadsSortBy === 'date' ? 'text-accent-400' : 'text-surface-500 hover:text-surface-300'"
                        class="transition-colors">Date</button>
                <span class="text-surface-600">|</span>
                <button @click="changeSortBy('size')"
                        :class="uploadsSortBy === 'size' ? 'text-accent-400' : 'text-surface-500 hover:text-surface-300'"
                        class="transition-colors">Size</button>
              </div>
              <!-- Order -->
              <div class="flex items-center gap-2 text-surface-400">
                <span>Order:</span>
                <button @click="changeSortOrder('desc')"
                        :class="uploadsSortOrder === 'desc' ? 'text-accent-400' : 'text-surface-500 hover:text-surface-300'"
                        class="transition-colors">Desc</button>
                <span class="text-surface-600">|</span>
                <button @click="changeSortOrder('asc')"
                        :class="uploadsSortOrder === 'asc' ? 'text-accent-400' : 'text-surface-500 hover:text-surface-300'"
                        class="transition-colors">Asc</button>
              </div>
            </div>

            <!-- Active filters -->
            <div v-if="uploadsUserFilter || uploadsTokenFilter" class="flex flex-wrap items-center gap-3">
              <div v-if="uploadsUserFilter" class="flex items-center gap-1.5 text-surface-300">
                <svg class="w-3.5 h-3.5 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
                </svg>
                user: <span class="font-mono text-accent-400">{{ uploadsUserFilter }}</span>
                <button @click="clearUserFilter" class="text-surface-500 hover:text-white">×</button>
              </div>
              <div v-if="uploadsTokenFilter" class="flex items-center gap-1.5 text-surface-300">
                <svg class="w-3.5 h-3.5 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
                </svg>
                token: <span class="font-mono text-accent-400">{{ uploadsTokenFilter.substring(0, 12) }}...</span>
                <button @click="clearTokenFilter" class="text-surface-500 hover:text-white">×</button>
              </div>
            </div>
          </div>

          <div v-if="uploadsLoading && uploads.length === 0" class="text-center py-12 text-surface-500">
            Loading uploads...
          </div>

          <div v-else-if="uploads.length === 0" class="text-center py-12 text-surface-500">
            No uploads
          </div>

          <div class="space-y-3">
            <div v-for="upload in uploads" :key="upload.id" class="glass-card p-4">
              <div class="flex flex-col sm:flex-row gap-4">
                <!-- Upload meta -->
                <div class="sm:w-1/3 text-sm space-y-1">
                  <a :href="getUploadUrl(upload)"
                     class="font-mono text-accent-400 hover:text-accent-300 transition-colors">
                    {{ upload.id }}
                  </a>
                  <p class="text-surface-500">uploaded: {{ formatDate(upload.createdAt) }}</p>
                  <p class="text-surface-500">expires: {{ formatDate(upload.expireAt) }}</p>
                  <p v-if="upload.user" class="text-surface-500">
                    user:
                    <button @click="filterUploadsByUser(upload.user)"
                            class="text-accent-400 hover:text-accent-300 transition-colors">
                      {{ upload.user }}
                    </button>
                  </p>
                  <p v-if="upload.token" class="text-surface-500">
                    token:
                    <button @click="filterUploadsByToken(upload.token)"
                            class="font-mono text-accent-400/70 hover:text-accent-300 transition-colors">
                      {{ upload.token.substring(0, 8) }}...
                    </button>
                  </p>
                </div>

                <!-- Files -->
                <div class="flex-1 min-w-0 text-sm space-y-1">
                  <div v-for="file in (upload.files || []).filter(f => f.status === 'uploaded')"
                       :key="file.id"
                       class="flex items-center justify-between gap-2">
                    <a :href="getFileURL(upload.id, file.id, file.fileName)"
                       class="text-surface-300 hover:text-white transition-colors truncate">
                      {{ file.fileName }}
                    </a>
                    <span class="text-surface-500 shrink-0">{{ humanReadableSize(file.fileSize) }}</span>
                  </div>
                  <p v-if="!upload.files || upload.files.length === 0"
                     class="text-surface-500 italic">No files</p>
                </div>

                <!-- Actions -->
                <div class="sm:w-20 flex sm:flex-col items-center sm:justify-center gap-2">
                  <button @click="handleDeleteUpload(upload)"
                          class="text-xs text-red-400 hover:text-red-300 border border-red-500/30
                                 rounded-lg px-3 py-1.5 hover:bg-red-500/10 transition-colors">
                    Remove
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Load more -->
          <div v-if="uploadsCursor" class="mt-4">
            <button @click="loadUploads(true)"
                    class="w-full glass-card p-3 text-sm text-surface-400 hover:text-white
                           hover:bg-surface-700/30 transition-colors text-center"
                    :disabled="uploadsLoading">
              {{ uploadsLoading ? 'Loading...' : 'Load more uploads' }}
            </button>
          </div>
        </template>
      </main>
    </div>

    <!-- ═══════ Confirm Dialog ═══════ -->
    <ConfirmDialog v-if="confirm"
                   :message="confirm.message"
                   @confirm="confirm.action()"
                   @cancel="confirm = null" />

    <!-- ═══════ Create User Modal ═══════ -->
    <Teleport to="body">
      <div v-if="showCreateUser"
           class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
           @click.self="showCreateUser = false">
        <div class="glass-card p-6 max-w-md w-full space-y-5 animate-fade-in max-h-[90vh] overflow-y-auto">
          <h2 class="text-lg font-semibold text-surface-200">Create User</h2>

          <div v-if="createError" class="text-sm text-red-400 bg-red-500/10 rounded-lg px-3 py-2">
            {{ createError }}
          </div>

          <!-- Provider -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Provider</label>
            <select v-model="createForm.provider" class="input-field w-full">
              <option value="local">local</option>
              <option value="google">google</option>
              <option value="ovh">ovh</option>
            </select>
          </div>

          <!-- Login -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Login</label>
            <input type="text" v-model="createForm.login" class="input-field w-full" placeholder="min 4 chars" />
          </div>

          <!-- Password (local only) -->
          <div v-if="createForm.provider === 'local'">
            <label class="block text-xs text-surface-500 mb-1">Password</label>
            <input type="password" v-model="createForm.password" class="input-field w-full" placeholder="min 8 chars" />
          </div>

          <!-- Name -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Name</label>
            <input type="text" v-model="createForm.name" class="input-field w-full" placeholder="Optional" />
          </div>

          <!-- Email -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Email</label>
            <input type="email" v-model="createForm.email" class="input-field w-full" placeholder="Optional" />
          </div>

          <!-- Quotas -->
          <div class="border-t border-surface-700/50 pt-4 space-y-4">
            <p class="text-xs text-surface-500 uppercase tracking-wider">Quotas</p>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs text-surface-500 mb-1">Max File Size (GB)</label>
                <input type="number" step="0.1" min="-1" v-model.number="createForm.maxFileSize"
                       @blur="createForm.maxFileSize = clampQuota(createForm.maxFileSize)"
                       class="input-field w-full" />
                <p class="text-xs text-surface-600 mt-0.5">{{ defaultSizeHint(config.maxFileSize) }}</p>
              </div>
              <div>
                <label class="block text-xs text-surface-500 mb-1">Max User Size (GB)</label>
                <input type="number" step="0.1" min="-1" v-model.number="createForm.maxUserSize"
                       @blur="createForm.maxUserSize = clampQuota(createForm.maxUserSize)"
                       class="input-field w-full" />
                <p class="text-xs text-surface-600 mt-0.5">{{ defaultSizeHint(config.maxUserSize) }}</p>
              </div>
            </div>
            <div>
              <label class="block text-xs text-surface-500 mb-1">Max TTL</label>
              <div class="flex gap-2">
                <input type="number" step="1" min="-1" v-model.number="createForm.maxTTL"
                       @blur="createForm.maxTTL = clampQuota(createForm.maxTTL)"
                       class="input-field flex-1" />
                <select v-model.number="createTTLUnit" class="input-field w-28">
                  <option v-for="u in TTL_UNITS" :key="u.seconds" :value="u.seconds">{{ u.label }}</option>
                </select>
              </div>
              <p class="text-xs text-surface-600 mt-0.5">{{ defaultTTLHint(config.maxTTL) }}</p>
            </div>

            <!-- Admin -->
            <label class="flex items-center gap-2 text-sm text-surface-300 cursor-pointer">
              <input type="checkbox" v-model="createForm.admin"
                     class="w-4 h-4 rounded border-surface-600 bg-surface-800
                            text-accent-500 focus:ring-accent-500/30" />
              Admin
            </label>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <button @click="showCreateUser = false" class="btn-ghost text-sm px-4 py-2">Cancel</button>
            <button @click="submitCreateUser" :disabled="createSaving"
                    class="btn-primary px-4 py-2 text-sm">
              {{ createSaving ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ═══════ Edit User Modal ═══════ -->
    <Teleport to="body">
      <div v-if="showEditUser"
           class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
           @click.self="showEditUser = false">
        <div class="glass-card p-6 max-w-md w-full space-y-5 animate-fade-in max-h-[90vh] overflow-y-auto">
          <h2 class="text-lg font-semibold text-surface-200">Edit User</h2>

          <div v-if="editError" class="text-sm text-red-400 bg-red-500/10 rounded-lg px-3 py-2">
            {{ editError }}
          </div>

          <!-- Provider & Login (read-only) -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs text-surface-500 mb-1">Provider</label>
              <div class="input-field bg-surface-800/50 text-surface-400 cursor-not-allowed">{{ editForm.provider }}</div>
            </div>
            <div>
              <label class="block text-xs text-surface-500 mb-1">Login</label>
              <div class="input-field bg-surface-800/50 text-surface-400 cursor-not-allowed">{{ editForm.login }}</div>
            </div>
          </div>

          <div>
            <label class="block text-xs text-surface-500 mb-1">Name</label>
            <input type="text" v-model="editForm.name" class="input-field w-full" placeholder="Display name" />
          </div>

          <div>
            <label class="block text-xs text-surface-500 mb-1">Email</label>
            <input type="email" v-model="editForm.email" class="input-field w-full" placeholder="Email" />
          </div>

          <div v-if="editForm.provider === 'local'">
            <label class="block text-xs text-surface-500 mb-1">Password</label>
            <input type="password" v-model="editForm.password" class="input-field w-full"
                   placeholder="Leave blank to keep current" />
          </div>

          <!-- Quotas -->
          <div class="border-t border-surface-700/50 pt-4 space-y-4">
            <p class="text-xs text-surface-500 uppercase tracking-wider">Quotas</p>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs text-surface-500 mb-1">Max File Size (GB)</label>
                <input type="number" step="0.1" min="-1" v-model.number="editForm.maxFileSize"
                       @blur="editForm.maxFileSize = clampQuota(editForm.maxFileSize)"
                       class="input-field w-full" />
                <p class="text-xs text-surface-600 mt-0.5">{{ defaultSizeHint(config.maxFileSize) }}</p>
              </div>
              <div>
                <label class="block text-xs text-surface-500 mb-1">Max User Size (GB)</label>
                <input type="number" step="0.1" min="-1" v-model.number="editForm.maxUserSize"
                       @blur="editForm.maxUserSize = clampQuota(editForm.maxUserSize)"
                       class="input-field w-full" />
                <p class="text-xs text-surface-600 mt-0.5">{{ defaultSizeHint(config.maxUserSize) }}</p>
              </div>
            </div>
            <div>
              <label class="block text-xs text-surface-500 mb-1">Max TTL</label>
              <div class="flex gap-2">
                <input type="number" step="1" min="-1" v-model.number="editForm.maxTTL"
                       @blur="editForm.maxTTL = clampQuota(editForm.maxTTL)"
                       class="input-field flex-1" />
                <select v-model.number="editTTLUnit" class="input-field w-28">
                  <option v-for="u in TTL_UNITS" :key="u.seconds" :value="u.seconds">{{ u.label }}</option>
                </select>
              </div>
              <p class="text-xs text-surface-600 mt-0.5">{{ defaultTTLHint(config.maxTTL) }}</p>
            </div>

            <label class="flex items-center gap-2 text-sm text-surface-300 cursor-pointer">
              <input type="checkbox" v-model="editForm.admin"
                     class="w-4 h-4 rounded border-surface-600 bg-surface-800
                            text-accent-500 focus:ring-accent-500/30" />
              Admin
            </label>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <button @click="showEditUser = false" class="btn-ghost text-sm px-4 py-2">Cancel</button>
            <button @click="submitEditUser" :disabled="editSaving"
                    class="btn-primary px-4 py-2 text-sm">
              {{ editSaving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
