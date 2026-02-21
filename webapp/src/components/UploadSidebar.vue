<script setup>
import { computed, ref } from 'vue'
import { config, isFeatureEnabled, isFeatureForced } from '../config.js'
import { secondsToTTL } from '../utils.js'
import { generatePassphrase } from '../crypto.js'

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  effectiveMaxTTL: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['update:settings'])

function updateSetting(key, value) {
  emit('update:settings', { ...props.settings, [key]: value })
}

// Password generation
const copied = ref(false)
const e2eeCopied = ref(false)
const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*'

function generatePassword() {
  const array = new Uint8Array(20)
  crypto.getRandomValues(array)
  return Array.from(array, b => chars[b % chars.length]).join('')
}

function togglePassword() {
  if (isFeatureForced('password')) return
  const enabling = !props.settings.passwordEnabled
  if (enabling) {
    emit('update:settings', {
      ...props.settings,
      passwordEnabled: true,
      login: 'user',
      password: generatePassword(),
    })
  } else {
    updateSetting('passwordEnabled', false)
  }
}

function copyPassword() {
  if (!props.settings.password) return
  navigator.clipboard.writeText(props.settings.password)
  copied.value = true
  setTimeout(() => { copied.value = false }, 1500)
}

// E2EE toggle
function toggleE2EE() {
  const enabling = !props.settings.e2eeEnabled
  if (enabling) {
    emit('update:settings', {
      ...props.settings,
      e2eeEnabled: true,
      e2eePassphrase: generatePassphrase(),
    })
  } else {
    emit('update:settings', {
      ...props.settings,
      e2eeEnabled: false,
      e2eePassphrase: '',
    })
  }
}

function copyE2EEPassphrase() {
  if (!props.settings.e2eePassphrase) return
  navigator.clipboard.writeText(props.settings.e2eePassphrase)
  e2eeCopied.value = true
  setTimeout(() => { e2eeCopied.value = false }, 1500)
}

// TTL handling
const defaultTTL = computed(() => secondsToTTL(config.defaultTTL))
const ttlValue = computed({
  get: () => props.settings.ttlValue,
  set: (v) => updateSetting('ttlValue', Number(v)),
})
const ttlUnit = computed({
  get: () => props.settings.ttlUnit,
  set: (v) => updateSetting('ttlUnit', v),
})
const maxTTL = computed(() => {
  const val = props.effectiveMaxTTL || config.maxTTL
  return val && val > 0 ? secondsToTTL(val) : null
})

// "Never expires" is allowed when there is no maxTTL limit
const canNeverExpire = computed(() => {
  const val = props.effectiveMaxTTL || config.maxTTL
  return !val || val <= 0
})

function toggleNeverExpires() {
  updateSetting('neverExpires', !props.settings.neverExpires)
}

const hasAnySettings = computed(() =>
  isFeatureEnabled('one_shot') ||
  isFeatureEnabled('stream') ||
  isFeatureEnabled('removable') ||
  isFeatureEnabled('password') ||
  isFeatureEnabled('comments') ||
  isFeatureEnabled('extend_ttl') ||
  isFeatureEnabled('set_ttl') ||
  isFeatureEnabled('e2ee')
)
</script>

