<script setup>
import { computed } from 'vue'
import AppHeader from './components/AppHeader.vue'
import NotificationBanner from './components/NotificationBanner.vue'
import { branding } from './branding.js'
import { config } from './config.js'

const bgStyle = computed(() => {
    const style = {}
    if (branding.backgroundImage) {
        style.backgroundImage = `url(${branding.backgroundImage})`
        style.backgroundSize = 'cover'
        style.backgroundPosition = 'center center'
        style.backgroundAttachment = 'fixed'
        style.backgroundRepeat = 'no-repeat'
    }
    if (branding.backgroundColor) {
        style.backgroundColor = branding.backgroundColor
    }
    return style
})

const overlayStyle = computed(() => ({
    backgroundColor: `rgba(0, 0, 0, ${branding.overlayOpacity ?? 0.55})`,
}))

const hasBackground = computed(() => !!branding.backgroundImage)
</script>

<template>
  <div class="min-h-screen flex flex-col relative" :style="bgStyle">
    <!-- Dark overlay for readability -->
    <div v-if="hasBackground"
         class="fixed inset-0 z-0 pointer-events-none"
         :style="overlayStyle"></div>

    <!-- Header -->
    <AppHeader class="relative z-50" />
    <NotificationBanner />

    <!-- Main Content Area -->
    <div class="flex-1 flex relative z-10">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>

    <!-- Abuse contact footer -->
    <footer v-if="config.abuseContact"
            class="relative z-10 text-center text-xs text-surface-400 py-3">
      For abuse contact {{ config.abuseContact }}
    </footer>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

