<script setup>
import { ref, onMounted, computed, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getUpload, removeUpload, removeFile as apiRemoveFile, uploadFile, getFileURL } from '../api.js'
import { generateRef, isTextFile } from '../utils.js'
import { getToken, setToken } from '../tokenStore.js'
import { consumePendingFiles } from '../pendingUploadStore.js'
import { marked } from 'marked'
import DownloadSidebar from '../components/DownloadSidebar.vue'
import FileRow from '../components/FileRow.vue'
import CopyButton from '../components/CopyButton.vue'
import QrCodeDialog from '../components/QrCodeDialog.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { defineAsyncComponent } from 'vue'
const CodeEditor = defineAsyncComponent(() => import('../components/CodeEditor.vue'))

const props = defineProps({
  id: { type: String, required: true },
})

const router = useRouter()

const upload = ref(null)
const loading = ref(true)
const error = ref(null)
const uploadError = ref(null)
const fileInput = ref(null)

// Staged files pending upload
const pendingFiles = ref([])
const isAddingFiles = ref(false)

// BasicAuth for password-protected uploads (passed from UploadView via pending store)
let pendingBasicAuth = null

// Track whether uploads were cancelled
let uploadsCancelled = false

// QR code dialog
const showQr = ref(false)
const qrTitle = ref('')
const qrUrl = ref('')

// Confirmation dialog state
const confirmDialog = ref(null)

// File viewer state
const viewingFile = ref(null)
const viewingContent = ref('')
const viewingLoading = ref(false)
const viewingError = ref(null)
const lastAutoViewedId = ref(null)

async function viewFile(file) {
  // If already viewing this file, close it
  if (viewingFile.value?.id === file.id) {
    closeViewer()
    return
  }
  viewingFile.value = file
  viewingContent.value = ''
  viewingLoading.value = true
  viewingError.value = null
  try {
    const url = getFileURL(props.id, file.id, file.fileName)
    const resp = await fetch(url, { credentials: 'same-origin' })
    if (!resp.ok) throw new Error(`Failed to fetch file (${resp.status})`)
    const text = await resp.text()
    viewingContent.value = text
  } catch (err) {
    viewingError.value = err.message || 'Failed to load file content'
  } finally {
    viewingLoading.value = false
    nextTick(() => {
      document.getElementById('file-viewer-panel')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    })
  }
}

function closeViewer() {
  viewingFile.value = null
  viewingContent.value = ''
  viewingError.value = null
}

// Active files for the top panel
// During uploads, only show files the user can interact with:
//  - Non-streaming: only 'uploaded' (file complete on server)
//  - Streaming: 'uploading' + 'uploaded' (download works via live stream)
// When not uploading (e.g. friend viewing), show all non-removed files
const activeFiles = computed(() => {
  if (!upload.value?.files) return []
  return upload.value.files.filter(f => {
    if (f.status === 'removed' || f.status === 'deleted') return false
    if (isAddingFiles.value) {
      if (upload.value.stream) {
        return f.status === 'uploading' || f.status === 'uploaded'
      } else {
        return f.status === 'uploaded'
      }
    }
    return true
  })
})

// Upload token from in-memory store (set after upload or from admin URL)
const uploadToken = computed(() => getToken(props.id))

// Check if user is upload admin
const isAdmin = computed(() => upload.value?.admin || false)
const canRemoveFiles = computed(() =>
  upload.value?.removable || upload.value?.admin
)

async function fetchUpload() {
  loading.value = true
  error.value = null
  try {
    upload.value = await getUpload(props.id, uploadToken.value)
  } catch (err) {
    error.value = err.message || 'Failed to load upload'
  } finally {
    loading.value = false
  }
}

async function deleteUpload() {
  confirmDialog.value = {
    title: 'Delete Upload',
    message: 'Are you sure you want to delete this upload? This action cannot be undone.',
    confirmText: 'Delete',
    onConfirm: async () => {
      try {
        await removeUpload(props.id, uploadToken.value)
        // Redirect to home page
        router.push({ path: '/' })
      } catch (err) {
        error.value = err.message || 'Failed to delete upload'
        confirmDialog.value = null
      } finally {
        confirmDialog.value = null
      }
    }
  }
}

