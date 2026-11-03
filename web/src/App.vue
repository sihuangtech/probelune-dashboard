<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import zhTw from 'element-plus/es/locale/lang/zh-tw'
import ja from 'element-plus/es/locale/lang/ja'
import ko from 'element-plus/es/locale/lang/ko'
import de from 'element-plus/es/locale/lang/de'
import fr from 'element-plus/es/locale/lang/fr'
import es from 'element-plus/es/locale/lang/es'
import { useRoute } from 'vue-router'
import { api, user, displayTimezone } from './lib/api'
import type { User } from './lib/models'
import { t, language } from './lib/i18n'
import LoginView from './views/LoginView.vue'
import AppShell from './components/AppShell.vue'
import AppTools from './components/AppTools.vue'
const ready = ref(false),
  initialized = ref(true),
  error = ref('')
const route = useRoute()
const elementLocales: Record<string, typeof en> = {
  en,
  'zh-CN': zhCn,
  'zh-TW': zhTw,
  ja,
  ko,
  de,
  fr,
  es,
}
const uiLocale = computed(() => elementLocales[language.value] || en)
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
    <div v-else-if="!ready" class="entry-layout">
      <header class="entry-tools"><AppTools /></header>
      <div class="boot-screen">
        <span class="brand-orbit">◒</span>
        <h1>{{ t('appName') }}</h1>
        <p>{{ error || t('loading') }}</p>
        <el-button v-if="error" @click="load">{{ t('retry') }}</el-button>
      </div>
    </div>
    <LoginView v-else-if="!user" :initialized="initialized" @signed-in="initialized = true" />
    <AppShell v-else />
  </el-config-provider>
</template>
