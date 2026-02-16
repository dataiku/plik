<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { auth, logout } from '../authStore.js'
import { config, isFeatureEnabled } from '../config.js'
import {
    getUserUploads, deleteUserUploads, removeUpload,
    getUserTokens, createToken, revokeToken,
    deleteAccount, getFileURL, updateUser, getUserStatistics
} from '../api.js'
import {
    humanReadableSize, getUploadUrl, quotaLabel, ttlLabel,
    clampQuota, defaultSizeHint, defaultTTLHint, TTL_UNITS,
    formatDate, buildEditForm, buildEditPayload,
} from '../utils.js'
import CopyButton from '../components/CopyButton.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const router = useRouter()

// ── Display mode ──
const display = ref('stats') // 'stats' | 'uploads' | 'tokens'
const tokenFilter = ref(null)

// ── Uploads ──
const uploads = ref([])
const uploadsCursor = ref(null)
const uploadsLoading = ref(false)

// ── Tokens ──
const tokens = ref([])
const tokensCursor = ref(null)
const tokensLoading = ref(false)
const newTokenComment = ref('')

// ── Confirmation state ──
const confirm = ref(null) // { message, action }

// ── Edit account ──
const showEditAccount = ref(false)
const editForm = ref({})
const editSaving = ref(false)
const editError = ref('')

// ── Helpers ──

const editTTLUnit = ref(60)

// ── Stats ──
const userStats = ref(null)
const statsLoading = ref(false)

async function loadUserStats() {
    statsLoading.value = true
    try {
        userStats.value = await getUserStatistics()
    } catch (e) {
        console.warn('Failed to load user stats:', e)
    } finally {
        statsLoading.value = false
    }
}

// Effective default TTL = min(config.defaultTTL, user.maxTTL) when user has a limit
const effectiveDefaultTTL = computed(() => {
    const cfgTTL = config.defaultTTL || 0
    const userMaxTTL = auth.user?.maxTTL || 0
    if (userMaxTTL > 0 && cfgTTL > 0) return Math.min(cfgTTL, userMaxTTL)
    if (userMaxTTL > 0) return userMaxTTL
    return cfgTTL
})

// ── Token lookup map (token UUID → comment) ──
const tokenMap = computed(() => {
    const map = {}
    for (const t of tokens.value) {
        map[t.token] = t.comment || ''
    }
    return map
})

function tokenLabel(tokenStr) {
    const comment = tokenMap.value[tokenStr]
    if (comment) return comment
    return tokenStr.substring(0, 8) + '...'
}

