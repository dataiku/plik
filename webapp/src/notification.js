import { ref } from 'vue'

// Notification types: 'error' | 'success'
export const notification = ref(null)

let dismissTimer = null

/**
 * Show an error notification banner.
 * @param {string} message - Human-readable error message
 */
export function showError(message) {
    show('error', message)
}

/**
 * Show a success notification banner.
 * @param {string} message - Human-readable success message
 */
export function showSuccess(message) {
    show('success', message)
}

/**
 * Dismiss the current notification.
 */
export function dismiss() {
    notification.value = null
    if (dismissTimer) {
        clearTimeout(dismissTimer)
        dismissTimer = null
    }
}

function show(type, message) {
    dismiss()
    notification.value = { type, message }
    dismissTimer = setTimeout(dismiss, 5000)
}
