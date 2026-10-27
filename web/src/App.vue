<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import { useRoute } from 'vue-router'
import { api, user, displayTimezone } from './lib/api'
import type { User } from './lib/models'
import { t, toggleLanguage, language } from './lib/i18n'
import LoginView from './views/LoginView.vue'
import AppShell from './components/AppShell.vue'
const ready = ref(false),
  initialized = ref(true),
  error = ref('')
const route = useRoute()
const uiLocale = computed(() => (language.value === 'zh' ? zhCn : en))
async function load() {
  error.value = ''
  try {
    const status = await api<{ initialized: boolean }>('/auth/status')
    initialized.value = status.initialized
    if (status.initialized) user.value = await api<User>('/auth/me').catch(() => null)
    ready.value = true
  } catch (e) {
    error.value = String(e)
  }
}
onMounted(load)
watch(user, async (current) => {
  if (current) {
    try {
      displayTimezone.value = (await api<{ timezone: string }>('/settings')).timezone
    } catch (e) {
      error.value = String(e)
    }
  }
})
</script>

<template>
  <el-config-provider :locale="uiLocale">
    <router-view v-if="route.path === '/status'" />
    <div v-else-if="!ready" class="boot-screen">
      <span class="brand-orbit">◒</span>
      <h1>Probelune</h1>
      <p>{{ error || t('loading') }}</p>
      <el-button v-if="error" @click="load">{{ t('retry') }}</el-button>
    </div>
    <LoginView v-else-if="!user" :initialized="initialized" @signed-in="initialized = true" />
    <AppShell v-else />
    <button class="language-toggle" @click="toggleLanguage">
      {{ language === 'zh' ? 'EN' : '中文' }}
    </button>
  </el-config-provider>
</template>