// ── Uploads API ──
async function loadUploads(more = false) {
    uploadsLoading.value = true
    try {
        const opts = { limit: 50 }
        if (tokenFilter.value) opts.token = tokenFilter.value
        if (more && uploadsCursor.value) opts.after = uploadsCursor.value
        const data = await getUserUploads(opts)
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

async function handleDeleteUpload(upload) {
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

async function handleDeleteAllUploads() {
    const label = tokenFilter.value ? `all uploads for token ${tokenFilter.value}` : 'ALL your uploads'
    confirm.value = {
        message: `Delete ${label}? This cannot be undone.`,
        action: async () => {
            try {
                await deleteUserUploads(tokenFilter.value)
                uploads.value = []
                uploadsCursor.value = null
            } catch (err) {
                console.error('Failed to delete uploads', err)
            }
            confirm.value = null
        }
    }
}

function filterByToken(token) {
    tokenFilter.value = token
    uploads.value = []
    uploadsCursor.value = null
    display.value = 'uploads'
    loadUploads()
}

function clearTokenFilter() {
    tokenFilter.value = null
    uploads.value = []
    uploadsCursor.value = null
    loadUploads()
}

// ── Tokens API ──
async function loadTokens(more = false) {
    tokensLoading.value = true
    try {
        const opts = { limit: 50 }
        if (more && tokensCursor.value) opts.after = tokensCursor.value
        const data = await getUserTokens(opts)
        if (more) {
            tokens.value = [...tokens.value, ...data.results]
        } else {
            tokens.value = data.results || []
        }
        tokensCursor.value = data.after || null
    } catch (err) {
        console.error('Failed to load tokens', err)
    } finally {
        tokensLoading.value = false
    }
}

async function handleCreateToken() {
    try {
        const token = await createToken(newTokenComment.value.trim() || undefined)
        tokens.value = [token, ...tokens.value]
        newTokenComment.value = ''
    } catch (err) {
        console.error('Failed to create token', err)
    }
}

async function handleRevokeToken(token) {
    confirm.value = {
        message: `Revoke token ${token.token.substring(0, 8)}...? Uploads created with this token will remain.`,
        action: async () => {
            try {
                await revokeToken(token.token)
                tokens.value = tokens.value.filter(t => t.token !== token.token)
            } catch (err) {
                console.error('Failed to revoke token', err)
            }
            confirm.value = null
        }
    }
}

// ── Account ──
async function handleLogout() {
    await logout()
    router.push('/')
}

async function handleDeleteAccount() {
    confirm.value = {
        message: 'Delete your account and ALL uploads? This cannot be undone.',
        action: async () => {
            try {
                await deleteAccount()
                auth.user = null
                router.push('/')
            } catch (err) {
                console.error('Failed to delete account', err)
            }
            confirm.value = null
        }
    }
}

function openEditAccount() {
    const { form, ttlUnit } = buildEditForm(auth.user)
    editForm.value = form
    editTTLUnit.value = ttlUnit
    editError.value = ''
    showEditAccount.value = true
}

async function saveEditAccount() {
    editSaving.value = true
    editError.value = ''
    try {
        const payload = buildEditPayload(editForm.value, editTTLUnit.value)
        const updated = await updateUser(payload)
        Object.assign(auth.user, updated)
        showEditAccount.value = false
    } catch (err) {
        editError.value = err.message || 'Failed to save'
    } finally {
        editSaving.value = false
    }
}

// ── Display switching ──
function showStats() {
    display.value = 'stats'
    loadUserStats()
}

function showUploads() {
    display.value = 'uploads'
    tokenFilter.value = null
    uploads.value = []
    loadUploads()
}

function showTokens() {
    display.value = 'tokens'
    loadTokens()
}

// ── Init ──
onMounted(() => {
    if (!auth.user) {
        router.push('/login')
        return
    }
    loadUserStats()
    loadTokens()  // needed for token comment lookup map
})
</script>

<template>
  <div class="w-full max-w-screen-2xl mx-auto px-4 sm:px-6 py-6">
    <div class="flex flex-col md:flex-row gap-6">

      <!-- ═══════ Sidebar ═══════ -->
      <aside class="w-full md:w-72 shrink-0 space-y-4">

        <!-- User Info Card -->
        <div class="glass-card p-5 text-center space-y-3">
          <div class="w-14 h-14 rounded-full bg-accent-500/20 flex items-center justify-center mx-auto">
            <svg class="w-7 h-7 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div>
            <p class="text-surface-200 font-medium">{{ auth.user?.login || auth.user?.name }}</p>
            <p class="text-xs text-surface-500">{{ auth.user?.provider }}</p>
            <p v-if="auth.user?.name" class="text-xs text-surface-400 mt-1">{{ auth.user.name }}</p>
            <p v-if="auth.user?.email" class="text-xs text-surface-400">{{ auth.user.email }}</p>
            <span v-if="auth.user?.admin"
                  class="inline-block mt-1 text-xs bg-emerald-500/20 text-emerald-400 px-2 py-0.5 rounded-full">
              admin
            </span>
          </div>
        </div>

        <!-- Nav Buttons -->
        <div class="glass-card p-2 space-y-1">
          <button @click="showStats"
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

          <button @click="showUploads"
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

          <button @click="showTokens"
                  :class="display === 'tokens'
                    ? 'bg-accent-500/10 text-accent-400 border-l-2 border-accent-400'
                    : 'text-surface-300 hover:text-white hover:bg-surface-700/50 border-l-2 border-transparent'"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            Tokens
          </button>
        </div>

        <!-- Account Actions -->
        <div class="glass-card p-2 space-y-1">
          <button @click="handleLogout"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm
                         text-surface-300 hover:text-white hover:bg-surface-700/50 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sign out
          </button>

          <button @click="openEditAccount"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm
                         text-surface-300 hover:text-white hover:bg-surface-700/50 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit account
          </button>

          <button v-if="display === 'uploads'"
                  @click="handleDeleteAllUploads"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm
                         text-red-400/70 hover:text-red-400 hover:bg-red-500/10 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Delete uploads
          </button>

          <button v-if="isFeatureEnabled('delete_account')"
                  @click="handleDeleteAccount"
                  class="w-full py-2.5 rounded-lg flex items-center gap-3 px-3 text-sm
                         text-red-400/70 hover:text-red-400 hover:bg-red-500/10 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            Delete account
          </button>
        </div>
      </aside>

      <!-- ═══════ Main Content ═══════ -->
      <main class="flex-1 min-w-0">

        <!-- ─── Stats View ─── -->
        <template v-if="display === 'stats'">
          <div v-if="statsLoading" class="text-center py-12 text-surface-500">Loading stats...</div>

          <div v-else class="space-y-4">
            <!-- User Configuration -->
            <div class="glass-card p-5">
              <h3 class="text-sm text-surface-400 uppercase tracking-wider mb-4">User Configuration</h3>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
                <div>
                  <p class="text-xs text-surface-500">Max File Size</p>
                  <p class="text-surface-200 font-medium">{{ quotaLabel(auth.user?.maxFileSize) }}</p>
                  <p v-if="!auth.user?.maxFileSize && config.maxFileSize" class="text-xs text-surface-500">({{ quotaLabel(config.maxFileSize) }})</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Max User Size</p>
                  <p class="text-surface-200 font-medium">{{ quotaLabel(auth.user?.maxUserSize) }}</p>
                  <p v-if="!auth.user?.maxUserSize && config.maxUserSize" class="text-xs text-surface-500">({{ quotaLabel(config.maxUserSize) }})</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Default TTL</p>
                  <p class="text-surface-200 font-medium">{{ ttlLabel(effectiveDefaultTTL) }}</p>
                </div>
                <div>
                  <p class="text-xs text-surface-500">Max TTL</p>
                  <p class="text-surface-200 font-medium">{{ ttlLabel(auth.user?.maxTTL) }}</p>
                  <p v-if="!auth.user?.maxTTL && config.maxTTL" class="text-xs text-surface-500">({{ ttlLabel(config.maxTTL) }})</p>
                </div>
              </div>
            </div>

            <!-- User Statistics -->
            <div class="glass-card p-5">
              <h3 class="text-sm text-surface-400 uppercase tracking-wider mb-4">User Statistics</h3>
              <div v-if="userStats" class="grid grid-cols-3 gap-6 text-center">
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ userStats.uploads }}</p>
                  <p class="text-xs text-surface-500">Uploads</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ userStats.files }}</p>
                  <p class="text-xs text-surface-500">Files</p>
                </div>
                <div>
                  <p class="text-2xl font-bold text-surface-200">{{ humanReadableSize(userStats.totalSize) }}</p>
                  <p class="text-xs text-surface-500">Total Size</p>
                </div>
              </div>
              <p v-else class="text-sm text-surface-500 text-center py-2">No stats available</p>
            </div>
          </div>
        </template>

        <!-- ─── Uploads View ─── -->
        <template v-if="display === 'uploads'">

          <!-- Token filter bar -->
          <div v-if="tokenFilter"
               class="glass-card p-3 mb-4 flex items-center justify-between text-sm">
            <div class="flex items-center gap-2 text-surface-300">
              <svg class="w-4 h-4 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
              </svg>
              Token: <span class="font-mono text-accent-400">{{ tokenFilter }}</span>
            </div>
            <button @click="clearTokenFilter" class="text-surface-400 hover:text-white transition-colors">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Loading -->
          <div v-if="uploadsLoading && uploads.length === 0"
               class="text-center py-12 text-surface-500">
            Loading uploads...
          </div>

          <!-- Empty state -->
          <div v-else-if="uploads.length === 0"
               class="text-center py-12 text-surface-500">
            No uploads yet
          </div>

          <!-- Upload cards -->
          <div class="space-y-3">
            <div v-for="upload in uploads" :key="upload.id"
                 class="glass-card p-4">
              <div class="flex flex-col sm:flex-row gap-4">
                <!-- Upload meta -->
                <div class="sm:w-1/3 text-sm space-y-1">
                  <a :href="getUploadUrl(upload)"
                     class="font-mono text-accent-400 hover:text-accent-300 transition-colors">
                    {{ upload.id }}
                  </a>
                  <p class="text-surface-500">uploaded: {{ formatDate(upload.createdAt) }}</p>
                  <p class="text-surface-500">expires: {{ formatDate(upload.expireAt) }}</p>
                  <p v-if="upload.token" class="text-surface-500">
                    token:
                    <button @click="filterByToken(upload.token)"
                            class="text-accent-400 hover:text-accent-300 transition-colors">
                      {{ tokenLabel(upload.token) }}
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

        <!-- ─── Tokens View ─── -->
        <template v-if="display === 'tokens'">

          <!-- Create token -->
          <div class="glass-card p-4 mb-4 space-y-3">
            <p class="text-sm text-surface-400 text-center">
              Tokens authenticate the CLI client. Add them to your
              <span class="font-mono text-surface-300">~/.plikrc</span>
              file.
            </p>
            <div class="flex gap-2">
              <input type="text"
                     v-model="newTokenComment"
                     class="input-field flex-1"
                     placeholder="Comment (optional)"
                     @keyup.enter="handleCreateToken" />
              <button @click="handleCreateToken"
                      class="btn-primary px-4 text-sm whitespace-nowrap">
                Create token
              </button>
            </div>
          </div>

          <!-- Loading -->
          <div v-if="tokensLoading && tokens.length === 0"
               class="text-center py-12 text-surface-500">
            Loading tokens...
          </div>

          <!-- Empty state -->
          <div v-else-if="tokens.length === 0"
               class="text-center py-8 text-surface-500">
            No tokens yet
          </div>

          <!-- Token list -->
          <div class="space-y-2">
            <div v-for="token in tokens" :key="token.token"
                 class="glass-card p-4 flex flex-col sm:flex-row items-start sm:items-center gap-3">
              <!-- Token value -->
              <div class="flex-1 min-w-0 space-y-1">
                <p v-if="token.comment" class="text-sm text-surface-200 truncate">{{ token.comment }}</p>
                <div class="flex items-center gap-2">
                  <button @click="filterByToken(token.token)"
                          class="font-mono text-xs text-accent-400/70 hover:text-accent-300 transition-colors
                                 truncate text-left"
                          :title="'Show uploads for this token'">
                    {{ token.token }}
                  </button>
                  <CopyButton :text="token.token" size="sm" />
                </div>
              </div>
              <!-- Created date -->
              <span class="text-xs text-surface-500 shrink-0">{{ formatDate(token.createdAt) }}</span>
              <!-- Revoke -->
              <button @click="handleRevokeToken(token)"
                      class="text-xs text-red-400 hover:text-red-300 border border-red-500/30
                             rounded-lg px-3 py-1.5 hover:bg-red-500/10 transition-colors shrink-0">
                Revoke
              </button>
            </div>
          </div>

          <!-- Load more -->
          <div v-if="tokensCursor" class="mt-4">
            <button @click="loadTokens(true)"
                    class="w-full glass-card p-3 text-sm text-surface-400 hover:text-white
                           hover:bg-surface-700/30 transition-colors text-center"
                    :disabled="tokensLoading">
              {{ tokensLoading ? 'Loading...' : 'Load more tokens' }}
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
    <!-- ═══════ Edit Account Modal ═══════ -->
    <Teleport to="body">
      <div v-if="showEditAccount"
           class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
           @click.self="showEditAccount = false">
        <div class="glass-card p-6 max-w-md w-full space-y-5 animate-fade-in max-h-[90vh] overflow-y-auto">
          <h2 class="text-lg font-semibold text-surface-200">Edit Account</h2>

          <!-- Error -->
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

          <!-- Name -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Name</label>
            <input type="text" v-model="editForm.name" class="input-field w-full" placeholder="Display name" />
          </div>

          <!-- Email -->
          <div>
            <label class="block text-xs text-surface-500 mb-1">Email</label>
            <input type="email" v-model="editForm.email" class="input-field w-full" placeholder="Email" />
          </div>

          <!-- Password (local only) -->
          <div v-if="editForm.provider === 'local'">
            <label class="block text-xs text-surface-500 mb-1">Password</label>
            <input type="password" v-model="editForm.password" class="input-field w-full"
                   placeholder="Leave blank to keep current" />
          </div>

          <!-- Admin-only fields -->
          <template v-if="auth.user?.admin">
            <div class="border-t border-surface-700/50 pt-4 space-y-4">
              <p class="text-xs text-surface-500 uppercase tracking-wider">Admin Settings</p>

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
          </template>

          <!-- Actions -->
          <div class="flex justify-end gap-2 pt-2">
            <button @click="showEditAccount = false" class="btn-ghost text-sm px-4 py-2">Cancel</button>
            <button @click="saveEditAccount" :disabled="editSaving"
                    class="btn-primary px-4 py-2 text-sm">
              {{ editSaving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
