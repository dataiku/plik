<script setup>
import { notification, dismiss } from '../notification.js'
</script>

<template>
  <Transition name="notification">
    <div v-if="notification"
         class="fixed top-16 left-1/2 -translate-x-1/2 z-100 max-w-lg w-full mx-4"
         role="alert">
      <div class="flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg backdrop-blur-md border"
           :class="notification.type === 'error'
             ? 'bg-red-500/15 border-red-500/30 text-red-300'
             : 'bg-emerald-500/15 border-emerald-500/30 text-emerald-300'">
        <!-- Icon -->
        <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path v-if="notification.type === 'error'"
                stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
          <path v-else
                stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 13l4 4L19 7" />
        </svg>

        <!-- Message -->
        <span class="text-sm flex-1">{{ notification.message }}</span>

        <!-- Dismiss -->
        <button @click="dismiss"
                class="shrink-0 p-1 rounded-lg hover:bg-white/10 transition-colors"
                aria-label="Dismiss notification">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.notification-enter-active {
  transition: all 0.3s ease-out;
}
.notification-leave-active {
  transition: all 0.2s ease-in;
}
.notification-enter-from {
  opacity: 0;
  transform: translate(-50%, -1rem);
}
.notification-leave-to {
  opacity: 0;
  transform: translate(-50%, -0.5rem);
}
</style>
