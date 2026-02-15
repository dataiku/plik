// In-memory store to pass initial files from UploadView → DownloadView across navigation
// Files + basicAuth are stashed after createUpload(), consumed once by DownloadView on mount
const pending = new Map()

export function setPendingFiles(uploadId, files, basicAuth) {
    pending.set(uploadId, { files, basicAuth })
}

export function consumePendingFiles(uploadId) {
    const data = pending.get(uploadId)
    pending.delete(uploadId)
    return data || null
}