async function deleteFile(file) {
  confirmDialog.value = {
    title: 'Delete File',
    message: `Are you sure you want to delete "${file.fileName}"? This action cannot be undone.`,
    confirmText: 'Delete',
    onConfirm: async () => {
      try {
        await apiRemoveFile(
          { id: props.id, stream: upload.value.stream, uploadToken: uploadToken.value },
          file,
        )
        await fetchUpload()
      } catch (err) {
        error.value = err.message || 'Failed to delete file'
        confirmDialog.value = null
      } finally {
        confirmDialog.value = null
      }
    }
  }
}

function triggerAddFiles() {
  fileInput.value?.click()
}

function onFilesSelected(event) {
  const selectedFiles = Array.from(event.target.files)
  event.target.value = ''

  const existingNames = new Set(pendingFiles.value.map(f => f.fileName))
  for (const file of selectedFiles) {
    if (existingNames.has(file.name)) continue
    existingNames.add(file.name)
    pendingFiles.value.push({
      reference: generateRef(),
      fileName: file.name,
      size: file.size,
      file: file,
      status: 'toUpload',
      progress: 0,
    })
  }
}

function removePendingFile(file) {
  pendingFiles.value = pendingFiles.value.filter(f => f.reference !== file.reference)
}

async function cancelFileUpload(file) {
  if (file.abort) {
    file.abort()
  }
  pendingFiles.value = pendingFiles.value.filter(f => f.reference !== file.reference)

  // Give the server time to clean up the aborted file (uploading → removed → deleted)
  await new Promise(resolve => setTimeout(resolve, 200))
  await fetchUpload()
}

function cancelAllUploads() {
  uploadsCancelled = true
  for (const file of pendingFiles.value) {
    if (file.abort) {
      file.abort()
    }
  }
  pendingFiles.value = []
  isAddingFiles.value = false
}

async function uploadPendingFiles() {
  if (!pendingFiles.value.length || isAddingFiles.value) return
  isAddingFiles.value = true
  uploadsCancelled = false

  const basicAuth = pendingBasicAuth
  pendingBasicAuth = null

  const isStream = upload.value.stream
  const filesToUpload = [...pendingFiles.value]
    .filter(fileEntry => pendingFiles.value.includes(fileEntry))

  // Locally update a server file's status (reactive, no full refresh needed)
  function setServerFileStatus(fileId, status) {
    const serverFile = upload.value?.files?.find(f => f.id === fileId)
    if (serverFile) serverFile.status = status
  }

  // Upload a single file entry
  function doUpload(fileEntry) {
    if (uploadsCancelled) return Promise.resolve()

    fileEntry.status = 'uploading'

    const { promise, abort } = uploadFile(
      { id: props.id, stream: isStream, uploadToken: uploadToken.value },
      { id: fileEntry.id, fileName: fileEntry.fileName, file: fileEntry.file },
      (progress) => { fileEntry.progress = progress },
      basicAuth,
      // For streaming: mark server file as 'uploading' on connection open
      // so it appears in the top panel with a working download link
      isStream ? () => setServerFileStatus(fileEntry.id, 'uploading') : undefined,
    )

    fileEntry.abort = abort

    return promise.then((result) => {
      fileEntry.status = 'uploaded'
      fileEntry.id = result.id
      // Locally mark server file as uploaded — appears in top panel
      setServerFileStatus(result.id, 'uploaded')
    }).catch((err) => {
      if (!err.cancelled) {
        fileEntry.status = 'error'
        fileEntry.error = err.message || 'Upload failed'
        uploadError.value = err.message || `Failed to upload ${fileEntry.fileName}`
        // No local status change on error — file stays in bottom panel
      }
    })
  }

  // Concurrency-limited pool: max 5 uploads at a time
  const MAX_CONCURRENT = 5
  const queue = [...filesToUpload]
  const workers = Array.from({ length: Math.min(MAX_CONCURRENT, queue.length) }, async () => {
    while (queue.length > 0 && !uploadsCancelled) {
      const fileEntry = queue.shift()
      await doUpload(fileEntry)
    }
  })

  await Promise.allSettled(workers)

  // Clear completed/errored pending files, keep only remaining toUpload files
  pendingFiles.value = pendingFiles.value.filter(f => f.status === 'toUpload')
  isAddingFiles.value = false

  // Final refresh to sync with server truth
  if (!uploadsCancelled) {
    await fetchUpload()
  }
}

