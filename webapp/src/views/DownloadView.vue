<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getUpload, removeUpload, removeFile as apiRemoveFile, uploadFile, getFileURL } from '../api.js'
import { generateRef } from '../utils.js'
import { getToken, setToken } from '../tokenStore.js'
import { marked } from 'marked'
import DownloadSidebar from '../components/DownloadSidebar.vue'
import FileRow from '../components/FileRow.vue'
import CopyButton from '../components/CopyButton.vue'
import QrCodeDialog from '../components/QrCodeDialog.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import CodeEditor from '../components/CodeEditor.vue'

const props = defineProps({
  id: { type: String, required: true },
})

const router = useRouter()

const upload = ref(null)
const loading = ref(true)
const error = ref(null)
const fileInput = ref(null)

// Staged files pending upload
const pendingFiles = ref([])
const isAddingFiles = ref(false)

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
  }
}

function closeViewer() {
  viewingFile.value = null
  viewingContent.value = ''
  viewingError.value = null
}

// Active (non-removed) files — includes missing, uploading, and uploaded
const activeFiles = computed(() => {
  if (!upload.value?.files) return []
  return upload.value.files.filter(f => f.status !== 'removed' && f.status !== 'deleted')
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

    // If no active files and user is not admin (can't add files), redirect to home
    if (activeFiles.value.length === 0 && !isAdmin.value) {
      router.push({ path: '/' })
      return
    }
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

        // If no active files left after deletion, redirect to home
        if (activeFiles.value.length === 0) {
          router.push({ path: '/' })
        }
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

async function uploadPendingFiles() {
  if (!pendingFiles.value.length || isAddingFiles.value) return
  isAddingFiles.value = true

  for (const fileEntry of pendingFiles.value) {
    fileEntry.status = 'uploading'
    try {
      const result = await uploadFile(
        { id: props.id, stream: upload.value.stream, uploadToken: uploadToken.value },
        { fileName: fileEntry.fileName, file: fileEntry.file },
        (progress) => { fileEntry.progress = progress },
        null,
      )
      fileEntry.status = 'uploaded'
      fileEntry.id = result.id
    } catch (err) {
      fileEntry.status = 'error'
      fileEntry.error = err.message || 'Upload failed'
      error.value = err.message || `Failed to upload ${fileEntry.fileName}`
    }
  }

  // Clear pending files and refresh the upload
  pendingFiles.value = []
  isAddingFiles.value = false
  await fetchUpload()
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

onMounted(() => {
  // If uploadToken is in the URL (from admin URL), save it to memory and strip from URL
  const queryToken = router.currentRoute.value.query.uploadToken
  if (queryToken) {
    setToken(props.id, queryToken)
    router.replace({ path: '/', query: { id: props.id } })
  }
  fetchUpload()
})
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
          <div v-if="viewingFile" class="glass-card overflow-hidden animate-fade-in">
            <div class="flex items-center justify-between border-b border-surface-700/50 px-4 py-2">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
                <span class="text-sm font-medium text-surface-200">{{ viewingFile.fileName }}</span>
              </div>
              <button class="text-surface-400 hover:text-white transition-colors"
                      @click="closeViewer">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="viewingLoading" class="flex items-center justify-center py-8">
              <div class="animate-spin rounded-full h-6 w-6 border-2 border-accent-500 border-t-transparent" />
              <span class="ml-3 text-sm text-surface-400">Loading file content...</span>
            </div>
            <div v-else-if="viewingError" class="p-4 text-sm text-danger-500">{{ viewingError }}</div>
            <div v-else class="p-2">
              <CodeEditor
                :model-value="viewingContent"
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
            </div>

            <FileRow v-for="file in activeFiles"
                     :key="file.id"
                     :file="file"
                     :upload-id="id"
                     mode="download"
                     :can-remove="canRemoveFiles"
                     @remove="deleteFile"
                     @show-qr="openQrFile"
                     @view="viewFile" />
          </div>

          <!-- Pending Files (staged for upload) -->
          <div v-if="pendingFiles.length" class="space-y-2">
            <div class="flex items-center justify-between px-1">
              <h3 class="text-sm font-medium text-surface-400">
                {{ pendingFiles.length }} file{{ pendingFiles.length > 1 ? 's' : '' }} to add
              </h3>
            </div>

            <FileRow v-for="file in pendingFiles"
                     :key="file.reference"
                     :file="file"
                     :mode="isAddingFiles ? 'uploading' : 'upload'"
                     @remove="removePendingFile" />
          </div>

          <!-- Upload Pending Files Button -->
          <div v-if="pendingFiles.length && !isAddingFiles" class="flex justify-end">
            <button class="btn-success px-8 py-3 text-base" @click="uploadPendingFiles">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Upload
            </button>
          </div>

          <!-- Adding Files Spinner -->
          <div v-if="isAddingFiles" class="flex items-center justify-center py-4">
            <div class="animate-spin rounded-full h-6 w-6 border-2 border-accent-500 border-t-transparent" />
            <span class="ml-3 text-sm text-surface-400">Uploading files...</span>
          </div>

          <!-- No files -->
          <div v-if="!activeFiles.length && !pendingFiles.length" class="glass-card p-8 text-center">
            <p class="text-surface-400">No files in this upload</p>
          </div>

          <!-- Download Links Panel -->
          <div v-if="fileLinks().length" class="glass-card p-4 space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider">
                Download Links
              </h3>
              <CopyButton
                :text="fileLinks().map(f => f.url).join('\n')"
                label="Copy All"
                size="sm" />
            </div>
            <div v-for="fl in fileLinks()" :key="fl.id"
                 class="flex items-center gap-2 group">
              <a :href="fl.url"
                 class="text-sm text-accent-400 hover:text-accent-300 transition-colors truncate flex-1">
                {{ fl.fileName }}
              </a>
              <!-- QR code button -->
              <button class="btn bg-surface-700/50 text-surface-400 hover:text-white px-2 py-1 text-xs opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Show QR code"
                      @click="openQrFile(fl)">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4zM17 14v3h3M14 17h3v3" />
                </svg>
              </button>
              <CopyButton :text="fl.url" size="sm" />
            </div>
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
