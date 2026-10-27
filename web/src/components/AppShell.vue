<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Connection, Aim, Bell, Setting, Grid, Warning } from '@element-plus/icons-vue'
import { api, user } from '../lib/api'
import { t } from '../lib/i18n'
const route = useRoute()
const navigation = [
  { path: '/', key: 'overview', icon: Grid },
  { path: '/monitors', key: 'monitors', icon: Monitor },
  { path: '/nodes', key: 'nodes', icon: Connection },
  { path: '/diagnostics', key: 'diagnostics', icon: Aim },
  { path: '/incidents', key: 'incidents', icon: Warning },
  { path: '/alerts', key: 'alerts', icon: Bell },
  { path: '/settings', key: 'settings', icon: Setting },
]
const title = computed(
  () =>
    navigation.find(
      (n) => n.path === route.path || (n.path !== '/' && route.path.startsWith(n.path)),
    )?.key || 'overview',
)
async function logout() {
  await api('/auth/logout', 'POST')
  user.value = null
}
</script>
<template>
  <div class="workspace">
    <aside class="sidebar">
      <router-link to="/" class="brand"
        ><span class="brand-orbit">◒</span><span>Probelune<small>巡 汐</small></span></router-link
      >
      <p class="eyebrow nav-caption">{{ t('monitorNetwork') }}</p>
      <nav aria-label="Main navigation">
        <router-link
          v-for="item in navigation"
          :key="item.path"
          :to="item.path"
          :class="{
            selected:
              route.path === item.path || (item.path !== '/' && route.path.startsWith(item.path)),
          }"
          ><el-icon><component :is="item.icon" /></el-icon
          ><span>{{ t(item.key) }}</span></router-link
        >
      </nav>
      <div class="sidebar-bottom">
        <span class="status-dot" /> SELF-HOSTED <small>v0.1.0</small>
      </div>
    </aside>
    <div class="content-shell">
      <header class="topbar">
        <div>
          <span class="eyebrow">PROBELUNE / </span><strong>{{ t(title) }}</strong>
        </div>
        <div class="account">
          <span>{{ user?.username }}</span
          ><span class="role-label">{{ t(user?.role || 'viewer') }}</span
          ><button @click="logout">{{ t('logout') }}</button>
        </div>
      </header>
      <main class="page-content"><router-view /></main>
      <footer>Probelune · {{ t('tagline') }}</footer>
    </div>
  </div>
</template>
