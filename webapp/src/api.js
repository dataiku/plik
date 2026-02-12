// Plik API Service
// Thin fetch() wrappers for all Plik server endpoints

const base = window.location.origin + window.location.pathname.replace(/\/$/, '')

// Impersonate state — set by authStore, injected in every request
let _impersonateUserId = null

export function setImpersonateUser(userId) {
    _impersonateUserId = userId || null
}

// Read the XSRF token from the plik-xsrf cookie
function getXsrfToken() {
    const match = document.cookie.match(/(?:^|;\s*)plik-xsrf=([^;]+)/)
    return match ? match[1] : ''
}

async function apiCall(url, method = 'GET', data = null, headers = {}) {
    const opts = {
        method,
        credentials: 'same-origin',
        headers: {
            'X-ClientApp': 'web_client',
            ...headers,
        },
    }

    // Impersonate header for admin user switching
    if (_impersonateUserId) {
        opts.headers['X-Plik-Impersonate'] = _impersonateUserId
    }

    // XSRF protection: send token on mutating requests
    if (method !== 'GET' && method !== 'HEAD') {
        const xsrf = getXsrfToken()
        if (xsrf) opts.headers['X-XSRFToken'] = xsrf
    }

    if (data && method !== 'GET') {
        opts.headers['Content-Type'] = 'application/json'
        opts.body = JSON.stringify(data)
    }

    const resp = await fetch(url, opts)

    if (!resp.ok) {
        let message = 'Unknown error'
        try {
            const text = await resp.text()
            try {
                const body = JSON.parse(text)
                message = body.message || body || message
            } catch {
                message = text || message
            }
        } catch {
            // Body unreadable, use default message
        }
        throw { status: resp.status, message }
    }

    // Try to parse as JSON first, fall back to text
    const text = await resp.text()
    if (!text) return null

    try {
        return JSON.parse(text)
    } catch {
        // If not valid JSON (like "ok" from DELETE), return the text
        return text
    }
}

// ── Config & Version ──

export function getConfig() {
    return apiCall(`${base}/config`)
}

export function getVersion() {
    return apiCall(`${base}/version`)
}

// ── Authentication ──

export function login(loginName, password) {
    return apiCall(`${base}/auth/local/login`, 'POST', { login: loginName, password })
}

export function oidcLogin() {
    return apiCall(`${base}/auth/oidc/login`)
}

export function logout() {
    return apiCall(`${base}/auth/logout`, 'GET')
}

export function getUser() {
    return apiCall(`${base}/me`)
}

export function deleteAccount() {
    return apiCall(`${base}/me`, 'DELETE')
}

export function getUserStatistics() {
    return apiCall(`${base}/me/stats`)
}

export function updateUser(user) {
    return apiCall(`${base}/user/${encodeURIComponent(user.id)}`, 'POST', user)
}

// ── Admin ──

export function getServerStats() {
    return apiCall(`${base}/stats`)
}

export function getAdminUsers({ after, limit } = {}) {
    const params = new URLSearchParams()
    if (after) params.set('after', after)
    if (limit) params.set('limit', limit)
    const qs = params.toString()
    return apiCall(`${base}/users${qs ? '?' + qs : ''}`)
}

export function createUser(userData) {
    return apiCall(`${base}/user`, 'POST', userData)
}

export function deleteUser(userId) {
    return apiCall(`${base}/user/${encodeURIComponent(userId)}`, 'DELETE')
}

export function getAdminUploads({ user, token, sort, order, after, limit } = {}) {
    const params = new URLSearchParams()
    if (user) params.set('user', user)
    if (token) params.set('token', token)
    if (sort) params.set('sort', sort)
    if (order) params.set('order', order)
    if (after) params.set('after', after)
    if (limit) params.set('limit', limit)
    const qs = params.toString()
    return apiCall(`${base}/uploads${qs ? '?' + qs : ''}`)
}

// ── User Uploads ──

export function getUserUploads({ token, after, order, limit } = {}) {
    const params = new URLSearchParams()
    if (token) params.set('token', token)
    if (after) params.set('after', after)
    if (order) params.set('order', order)
    if (limit) params.set('limit', String(limit))
    const qs = params.toString()
    return apiCall(`${base}/me/uploads${qs ? '?' + qs : ''}`)
}

export function deleteUserUploads(token) {
    const qs = token ? `?token=${encodeURIComponent(token)}` : ''
    return apiCall(`${base}/me/uploads${qs}`, 'DELETE')
}

