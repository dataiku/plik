<script setup>
import { ref, computed } from 'vue'
import { humanReadableSize, isTextFile as checkIsTextFile } from '../utils.js'
import { getFileURL } from '../api.js'
import CopyButton from './CopyButton.vue'

const props = defineProps({
  file: { type: Object, required: true },
  uploadId: { type: String, default: '' },
  mode: { type: String, default: 'upload' }, // 'upload' | 'uploading' | 'download'
  canRemove: { type: Boolean, default: false },
})

const emit = defineEmits(['remove', 'update-name', 'show-qr', 'view'])

const isTextFile = computed(() => {
  if (props.file.status !== 'uploaded') return false
  return checkIsTextFile(props.file)
})

const showDetails = ref(false)

function onNameInput(e) {
  emit('update-name', e.target.textContent.trim())
}

function fileUrl() {
  if (!props.uploadId || !props.file.id) return ''
  return getFileURL(props.uploadId, props.file.id, props.file.fileName)
}
</script>

<template>
  <div class="file-row animate-fade-in flex-wrap">
    <div class="flex flex-wrap items-center gap-2 md:gap-3 flex-1 min-w-0">
      <!-- File icon -->
      <div class="shrink-0">
        <svg class="w-5 h-5 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
        </svg>
      </div>

      <!-- File name -->
      <div class="flex-1 min-w-0">
        <!-- Editable name (upload mode) -->
        <div v-if="mode === 'upload'"
             class="text-sm text-surface-100 truncate cursor-text outline-none
                    hover:text-white focus:ring-1 focus:ring-accent-500/50 rounded px-1 -mx-1"
             contenteditable="true"
             @blur="onNameInput"
             @keydown.enter.prevent="$event.target.blur()">
          {{ file.fileName }}
        </div>

        <!-- Clickable name → toggle details (download mode) -->
        <button v-else-if="mode === 'download'"
                class="text-sm text-surface-100 truncate hover:text-accent-400 transition-colors text-left w-full"
                @click="showDetails = !showDetails">
          <span class="inline-flex items-center gap-1">
            <svg class="w-3 h-3 text-surface-500 transition-transform duration-200"
                 :class="showDetails ? 'rotate-90' : ''"
                 fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
            {{ file.fileName }}
          </span>
        </button>

        <!-- Static name -->
        <div v-else class="text-sm text-surface-100 truncate">
          {{ file.fileName }}
        </div>

        <!-- Progress bar (uploading mode) -->
        <div v-if="mode === 'uploading' && file.status === 'uploading'" class="mt-1.5">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: (file.progress || 0) + '%' }" />
          </div>
          <span class="text-xs text-surface-400 mt-0.5">{{ file.progress || 0 }}%</span>
        </div>

        <!-- Upload complete indicator -->
        <div v-if="mode === 'uploading' && file.status === 'uploaded'" class="mt-1">
          <span class="text-xs text-success-500 flex items-center gap-1">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            Uploaded
          </span>
        </div>
      </div>

      <!-- File size -->
      <div class="text-sm text-surface-400 shrink-0 tabular-nums">
        {{ humanReadableSize(file.fileSize || file.size) }}
      </div>

      <!-- Status badge for non-uploaded files (download mode) -->
      <span v-if="mode === 'download' && file.status === 'missing'"
            class="text-xs text-warning-500 bg-warning-500/10 px-2 py-0.5 rounded-full shrink-0">
        Waiting for upload
      </span>
      <span v-if="mode === 'download' && file.status === 'uploading'"
            class="text-xs text-accent-400 bg-accent-500/10 px-2 py-0.5 rounded-full shrink-0 inline-flex items-center gap-1">
        <div class="animate-spin rounded-full h-3 w-3 border border-accent-400 border-t-transparent" />
        Uploading…
      </span>

      <!-- Actions -->
      <div class="flex items-center gap-1 shrink-0">

        <!-- QR Code button (download mode) -->
        <button v-if="mode === 'download' && file.status === 'uploaded'"
                class="btn bg-surface-700/50 text-surface-400 hover:text-white px-2 py-1.5 text-xs"
                title="Show QR code"
                @click="emit('show-qr', file)">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4zM17 14v3h3M14 17h3v3" />
          </svg>
        </button>

        <!-- Copy link (download mode) -->
        <CopyButton v-if="mode === 'download' && file.status === 'uploaded'"
                    :text="fileUrl()" />

        <!-- View button (download mode, text files only) -->
        <button v-if="mode === 'download' && file.status === 'uploaded' && isTextFile"
                class="btn bg-accent-500/10 text-accent-400 hover:bg-accent-500/20 px-2 py-1.5 text-xs"
                title="View file content"
                @click="emit('view', file)">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <span class="hidden md:inline">View</span>
        </button>

        <!-- Download button (download mode) -->
        <a v-if="mode === 'download' && file.status === 'uploaded'"
           :href="fileUrl() + '?dl=1'"
           class="btn bg-success-500/10 text-success-500 hover:bg-success-500/20 px-2 md:px-3 py-1.5 text-xs">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          <span class="hidden md:inline">Download</span>
        </a>

        <!-- Remove button -->
        <button v-if="(mode === 'upload' || canRemove) && file.status !== 'uploading'"
                class="btn bg-danger-500/10 text-danger-500 hover:bg-danger-500/20 px-2 py-1.5 text-xs"
                title="Remove file"
                @click="emit('remove', file)">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Expanded file details -->
    <div v-if="showDetails && mode === 'download'"
         class="w-full mt-2 pt-2 border-t border-surface-700/50 text-xs text-surface-400 space-y-1 animate-fade-in">
      <div v-if="file.fileType" class="flex gap-2">
        <span class="text-surface-500 w-14">Type:</span>
        <span class="text-surface-300">{{ file.fileType }}</span>
      </div>
      <div v-if="file.fileMd5" class="flex gap-2">
        <span class="text-surface-500 w-14">MD5:</span>
        <span class="text-surface-300 font-mono">{{ file.fileMd5 }}</span>
      </div>
      <div v-if="file.createdAt" class="flex gap-2">
        <span class="text-surface-500 w-14">Created:</span>
        <span class="text-surface-300">{{ new Date(file.createdAt).toLocaleString() }}</span>
      </div>
    </div>
  </div>
</template>
