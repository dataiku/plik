<script setup>
import { computed, ref } from 'vue'
import { config, isFeatureEnabled, isFeatureForced } from '../config.js'
import { secondsToTTL } from '../utils.js'

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

const hasAnySettings = computed(() =>
  isFeatureEnabled('one_shot') ||
  isFeatureEnabled('stream') ||
  isFeatureEnabled('removable') ||
  isFeatureEnabled('password') ||
  isFeatureEnabled('comments') ||
  isFeatureEnabled('extend_ttl') ||
  isFeatureEnabled('set_ttl')
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
      <div class="flex items-center gap-2">
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
      <p v-if="maxTTL" class="text-xs text-surface-500 mt-1">
        Max: {{ maxTTL.value }} {{ maxTTL.unit }}
      </p>
    </div>
  </aside>
</template>