// File download links
function fileLinks() {
  if (!upload.value?.files) return []
  return upload.value.files
    .filter(f => f.status === 'uploaded')
    .map(f => ({
      ...f,
      url: getFileURL(props.id, f.id, f.fileName),
    }))
}

// QR code helpers
function openQrUpload() {
  qrTitle.value = 'Upload Link'
  qrUrl.value = window.location.href
  showQr.value = true
}

function openQrFile(file) {
  qrTitle.value = file.fileName
  qrUrl.value = getFileURL(props.id, file.id, file.fileName)
  showQr.value = true
}

onMounted(async () => {
  // If uploadToken is in the URL (from admin URL), save it to memory and strip from URL
  const queryToken = router.currentRoute.value.query.uploadToken
  if (queryToken) {
    setToken(props.id, queryToken)
    router.replace({ path: '/', query: { id: props.id } })
  }

  await fetchUpload()

  // Consume pending files from UploadView (if any)
  const pending = consumePendingFiles(props.id)
  if (pending) {
    pendingBasicAuth = pending.basicAuth
    pendingFiles.value = pending.files
    // Auto-start uploading
    uploadPendingFiles()
  }
})

// Auto-show view panel if there is exactly one text file
watch(activeFiles, (files) => {
  if (files.length === 1) {
    const file = files[0]
    if (file.status === 'uploaded' && isTextFile(file) && lastAutoViewedId.value !== file.id) {
      lastAutoViewedId.value = file.id
      viewFile(file)
    }
  } else if (files.length > 1) {
    // If more files added, we might want to reset scroll but keep viewer if it was manually opened.
    // However, the requirement is only for "if upload has only one file".
  }
}, { immediate: true })
</script>

