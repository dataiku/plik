import { createRouter, createWebHashHistory } from 'vue-router'
import { config } from './config.js'
import { auth } from './authStore.js'
import RootView from './views/RootView.vue'
import LoginView from './views/LoginView.vue'
import HomeView from './views/HomeView.vue'
import AdminView from './views/AdminView.vue'
import ClientsView from './views/ClientsView.vue'

const routes = [
    {
        path: '/',
        name: 'root',
        component: RootView,
    },
    {
        path: '/login',
        name: 'login',
        component: LoginView,
    },
    {
        path: '/home',
        name: 'home',
        component: HomeView,
    },
    {
        path: '/admin',
        name: 'admin',
        component: AdminView,
    },
    {
        path: '/clients',
        name: 'clients',
        component: ClientsView,
    },
    {
        path: '/upload/:id',
        redirect: to => {
            // Redirect old URLs to new format
            return { path: '/', query: { id: to.params.id, ...to.query } }
        }
    },
]

const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

// Redirect authenticated users away from the login page
// Redirect to login when authentication is forced and user is not logged in
// Allow download pages (/?id=...) and the login page itself
router.beforeEach((to) => {
    // If already authenticated, skip the login page
    if (to.name === 'login' && auth.user) return { name: 'root' }

    if (config.feature_authentication !== 'forced') return true
    if (auth.user) return true
    if (to.name === 'login') return true
    // Allow CLI client download page without authentication
    if (to.name === 'clients') return true
    // Allow download pages — they have ?id= query
    if (to.name === 'root' && to.query.id) return true
    return { name: 'login' }
})

export default router
