import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: () => import('../views/OverviewView.vue') },
    { path: '/monitors', component: () => import('../views/MonitorsView.vue') },
    { path: '/monitors/:id', component: () => import('../views/MonitorDetail.vue') },
    { path: '/nodes', component: () => import('../views/NodesView.vue') },
    { path: '/diagnostics', component: () => import('../views/DiagnosticsView.vue') },
    { path: '/incidents', component: () => import('../views/IncidentsView.vue') },
    { path: '/alerts', component: () => import('../views/AlertsView.vue') },
    { path: '/settings', component: () => import('../views/SettingsView.vue') },
    { path: '/status', component: () => import('../views/StatusView.vue') },
  ],
})

export default router
