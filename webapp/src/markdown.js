import { marked } from 'marked'
import DOMPurify from 'dompurify'

/**
 * Render Markdown text to sanitized HTML.
 *
 * Uses DOMPurify to prevent XSS from user-supplied content
 * (e.g. upload comments rendered via v-html).
 *
 * @param {string} text - Raw Markdown text
 * @returns {string} Sanitized HTML string
 */
export function renderMarkdown(text) {
    if (!text) return ''
    const html = marked.parse(text, { breaks: true })
    return DOMPurify.sanitize(html)
}