<template>
  <div class="flex justify-center flex-1 min-h-0 overflow-x-hidden">
    <div class="flex flex-col md:flex-row flex-1 max-w-screen-2xl px-4 sm:px-6 min-h-0 overflow-hidden">
      <!-- Sidebar -->
      <DownloadSidebar
        v-if="upload"
        :upload="{ ...upload, admin: isAdmin }"
        @delete-upload="deleteUpload"
        @add-files="triggerAddFiles"
        @show-qr="openQrUpload" />

      <!-- Loading placeholder sidebar -->
      <aside v-else class="w-full md:w-72 md:shrink-0 p-4">
        <div class="sidebar-section animate-pulse">
          <div class="h-4 bg-surface-700 rounded w-1/2 mb-3" />
          <div class="h-8 bg-surface-700 rounded mb-2" />
          <div class="h-8 bg-surface-700 rounded" />
        </div>
      </aside>

      <!-- Main Content -->
      <main class="flex-1 py-4 md:pl-4 md:pr-0 overflow-y-auto">
      <div class="space-y-4">
        <!-- Loading -->
        <div v-if="loading" class="flex flex-col items-center justify-center py-16">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-accent-500 border-t-transparent" />
          <span class="mt-4 text-sm text-surface-400">Loading upload...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error"
             class="glass-card border-danger-500/50 p-6 text-center animate-fade-in">
          <svg class="w-12 h-12 text-danger-500 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-danger-500 font-medium">{{ error }}</p>
          <button class="btn-ghost mt-4" @click="fetchUpload">Try again</button>
        </div>

        <!-- Upload Content -->
        <template v-else-if="upload">
          <!-- Inline Error Banner (for errors during file upload) -->
          <div v-if="uploadError"
               class="glass-card border-danger-500/50 p-4 flex items-center gap-3 animate-fade-in">
            <svg class="w-5 h-5 text-danger-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span class="text-sm text-danger-500">{{ uploadError }}</span>
            <button class="ml-auto text-surface-400 hover:text-white" @click="uploadError = null">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <!-- Comment -->
          <div v-if="upload.comments" class="glass-card p-4 animate-fade-in">
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
              </svg>
              <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider">Comment</h3>
            </div>
            <div class="prose prose-sm max-w-none" v-html="marked.parse(upload.comments, { breaks: true })" />
          </div>

          <!-- File Viewer -->
          <div v-if="viewingFile" id="file-viewer-panel" class="glass-card overflow-hidden animate-fade-in">
            <div class="flex items-center justify-between border-b border-surface-700/50 px-4 py-2">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
                <span class="text-sm font-medium text-surface-200">{{ viewingFile.fileName }}</span>
              </div>
              <div class="flex items-center gap-2">
                <CopyButton v-if="viewingContent" :text="viewingContent" label="Copy" />
                <button class="text-surface-400 hover:text-white transition-colors"
                        @click="closeViewer">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div v-if="viewingLoading" class="flex items-center justify-center py-8">
              <div class="animate-spin rounded-full h-6 w-6 border-2 border-accent-500 border-t-transparent" />
              <span class="ml-3 text-sm text-surface-400">Loading file content...</span>
            </div>
            <div v-else-if="viewingError" class="p-4 text-sm text-danger-500">{{ viewingError }}</div>
            <div v-else class="p-2">
              <CodeEditor
                v-model="viewingContent"
                :filename="viewingFile.fileName"
                :readonly="true"
              />
            </div>
          </div>

          <!-- File List -->
          <div v-if="activeFiles.length" class="space-y-2">
            <div class="flex items-center justify-between px-1">
              <h3 class="text-sm font-medium text-surface-400">
                {{ activeFiles.length }} file{{ activeFiles.length > 1 ? 's' : '' }}
              </h3>
              <CopyButton
                v-if="fileLinks().length > 1"
                :text="fileLinks().map(f => f.url).join('\n')"
                label="Copy All Links"
                size="sm" />
            </div>

            <FileRow v-for="file in activeFiles"
                     :key="file.id"
                     :file="file"
                     :upload-id="id"
                     mode="download"
                     :can-remove="canRemoveFiles"
                     :is-stream="upload.stream"
                     @remove="deleteFile"
                     @show-qr="openQrFile"
                     @view="viewFile" />
          </div>

          <!-- Pending Files (staged for upload / uploading) -->
          <div v-if="pendingFiles.length" class="space-y-2">
            <div class="flex items-center justify-between px-1">
              <h3 class="text-sm font-medium text-surface-400">
                {{ pendingFiles.length }} file{{ pendingFiles.length > 1 ? 's' : '' }}
                {{ isAddingFiles ? 'uploading' : 'to add' }}
              </h3>
              <button v-if="isAddingFiles"
                      class="text-xs text-danger-500 hover:text-danger-400 transition-colors"
                      @click="cancelAllUploads">
                Cancel All
              </button>
            </div>

            <FileRow v-for="file in pendingFiles"
                     :key="file.reference"
                     :file="file"
                     :mode="isAddingFiles ? 'uploading' : 'upload'"
                     @remove="isAddingFiles ? cancelFileUpload(file) : removePendingFile(file)"
                     @cancel="cancelFileUpload" />
          </div>

          <!-- Upload Pending Files Button (only shown when files are staged but not yet uploading) -->
          <div v-if="pendingFiles.length && !isAddingFiles" class="flex justify-end">
            <button class="btn-success px-8 py-3 text-base" @click="uploadPendingFiles">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Upload
            </button>
          </div>

          <!-- Upload progress indicator -->
          <div v-if="isAddingFiles" class="flex items-center justify-center py-2">
            <div class="animate-spin rounded-full h-4 w-4 border-2 border-accent-500 border-t-transparent" />
            <span class="ml-2 text-xs text-surface-400">Uploading files...</span>
          </div>

          <!-- No files -->
          <div v-if="!activeFiles.length && !pendingFiles.length" class="glass-card p-8 text-center">
            <p class="text-surface-400">No files in this upload</p>
          </div>


        </template>
      </div>
    </main>

    <!-- Hidden file input for adding files -->
    <input ref="fileInput"
           type="file"
           multiple
           class="hidden"
           @change="onFilesSelected" />

    <!-- QR Code Dialog -->
    <QrCodeDialog v-if="showQr"
                  :title="qrTitle"
                  :url="qrUrl"
                  @close="showQr = false" />

    <!-- Confirm Dialog -->
    <ConfirmDialog v-if="confirmDialog"
                   :title="confirmDialog.title"
                   :message="confirmDialog.message"
                   :confirm-text="confirmDialog.confirmText"
                   @confirm="confirmDialog.onConfirm"
                   @cancel="confirmDialog = null" />
    </div>
  </div>
</template>
