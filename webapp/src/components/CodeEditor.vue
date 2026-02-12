<script setup>
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue'
import { EditorState } from '@codemirror/state'
import { EditorView, keymap, placeholder as cmPlaceholder, lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view'
import { defaultKeymap, indentWithTab, history, historyKeymap } from '@codemirror/commands'
import { syntaxHighlighting, bracketMatching, foldGutter, indentOnInput } from '@codemirror/language'
import { oneDark } from '@codemirror/theme-one-dark'
import { languages } from '@codemirror/language-data'

const props = defineProps({
  modelValue: { type: String, default: '' },
  filename: { type: String, default: '' },
  readonly: { type: Boolean, default: false },
  placeholder: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue', 'language-detected'])

const editorContainer = ref(null)
let view = null

// Map file extension to language
function getLanguageFromFilename(filename) {
  if (!filename) return null
  const ext = filename.split('.').pop()?.toLowerCase()
  if (!ext) return null

  // Find matching language from CodeMirror's language-data
  for (const lang of languages) {
    if (lang.extensions && lang.extensions.includes(ext)) {
      return lang
    }
    // Also check alias-based matching
    if (lang.alias && lang.alias.includes(ext)) {
      return lang
    }
  }
  return null
}

const detectedLanguage = computed(() => {
  const lang = getLanguageFromFilename(props.filename)
  if (lang) return lang.name
  if (props.filename) {
    const ext = props.filename.split('.').pop()?.toLowerCase()
    if (ext) return ext.toUpperCase()
  }
  return 'Plain Text'
})

// Custom theme to match Plik's dark glass-card aesthetic
const plikTheme = EditorView.theme({
  '&': {
    fontSize: '13px',
    backgroundColor: 'transparent',
  },
  '.cm-content': {
    fontFamily: 'ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Consolas, "DejaVu Sans Mono", monospace',
    caretColor: '#38bdf8',
    minHeight: '120px',
  },
  '&.cm-focused .cm-content': {
    caretColor: '#38bdf8',
  },
  '.cm-gutters': {
    backgroundColor: 'color-mix(in srgb, #1e293b 40%, transparent)',
    color: '#475569',
    border: 'none',
    borderRight: '1px solid color-mix(in srgb, #334155 50%, transparent)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'color-mix(in srgb, #334155 40%, transparent)',
    color: '#94a3b8',
  },
  '.cm-activeLine': {
    backgroundColor: 'color-mix(in srgb, #334155 25%, transparent)',
  },
  '&.cm-focused .cm-cursor': {
    borderLeftColor: '#38bdf8',
  },
  '&.cm-focused .cm-selectionBackground, ::selection': {
    backgroundColor: 'color-mix(in srgb, #0ea5e9 20%, transparent)',
    color: 'inherit',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'color-mix(in srgb, #0ea5e9 15%, transparent)',
  },
  '.cm-foldGutter': {
    color: '#475569',
  },
  '.cm-tooltip': {
    backgroundColor: '#1e293b',
    border: '1px solid #334155',
    color: '#e2e8f0',
  },
  '.cm-placeholder': {
    color: '#64748b',
    fontStyle: 'italic',
  },
  '.cm-scroller': {
    overflow: 'auto',
  },
})

async function createEditor() {
  if (!editorContainer.value) return

  // Build extensions
  const extensions = [
    lineNumbers(),
    highlightActiveLineGutter(),
    highlightActiveLine(),
    history(),
    bracketMatching(),
    foldGutter(),
    indentOnInput(),
    oneDark,
    plikTheme,
    keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
    EditorView.lineWrapping,
  ]

  if (props.placeholder) {
    extensions.push(cmPlaceholder(props.placeholder))
  }

  if (props.readonly) {
    extensions.push(EditorState.readOnly.of(true))
    extensions.push(EditorView.editable.of(false))
  } else {
    // Listen for changes and emit v-model updates
    extensions.push(EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        emit('update:modelValue', update.state.doc.toString())
      }
    }))
  }

  // Load language support if detected
  const langDesc = getLanguageFromFilename(props.filename)
  if (langDesc) {
    try {
      const langSupport = await langDesc.load()
      extensions.push(langSupport)
    } catch (e) {
      console.error('Failed to load language support:', e)
    }
  }

  const state = EditorState.create({
    doc: props.modelValue || '',
    extensions,
  })

  view = new EditorView({
    state,
    parent: editorContainer.value,
  })
}

function destroyEditor() {
  if (view) {
    view.destroy()
    view = null
  }
}

// Recreate editor when filename changes (language switch)
watch(() => props.filename, async () => {
  const currentContent = view ? view.state.doc.toString() : props.modelValue
  destroyEditor()
  // Use an intermediate to preserve content during rebuild
  const savedContent = currentContent
  emit('update:modelValue', savedContent)
  await createEditor()
})



let detectionTimeout = null

async function detectLanguage(content) {
  if (!content || content.length < 10) return

  try {
    const hljs = (await import('highlight.js')).default
    const result = hljs.highlightAuto(content)
    const detected = result.language

    if (detected && result.relevance > 5) { 
      const lowerDetected = detected.toLowerCase()
      const candidates = languages.filter(l => 
        (l.name.toLowerCase() === lowerDetected) || 
        (l.alias && l.alias.includes(lowerDetected))
      )
      
      // Prefer exact name match to avoid aliases (e.g. Starlark aliasing python)
      let cmLang = candidates.find(l => l.name.toLowerCase() === lowerDetected)
      if (!cmLang && candidates.length > 0) cmLang = candidates[0]


      if (cmLang && cmLang.extensions && cmLang.extensions.length > 0) {
        emit('language-detected', {
          language: cmLang.name,
          extension: cmLang.extensions[0]
        })
      }
    }
  } catch (e) {
    console.warn('Language detection failed:', e)
  }
}

// Sync external changes to modelValue (e.g. clipboard paste before editor is ready)
watch(() => props.modelValue, (newVal) => {
  if (view && newVal !== view.state.doc.toString()) {
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: newVal || '' }
    })
  }

  // Debounce detection
  if (detectionTimeout) clearTimeout(detectionTimeout)
  detectionTimeout = setTimeout(() => {
    detectLanguage(newVal)
  }, 1000)
})


onMounted(() => {
  createEditor()
})

onBeforeUnmount(() => {
  destroyEditor()
})
</script>

<template>
  <div class="code-editor-wrapper">
    <!-- Language badge -->
    <div class="flex items-center justify-between px-3 py-1.5 border-b border-surface-700/50">
      <div class="flex items-center gap-2">
        <svg class="w-3.5 h-3.5 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
        <span class="text-xs text-surface-400 font-medium">{{ detectedLanguage }}</span>
      </div>
      <div v-if="readonly" class="flex items-center gap-1">
        <svg class="w-3 h-3 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        <span class="text-xs text-surface-500">Read only</span>
      </div>
    </div>
    <!-- Editor mount point -->
    <div ref="editorContainer" class="code-editor-container" />
  </div>
</template>

<style>
.code-editor-wrapper {
  border-radius: 0.5rem;
  overflow: hidden;
  background-color: color-mix(in srgb, var(--color-surface-900) 80%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-surface-700) 50%, transparent);
}

.code-editor-container .cm-editor {
  max-height: 60vh;
  min-height: 150px;
}

.code-editor-container .cm-editor.cm-focused {
  outline: none;
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--color-accent-500) 30%, transparent);
}

.code-editor-container .cm-scroller {
  overflow: auto;
}
</style>
