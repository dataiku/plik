import { createApp } from 'vue'
import App from './App.vue'
import router from './router.js'
import { loadConfig } from './config.js'
import { checkSession } from './authStore.js'
import './style.css'

const app = createApp(App)
app.use(router)

// Load server config and check auth session before mounting
Promise.all([loadConfig(), checkSession()]).then(() => {
    app.mount('#app')
})
