<script setup>
import { computed } from 'vue'
import { formatDate } from '../utils.js'
import { getArchiveURL, getAdminURL } from '../api.js'
import CopyButton from './CopyButton.vue'

const props = defineProps({
  upload: { type: Object, required: true },
})

const emit = defineEmits(['delete-upload', 'add-files', 'show-qr'])

const expirationText = computed(() => {
  if (!props.upload.expireAt) return null
  const d = new Date(props.upload.expireAt)
  const now = new Date()
  if (d <= now) return 'Expired'
  const diffMs = d - now
  const diffDays = Math.floor(diffMs / 86400000)
  const diffHours = Math.floor((diffMs % 86400000) / 3600000)
  if (diffDays > 0) return `Expires in ${diffDays}d ${diffHours}h`
  return `Expires in ${diffHours}h`
})

const archiveUrl = computed(() => getArchiveURL(props.upload.id))

const adminUrl = computed(() => {
  if (!props.upload.admin || !props.upload.uploadToken) return null
  return getAdminURL(props.upload.id, props.upload.uploadToken)
})

// Admins can delete upload, or if upload is marked as removable
const canDeleteUpload = computed(() => props.upload.admin || props.upload.removable)
const canAddFiles = computed(() => props.upload.admin && !props.upload.stream)
</script>

<template>
  <aside class="w-full md:w-72 md:shrink-0 p-4 space-y-3 animate-slide-in">
    <!-- Upload Info -->
    <div class="sidebar-section">
      <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider mb-2">Upload Info</h3>

      <div v-if="expirationText" class="text-sm text-surface-300">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-warning-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ expirationText }}
        </div>
        <p class="text-xs text-surface-500 mt-1">{{ formatDate(upload.expireAt) }}</p>
      </div>

      <!-- Upload options badges -->
      <div class="flex flex-wrap gap-1.5 mt-2">
        <span v-if="upload.oneShot"
              class="text-xs px-2 py-0.5 rounded-full bg-warning-500/15 text-warning-500">
          One-shot
        </span>
        <span v-if="upload.removable"
              class="text-xs px-2 py-0.5 rounded-full bg-danger-500/15 text-danger-500">
          Removable
        </span>
        <span v-if="upload.stream"
              class="text-xs px-2 py-0.5 rounded-full bg-accent-500/15 text-accent-400">
          Stream
        </span>
        <span v-if="upload.protectedByPassword"
              class="text-xs px-2 py-0.5 rounded-full bg-surface-600/50 text-surface-300">
          🔒 Password
        </span>
      </div>
    </div>

    <!-- Admin URL (only for admins) -->
    <div v-if="adminUrl" class="sidebar-section">
      <div class="flex items-center gap-1 mb-2">
        <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider">Admin URL</h3>
        <div class="group relative">
          <svg class="w-3.5 h-3.5 text-surface-500 cursor-help" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div class="absolute left-0 -top-2 -translate-y-full hidden group-hover:block w-56 p-2 text-xs bg-surface-800 text-surface-200 rounded shadow-lg z-10">
            Share this URL with others to allow them to add files to this upload
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 p-2 rounded bg-surface-800/50 min-w-0 overflow-hidden">
        <span class="text-xs text-surface-300 truncate flex-1">{{ adminUrl }}</span>
        <CopyButton :text="adminUrl" size="sm" />
      </div>
    </div>

    <!-- Actions -->
    <div class="sidebar-section space-y-2">
      <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider mb-2">Actions</h3>

      <!-- Zip archive -->
      <a v-if="upload.files?.length && !upload.stream"
         :href="archiveUrl"
         class="btn-primary w-full">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
        </svg>
        Zip Archive
      </a>

      <!-- QR Code -->
      <button class="btn-primary w-full" @click="emit('show-qr')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4zM17 14v3h3M14 17h3v3" />
        </svg>
        QR Code
      </button>

      <!-- Add files -->
      <button v-if="canAddFiles"
              class="btn-primary w-full"
              @click="emit('add-files')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Files
      </button>

      <!-- Delete upload -->
      <button v-if="canDeleteUpload"
              class="btn-danger w-full"
              @click="emit('delete-upload')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
        Delete Upload
      </button>
    </div>
  </aside>
</template>