<template>
  <aside v-if="hasAnySettings" class="w-full md:w-72 md:shrink-0 p-4 space-y-3 animate-slide-in">
    <!-- Upload Settings -->
    <div class="sidebar-section">
      <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider mb-2">Upload Settings</h3>

      <!-- One Shot -->
      <label v-if="isFeatureEnabled('one_shot')"
             class="flex items-center justify-between py-1 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Destruct after download
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.oneShot"
                :disabled="isFeatureForced('one_shot')"
                @click="!isFeatureForced('one_shot') && updateSetting('oneShot', !settings.oneShot)">
          <span class="toggle-dot" />
        </button>
      </label>

      <!-- Streaming -->
      <label v-if="isFeatureEnabled('stream')"
             class="flex items-center justify-between py-1 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Streaming
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.stream"
                :disabled="isFeatureForced('stream')"
                @click="!isFeatureForced('stream') && updateSetting('stream', !settings.stream)">
          <span class="toggle-dot" />
        </button>
      </label>

      <!-- Removable -->
      <label v-if="isFeatureEnabled('removable')"
             class="flex items-center justify-between py-1 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Removable
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.removable"
                :disabled="isFeatureForced('removable')"
                @click="!isFeatureForced('removable') && updateSetting('removable', !settings.removable)">
          <span class="toggle-dot" />
        </button>
      </label>

      <!-- E2EE -->
      <div v-if="isFeatureEnabled('e2ee')">
        <label class="flex items-center justify-between py-1 cursor-pointer group">
          <span class="text-sm text-surface-200 group-hover:text-white transition-colors flex items-center gap-1.5">
            <svg class="w-3.5 h-3.5 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            End-to-End Encryption
          </span>
          <button type="button"
                  class="toggle-switch"
                  :class="{ 'opacity-50 cursor-not-allowed': isFeatureForced('e2ee') }"
                  :data-active="settings.e2eeEnabled"
                  :disabled="isFeatureForced('e2ee')"
                  @click="toggleE2EE">
            <span class="toggle-dot" />
          </button>
        </label>
        <div v-if="settings.e2eeEnabled" class="mt-2">
          <label class="text-xs text-surface-500 mb-1 block">Passphrase</label>
          <div class="relative">
            <input type="text"
                   class="input-field pr-9 font-mono text-xs"
                   :value="settings.e2eePassphrase"
                   @input="updateSetting('e2eePassphrase', $event.target.value)" />
            <button type="button"
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-surface-400 hover:text-white transition-colors"
                    title="Copy passphrase"
                    @click="copyE2EEPassphrase">
              <svg v-if="!e2eeCopied" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <rect x="9" y="9" width="13" height="13" rx="2" stroke-width="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke-width="2" />
              </svg>
              <svg v-else class="w-4 h-4 text-success-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Password -->
      <div v-if="isFeatureEnabled('password')">
        <label class="flex items-center justify-between py-1 cursor-pointer group">
          <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
            Password
          </span>
          <button type="button"
                  class="toggle-switch"
                  :data-active="settings.passwordEnabled"
                  :disabled="isFeatureForced('password')"
                  @click="togglePassword">
            <span class="toggle-dot" />
          </button>
        </label>
        <div v-if="settings.passwordEnabled" class="mt-2 space-y-2">
          <input type="text"
                 class="input-field"
                 placeholder="Login"
                 :value="settings.login"
                 @input="updateSetting('login', $event.target.value)" />
          <div class="relative">
            <input type="text"
                   class="input-field pr-9 font-mono text-xs"
                   placeholder="Password"
                   :value="settings.password"
                   @input="updateSetting('password', $event.target.value)" />
            <button type="button"
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-surface-400 hover:text-white transition-colors"
                    title="Copy password"
                    @click="copyPassword">
              <svg v-if="!copied" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <rect x="9" y="9" width="13" height="13" rx="2" stroke-width="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke-width="2" />
              </svg>
              <svg v-else class="w-4 h-4 text-success-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Comment -->
      <label v-if="isFeatureEnabled('comments')"
             class="flex items-center justify-between py-1 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Comment
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.commentEnabled"
                :disabled="isFeatureForced('comments')"
                @click="!isFeatureForced('comments') && updateSetting('commentEnabled', !settings.commentEnabled)">
          <span class="toggle-dot" />
        </button>
      </label>

      <!-- Extend TTL -->
      <label v-if="isFeatureEnabled('extend_ttl')"
             class="flex items-center justify-between py-1 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Extend TTL on access
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.extendTTL"
                :disabled="isFeatureForced('extend_ttl')"
                @click="!isFeatureForced('extend_ttl') && updateSetting('extendTTL', !settings.extendTTL)">
          <span class="toggle-dot" />
        </button>
      </label>
    </div>

    <!-- TTL Section -->
    <div v-if="isFeatureEnabled('set_ttl')" class="sidebar-section">
      <h3 class="text-xs font-semibold text-surface-400 uppercase tracking-wider mb-2">Expiration</h3>

      <!-- Never expires toggle -->
      <label v-if="canNeverExpire"
             class="flex items-center justify-between py-1 mb-2 cursor-pointer group">
        <span class="text-sm text-surface-200 group-hover:text-white transition-colors">
          Never expires
        </span>
        <button type="button"
                class="toggle-switch"
                :data-active="settings.neverExpires"
                @click="toggleNeverExpires">
          <span class="toggle-dot" />
        </button>
      </label>

      <div v-if="!settings.neverExpires" class="flex items-center gap-2">
        <input type="number"
               class="input-field w-20"
               min="1"
               :value="ttlValue"
               :disabled="isFeatureForced('set_ttl')"
               @input="ttlValue = $event.target.value" />
        <select class="input-field flex-1"
                :value="ttlUnit"
                :disabled="isFeatureForced('set_ttl')"
                @change="ttlUnit = $event.target.value">
          <option value="minutes">minutes</option>
          <option value="hours">hours</option>
          <option value="days">days</option>
        </select>
      </div>
      <p v-if="maxTTL && !settings.neverExpires" class="text-xs text-surface-500 mt-1">
        Max: {{ maxTTL.value }} {{ maxTTL.unit }}
      </p>
    </div>
  </aside>
</template>
