import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { notification, showError, showSuccess, dismiss } from '../notification.js'

describe('notification store', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        dismiss() // reset state
    })

    afterEach(() => {
        vi.useRealTimers()
    })

    it('starts with null', () => {
        expect(notification.value).toBeNull()
    })

    it('showError sets an error notification', () => {
        showError('Something broke')
        expect(notification.value).toEqual({ type: 'error', message: 'Something broke' })
    })

    it('showSuccess sets a success notification', () => {
        showSuccess('All good')
        expect(notification.value).toEqual({ type: 'success', message: 'All good' })
    })

    it('dismiss clears the notification', () => {
        showError('oops')
        dismiss()
        expect(notification.value).toBeNull()
    })

    it('auto-dismisses after 5 seconds', () => {
        showError('temporary')
        expect(notification.value).not.toBeNull()
        vi.advanceTimersByTime(4999)
        expect(notification.value).not.toBeNull()
        vi.advanceTimersByTime(1)
        expect(notification.value).toBeNull()
    })

    it('replaces previous notification on new show', () => {
        showError('first')
        showSuccess('second')
        expect(notification.value).toEqual({ type: 'success', message: 'second' })
    })

    it('resets auto-dismiss timer on replacement', () => {
        showError('first')
        vi.advanceTimersByTime(3000) // 3s into first
        showError('second') // resets timer
        vi.advanceTimersByTime(3000) // 3s into second (total 6s from first)
        expect(notification.value).not.toBeNull() // still visible
        vi.advanceTimersByTime(2000) // 5s into second
        expect(notification.value).toBeNull()
    })
})