// ── Tokens ──

export function getUserTokens({ after, limit } = {}) {
    const params = new URLSearchParams()
    if (after) params.set('after', after)
    if (limit) params.set('limit', String(limit))
    const qs = params.toString()
    return apiCall(`${base}/me/token${qs ? '?' + qs : ''}`)
}

export function createToken(comment) {
    return apiCall(`${base}/me/token`, 'POST', comment ? { comment } : {})
}

export function revokeToken(tokenStr) {
    return apiCall(`${base}/me/token/${tokenStr}`, 'DELETE')
}

// ── Upload CRUD ──

export function createUpload(params) {
    return apiCall(`${base}/upload`, 'POST', params)
}

export function getUpload(id, uploadToken) {
    const headers = {}
    if (uploadToken) headers['X-UploadToken'] = uploadToken
    return apiCall(`${base}/upload/${id}`, 'GET', null, headers)
}

export function removeUpload(id, uploadToken) {
    const headers = {}
    if (uploadToken) headers['X-UploadToken'] = uploadToken
    return apiCall(`${base}/upload/${id}`, 'DELETE', null, headers)
}

// ── File Operations ──

/**
 * Upload a file using XMLHttpRequest (for progress tracking)
 * @param {Object} upload - The upload object { id, stream, uploadToken }
 * @param {Object} fileEntry - { id, fileName, file (File object) }
 * @param {Function} onProgress - callback(percent)
 * @param {string} basicAuth - optional base64 auth string
 * @returns {Promise<Object>} - resolved file metadata
 */
export function uploadFile(upload, fileEntry, onProgress, basicAuth) {
    return new Promise((resolve, reject) => {
        const mode = upload.stream ? 'stream' : 'file'
        let url
        if (fileEntry.id) {
            url = `${base}/${mode}/${upload.id}/${fileEntry.id}/${fileEntry.fileName}`
        } else {
            // Adding file to existing upload
            url = `${base}/${mode}/${upload.id}`
        }

        const xhr = new XMLHttpRequest()
        xhr.open('POST', url)

        if (upload.uploadToken) {
            xhr.setRequestHeader('X-UploadToken', upload.uploadToken)
        }
        if (basicAuth) {
            xhr.setRequestHeader('Authorization', 'Basic ' + basicAuth)
        }

        // XSRF token for mutating request
        const xsrf = getXsrfToken()
        if (xsrf) xhr.setRequestHeader('X-XSRFToken', xsrf)

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable && onProgress) {
                onProgress(Math.round((e.loaded / e.total) * 100))
            }
        })

        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                try {
                    resolve(JSON.parse(xhr.responseText))
                } catch {
                    resolve(null)
                }
            } else {
                let message = 'Upload failed'
                try {
                    const body = JSON.parse(xhr.responseText)
                    message = body.message || message
                } catch { /* ignore */ }
                reject({ status: xhr.status, message })
            }
        })

        xhr.addEventListener('error', () => {
            reject({ status: 0, message: 'Connection failed' })
        })

        // Send the file as form data
        const formData = new FormData()
        formData.append('file', fileEntry.file, fileEntry.fileName)
        xhr.send(formData)
    })
}

export function removeFile(upload, file) {
    const mode = upload.stream ? 'stream' : 'file'
    const url = `${base}/${mode}/${upload.id}/${file.id}/${file.fileName}`
    const headers = {}
    if (upload.uploadToken) headers['X-UploadToken'] = upload.uploadToken
    return apiCall(url, 'DELETE', null, headers)
}

// ── URL Builders ──

let _downloadDomain = ''

export function setDownloadDomain(domain) {
    _downloadDomain = domain || ''
}

function downloadBase() {
    return _downloadDomain || base
}

export function getFileURL(uploadId, fileId, fileName) {
    return `${downloadBase()}/file/${uploadId}/${fileId}/${fileName}`
}

export function getArchiveURL(uploadId, fileName = 'archive.zip') {
    return `${downloadBase()}/archive/${uploadId}/${fileName}`
}

export function getAdminURL(uploadId, uploadToken) {
    const url = `${window.location.origin}${window.location.pathname}#/?id=${uploadId}`
    return uploadToken ? `${url}&uploadToken=${uploadToken}` : url
}

export function getQrCodeURL(url, size = 200) {
    return `${base}/qrcode?url=${encodeURIComponent(url)}&size=${size}`
}
